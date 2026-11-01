package kr.pulsenote.pn12usb;

import android.app.*;
import android.content.*;
import android.content.pm.ServiceInfo;
import android.hardware.usb.*;
import android.os.*;
import java.io.IOException;
import java.util.*;
import java.util.concurrent.ConcurrentHashMap;

// One foreground USB host service, up to three independent CDC sessions.
// A device disconnect NEVER stops the other devices.
public final class Pn12UsbService extends Service {
    public static final String START="kr.pulsenote.pn12usb.START",
        STOP="kr.pulsenote.pn12usb.STOP",REFRESH="kr.pulsenote.pn12usb.REFRESH",
        STATUS="kr.pulsenote.pn12usb.STATUS",KEY="kr.pulsenote.pn12usb.KEY",
        DEVICE="device",TEXT="text",NUMBER="number",MODEL="model",SIDE="side";
    public static volatile String status="미연결";
    public static volatile String connectedDeviceName=null;
    public static volatile int connectedKeyCount=15;
    public static volatile String connectedSide="R";
    private static final String CHANNEL="pn12_usb";
    private final Map<String,DeviceSession> sessions=new ConcurrentHashMap<>();
    private final Handler handler=new Handler(Looper.getMainLooper());
    private UsbManager manager;
    private DirectLaunchOverlay overlay;
    private volatile boolean stopping=false;

    private static final class DeviceSession {
        final UsbDevice device;
        volatile UsbDeviceConnection connection;
        volatile Thread thread;
        volatile boolean active=true, verified=false, identityConflict=false;
        volatile int count=0;
        volatile String side="R";
        DeviceSession(UsbDevice d){device=d;}
        void close(){active=false;UsbDeviceConnection c=connection;if(c!=null)c.close();}
    }
    public static boolean candidate(UsbDevice d) {
        boolean control=false,data=false;
        for(int i=0;i<d.getInterfaceCount();i++){
            UsbInterface f=d.getInterface(i);
            if(f.getInterfaceClass()==2&&f.getInterfaceSubclass()==2)control=true;
            if(f.getInterfaceClass()==10&&endpoints(f)!=null)data=true;
        }
        return control&&data;
    }
    private static UsbEndpoint[] endpoints(UsbInterface f){
        UsbEndpoint in=null,out=null;
        for(int i=0;i<f.getEndpointCount();i++){
            UsbEndpoint e=f.getEndpoint(i);
            if(e.getType()!=UsbConstants.USB_ENDPOINT_XFER_BULK)continue;
            if(e.getDirection()==UsbConstants.USB_DIR_IN)in=e;
            if(e.getDirection()==UsbConstants.USB_DIR_OUT)out=e;
        }
        return in==null||out==null?null:new UsbEndpoint[]{in,out};
    }
    public static boolean isConnected(String name){return name!=null && activeNames.contains(name);}
    private static final Set<String> activeNames=ConcurrentHashMap.newKeySet();

    @Override public void onCreate(){
        super.onCreate();
        manager=(UsbManager)getSystemService(USB_SERVICE);
        overlay=new DirectLaunchOverlay(this);
        ((NotificationManager)getSystemService(NOTIFICATION_SERVICE)).createNotificationChannel(
            new NotificationChannel(CHANNEL,"PulseNote USB",NotificationManager.IMPORTANCE_LOW));
    }
    private Notification notification(String message){
        PendingIntent p=PendingIntent.getActivity(this,0,new Intent(this,MainActivity.class),
            PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE);
        return new Notification.Builder(this,CHANNEL).setSmallIcon(android.R.drawable.ic_menu_manage)
            .setContentTitle("PulseNote Deck · 3 USB").setContentText(message).setOngoing(true)
            .setContentIntent(p).build();
    }
    private void broadcastStatus(String message,int model,String side){
        status=message;
        Intent i=new Intent(STATUS).setPackage(getPackageName());
        i.putExtra(TEXT,message);i.putExtra(MODEL,model);i.putExtra(SIDE,side);
        sendBroadcast(i);
        handler.post(()->{
            if(overlay!=null)overlay.refresh(status.startsWith("연결 완료"));
            if(status.startsWith("연결 완료"))
                ((NotificationManager)getSystemService(NOTIFICATION_SERVICE)).notify(44,notification(status));
        });
    }
    private void publishSessions(int model,String side){
        int devices=0,total=0;
        for(DeviceSession s:sessions.values())if(s.verified&&s.active){
            devices++;total+=s.count;
        }
        if(devices>0){
            connectedKeyCount=model;connectedSide=side;
            broadcastStatus("연결 완료 · "+devices+"대 · 총 "+total+"키",model,side);
        }else{
            connectedDeviceName=null;
            broadcastStatus(stopping?"연결 종료":"USB 연결 해제",0,"R");
        }
    }
    @Override public int onStartCommand(Intent intent,int flags,int startId){
        if(intent==null)return START_NOT_STICKY;
        String action=intent.getAction();
        if(REFRESH.equals(action)){
            if(overlay!=null)overlay.refresh(status.startsWith("연결 완료"));
            return START_NOT_STICKY;
        }
        if(STOP.equals(action)){
            stopping=true;
            for(DeviceSession s:sessions.values())s.close();
            sessions.clear();activeNames.clear();
            publishSessions(0,"R");stopSelf();
            return START_NOT_STICKY;
        }
        if(!START.equals(action))return START_NOT_STICKY;
        UsbDevice d=manager.getDeviceList().get(intent.getStringExtra(DEVICE));
        if(d==null||!candidate(d)||!manager.hasPermission(d)){
            if(sessions.isEmpty())broadcastStatus("권한 또는 USB 장치 오류",0,"R");
            return START_NOT_STICKY;
        }
        if(sessions.containsKey(d.getDeviceName()))return START_NOT_STICKY;
        if(sessions.size()>=3){
            broadcastStatus("동시 연결 장치 수는 최대 3대",0,"R");
            return START_NOT_STICKY;
        }
        try{
            if(Build.VERSION.SDK_INT>=29)
                startForeground(44,notification("USB 연결 확인 중"),
                    ServiceInfo.FOREGROUND_SERVICE_TYPE_CONNECTED_DEVICE);
            else startForeground(44,notification("USB 연결 확인 중"));
        }catch(RuntimeException e){
            broadcastStatus("USB 서비스 시작 실패: "+e.getClass().getSimpleName(),0,"R");
            return START_NOT_STICKY;
        }
        stopping=false;
        DeviceSession s=new DeviceSession(d);
        if(sessions.putIfAbsent(d.getDeviceName(),s)!=null)return START_NOT_STICKY;
        s.thread=new Thread(()->readDevice(s),"PND-CDC-"+d.getDeviceId());
        s.thread.start();
        return START_NOT_STICKY;
    }
    private void readDevice(DeviceSession s) {
        UsbDevice d=s.device;
        try{
            UsbInterface ctl=null,dat=null;
            for(int i=0;i<d.getInterfaceCount();i++){
                UsbInterface f=d.getInterface(i);
                if(f.getInterfaceClass()==2&&f.getInterfaceSubclass()==2)ctl=f;
                if(f.getInterfaceClass()==10&&endpoints(f)!=null)dat=f;
            }
            if(ctl==null||dat==null)throw new IOException("CDC 인터페이스 없음");
            UsbEndpoint[] eps=endpoints(dat);
            UsbDeviceConnection c=manager.openDevice(d);
            if(c==null)throw new IOException("USB 장치 열기 실패");
            s.connection=c;
            if(!c.claimInterface(ctl,true)||!c.claimInterface(dat,true))
                throw new IOException("USB 인터페이스 점유 실패");
            byte[] config={0,(byte)0xc2,1,0,0,0,8};
            if(c.controlTransfer(0x21,0x20,0,ctl.getId(),config,config.length,1000)!=7)
                throw new IOException("CDC 속도 설정 실패");
            if(c.controlTransfer(0x21,0x22,3,ctl.getId(),null,0,1000)<0)
                throw new IOException("CDC DTR 실패");
            Pn12Protocol protocol=new Pn12Protocol(new Pn12Protocol.Listener(){
                public void onHandshake(int count,String side){
                    for(DeviceSession other:sessions.values()){
                        if(other!=s&&other.verified&&other.count==count&&other.side.equals(side)){
                            s.identityConflict=true;return;
                        }
                    }
                    s.count=count;s.side=side;s.verified=true;
                    activeNames.add(d.getDeviceName());
                    connectedDeviceName=d.getDeviceName();
                    publishSessions(count,side);
                }
                public void onKey(int key){
                    if(!s.active||!s.verified)return;
                    handler.post(()->{
                        if(!s.active||!s.verified)return;
                        String result;
                        try{result=ActionDispatcher.run(Pn12UsbService.this,key,s.count,s.side);}
                        catch(Exception e){result="동작 오류: "+e.getClass().getSimpleName();}
                        Intent ev=new Intent(KEY).setPackage(getPackageName());
                        ev.putExtra(NUMBER,key);ev.putExtra(TEXT,result);
                        ev.putExtra(MODEL,s.count);ev.putExtra(SIDE,s.side);
                        sendBroadcast(ev);
                    });
                }
            });
            byte[] buffer=new byte[256];long started=SystemClock.elapsedRealtime(),lastHello=0;
            while(s.active&&manager.getDeviceList().containsKey(d.getDeviceName())){
                long now=SystemClock.elapsedRealtime();
                if(!protocol.isVerified()&&now-lastHello>=1500){
                    if(c.bulkTransfer(eps[1],Pn12Protocol.HELLO,Pn12Protocol.HELLO.length,1000)
                        !=Pn12Protocol.HELLO.length)throw new IOException("HELLO 송신 실패");
                    lastHello=now;
                }
                int len=c.bulkTransfer(eps[0],buffer,buffer.length,250);
                if(len>0)protocol.receive(buffer,len);
                if(protocol.isConflicting()||s.identityConflict)
                    throw new IOException("중복 또는 충돌하는 장치 식별 정보");
                if(!protocol.isVerified()&&now-started>12000)
                    throw new IOException("장치 식별 타임아웃");
            }
        }catch(Exception e){
            if(s.active)android.util.Log.w("PulseNoteUSB",d.getDeviceName()+": "+e.getMessage());
        }finally{
            s.close();
            sessions.remove(d.getDeviceName(),s);
            activeNames.remove(d.getDeviceName());
            if(!stopping)publishSessions(connectedKeyCount,connectedSide);
            if(sessions.isEmpty()&&!stopping)stopSelf();
        }
    }
    @Override public void onDestroy(){
        stopping=true;
        for(DeviceSession s:sessions.values())s.close();
        sessions.clear();activeNames.clear();
        if(overlay!=null)overlay.remove();
        super.onDestroy();
    }
    @Override public IBinder onBind(Intent i){return null;}
}
