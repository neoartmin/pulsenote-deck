package kr.pulsenote.pn12usb;
import android.app.*;
import android.content.*;
import android.content.pm.ServiceInfo;
import android.hardware.usb.*;
import android.os.*;
import java.io.IOException;
import java.util.*;
public final class Pn12UsbService extends Service {
 public static final String START="kr.pulsenote.pn12usb.START",STOP="kr.pulsenote.pn12usb.STOP",REFRESH="kr.pulsenote.pn12usb.REFRESH",STATUS="kr.pulsenote.pn12usb.STATUS",KEY="kr.pulsenote.pn12usb.KEY",DEVICE="device",TEXT="text",NUMBER="number";
 public static volatile String status="미연결";
 private static final String CHANNEL="pn12_usb";
 private UsbManager manager; private DirectLaunchOverlay overlay; private volatile UsbDeviceConnection active; private volatile int generation=0;
 private final Handler handler=new Handler(Looper.getMainLooper());
 public static boolean candidate(UsbDevice d) {
   boolean control=false,data=false;
   for(int i=0;i<d.getInterfaceCount();i++) {
     UsbInterface f=d.getInterface(i);
     if(f.getInterfaceClass()==2 && f.getInterfaceSubclass()==2) control=true;
     if(f.getInterfaceClass()==10 && endpoints(f)!=null) data=true;
   }
   return control && data;
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
 @Override public void onCreate(){
   super.onCreate(); manager=(UsbManager)getSystemService(USB_SERVICE); overlay=new DirectLaunchOverlay(this);
   ((NotificationManager)getSystemService(NOTIFICATION_SERVICE)).createNotificationChannel(new NotificationChannel(CHANNEL,"PN12 USB",NotificationManager.IMPORTANCE_LOW));
 }
 private Notification notification(String msg){
   PendingIntent p=PendingIntent.getActivity(this,0,new Intent(this,MainActivity.class),PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE);
   return new Notification.Builder(this,CHANNEL).setSmallIcon(android.R.drawable.ic_menu_manage).setContentTitle("PulseNote PN12").setContentText(msg).setOngoing(true).setContentIntent(p).build();
 }
 private void report(String s){
   status=s;
   Intent i=new Intent(STATUS).setPackage(getPackageName());i.putExtra(TEXT,s);sendBroadcast(i);
   if(s.startsWith("연결 완료"))((NotificationManager)getSystemService(NOTIFICATION_SERVICE)).notify(44,notification(s));
   handler.post(()->{if(overlay!=null)overlay.refresh(status.startsWith("연결 완료"));});
 }
 @Override public int onStartCommand(Intent i,int flags,int startId){
   if(i==null)return START_NOT_STICKY;
   if(REFRESH.equals(i.getAction())){if(overlay!=null)overlay.refresh(status.startsWith("연결 완료"));return START_NOT_STICKY;}
   if(STOP.equals(i.getAction())){report("연결 종료");stopSelf();return START_NOT_STICKY;}
   if(!START.equals(i.getAction()))return START_NOT_STICKY;
   UsbDevice d=manager.getDeviceList().get(i.getStringExtra(DEVICE));
   if(d==null||!candidate(d)||!manager.hasPermission(d)){report("권한 또는 USB 장치 오류");stopSelf();return START_NOT_STICKY;}
   try{if(Build.VERSION.SDK_INT>=29)startForeground(44,notification("연결 중"),ServiceInfo.FOREGROUND_SERVICE_TYPE_CONNECTED_DEVICE);else startForeground(44,notification("연결 중"));}
   catch(RuntimeException e){report("서비스 실행 실패");stopSelf();return START_NOT_STICKY;}
   final int my=++generation;
   UsbDeviceConnection previous=active;if(previous!=null)previous.close();
   report("연결 시도 중");
   new Thread(()->loop(d,my),"PN12-Reader").start();
   return START_NOT_STICKY;
 }
 private void loop(UsbDevice d,int my) {
   UsbDeviceConnection con=null;
   try{
     UsbInterface ctl=null,dat=null;
     for(int n=0;n<d.getInterfaceCount();n++){
       UsbInterface f=d.getInterface(n);
       if(f.getInterfaceClass()==2&&f.getInterfaceSubclass()==2)ctl=f;
       if(f.getInterfaceClass()==10&&endpoints(f)!=null)dat=f;
     }
     if(ctl==null||dat==null)throw new IOException("CDC 인터페이스 없음");
     UsbEndpoint[] endpoints=endpoints(dat);
     con=manager.openDevice(d);
     if(con==null)throw new IOException("USB 열기 실패");
     if(!con.claimInterface(ctl,true)||!con.claimInterface(dat,true))throw new IOException("USB 인터페이스 점유 실패");
     active=con;
     byte[] config={0,(byte)0xc2,1,0,0,0,8};
     int r=con.controlTransfer(0x21,0x20,0,ctl.getId(),config,config.length,1000);
     if(r!=7)throw new IOException("CDC 속도 설정 실패");
     if(con.controlTransfer(0x21,0x22,3,ctl.getId(),null,0,1000)<0)throw new IOException("CDC DTR 실패");
     Pn12Protocol protocol=new Pn12Protocol(new Pn12Protocol.Listener(){
       public void onHandshake(){report("연결 완료 · PN12 12키");}
       public void onKey(int key){handler.post(()->{
         if(my!=generation)return;
         String result;
         try{result=ActionDispatcher.run(Pn12UsbService.this,key);}catch(Exception e){result="동작 오류: "+e.getClass().getSimpleName();}
         Intent ev=new Intent(KEY).setPackage(getPackageName());ev.putExtra(NUMBER,key);ev.putExtra(TEXT,result);sendBroadcast(ev);
       });}
     });
     byte[] buffer=new byte[256];long start=SystemClock.elapsedRealtime(),last=0;
     while(my==generation&&manager.getDeviceList().containsKey(d.getDeviceName())){
       long now=SystemClock.elapsedRealtime();
       if(!protocol.isVerified()&&now-last>=1500){
         if(con.bulkTransfer(endpoints[1],Pn12Protocol.HELLO,Pn12Protocol.HELLO.length,1000)!=Pn12Protocol.HELLO.length)throw new IOException("HELLO 송신 실패");
         last=now;
       }
       int len=con.bulkTransfer(endpoints[0],buffer,buffer.length,250);
       if(len>0)protocol.receive(buffer,len);
       if(!protocol.isVerified()&&now-start>12000)throw new IOException("PN12 식별 타임아웃");
     }
     if(my==generation)report("USB 연결 해제");
   }catch(Exception e){if(my==generation)report("연결 오류: "+e.getMessage());}
   finally{if(con!=null)con.close();if(active==con)active=null;if(my==generation)stopSelf();}
 }
 @Override public void onDestroy(){generation++;if(overlay!=null)overlay.remove();UsbDeviceConnection con=active;active=null;if(con!=null)con.close();super.onDestroy();}
 @Override public IBinder onBind(Intent intent){return null;}
}
