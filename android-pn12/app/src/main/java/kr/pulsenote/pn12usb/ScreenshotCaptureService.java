package kr.pulsenote.pn12usb;

import android.app.*;
import android.content.*;
import android.content.pm.ServiceInfo;
import android.graphics.*;
import android.hardware.display.*;
import android.media.Image;
import android.media.ImageReader;
import android.media.projection.*;
import android.net.Uri;
import android.os.*;
import android.provider.MediaStore;
import android.util.DisplayMetrics;
import android.view.WindowManager;
import android.widget.Toast;
import java.io.*;
import java.nio.ByteBuffer;
import java.text.SimpleDateFormat;
import java.util.*;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * One MediaProjection consent starts ONE long-lived VirtualDisplay. Subsequent
 * screenshot key presses save the latest frame without using the consent token
 * again. The user can stop the session from the persistent notification or UI.
 * The OS can revoke screen sharing at any time (lock, quick settings, privacy).
 *
 * No screenshot is uploaded or saved until a button press; one initial picture
 * is saved after the user accepts the first consent.
 */
public final class ScreenshotCaptureService extends Service {
    public static final String START="kr.pulsenote.pn12usb.CAPTURE_START";
    public static final String CAPTURE="kr.pulsenote.pn12usb.CAPTURE_FRAME";
    public static final String STOP="kr.pulsenote.pn12usb.CAPTURE_STOP";
    public static final String STATE="kr.pulsenote.pn12usb.CAPTURE_STATE";
    public static final String RESULT_CODE="consent_result_code";
    public static final String RESULT_DATA="consent_result_data";

    private static final String CHANNEL="pnd_capture_active";
    private static final int NOTIFICATION_ID=48;
    private static volatile boolean ready=false;
    private static volatile boolean initializing=false;
    public static boolean isSessionReady(){return ready;}
    public static boolean isSessionStarting(){return initializing;}

    private final AtomicBoolean stopping=new AtomicBoolean(false);
    private final Handler main=new Handler(Looper.getMainLooper());
    private HandlerThread thread;
    private Handler worker;
    private MediaProjection projection;
    private VirtualDisplay display;
    private ImageReader reader;
    // Only the worker thread owns this frame. Holding at most one Image, with
    // ImageReader maxImages=3, leaves enough buffers for acquireLatestImage().
    private Image latestFrame;
    private int width,height,densityDpi;
    private int pendingCaptures;

    @Override public void onCreate(){
        super.onCreate();
        NotificationManager nm=(NotificationManager)getSystemService(NOTIFICATION_SERVICE);
        if(nm!=null)nm.createNotificationChannel(new NotificationChannel(CHANNEL,
            "PulseNote 캡처 세션",NotificationManager.IMPORTANCE_LOW));
    }

    private Notification notification(){
        Intent stop=new Intent(this,ScreenshotCaptureService.class).setAction(STOP);
        PendingIntent stopClick=PendingIntent.getService(this,71,stop,
            PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE);
        Intent open=new Intent(this,MainActivity.class).addFlags(
            Intent.FLAG_ACTIVITY_SINGLE_TOP|Intent.FLAG_ACTIVITY_CLEAR_TOP);
        PendingIntent openClick=PendingIntent.getActivity(this,72,open,
            PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE);
        return new Notification.Builder(this,CHANNEL)
            .setSmallIcon(android.R.drawable.ic_menu_camera)
            .setContentTitle("PulseNote 캡처 대기 모드")
            .setContentText("화면 공유 중 · 매크로 버튼으로 저장 · 종료는 여기서")
            .setContentIntent(openClick)
            .addAction(android.R.drawable.ic_menu_close_clear_cancel,"캡처 모드 종료",stopClick)
            .setOngoing(true).build();
    }

    private void broadcastState(){
        Intent changed=new Intent(STATE).setPackage(getPackageName());
        changed.putExtra("ready",ready);
        sendBroadcast(changed);
    }

    private void message(String value){
        main.post(()->Toast.makeText(getApplicationContext(),value,
            Toast.LENGTH_SHORT).show());
    }

    @Override public int onStartCommand(Intent intent,int flags,int startId){
        if(intent==null)return START_NOT_STICKY;
        final String action=intent.getAction();
        if(STOP.equals(action)){
            stopSession("캡처 대기 모드 종료");return START_NOT_STICKY;
        }
        if(CAPTURE.equals(action)){
            if(worker!=null && !stopping.get())worker.post(()->queueCapture());
            else if(!initializing)message("캡처 세션이 종료됐습니다. 다시 승인하세요");
            return START_NOT_STICKY;
        }
        if(!START.equals(action))return START_NOT_STICKY;
        // Never use a MediaProjection result token more than once.
        if(thread!=null||ready||initializing||stopping.get())return START_NOT_STICKY;
        final int result=intent.getIntExtra(RESULT_CODE,Activity.RESULT_CANCELED);
        final Intent grant;
        if(Build.VERSION.SDK_INT>=33)
            grant=intent.getParcelableExtra(RESULT_DATA,Intent.class);
        else grant=intent.getParcelableExtra(RESULT_DATA);
        if(result!=Activity.RESULT_OK||grant==null){
            stopSelf(startId);return START_NOT_STICKY;
        }
        initializing=true;
        try{
            if(Build.VERSION.SDK_INT>=29)
                startForeground(NOTIFICATION_ID,notification(),
                    ServiceInfo.FOREGROUND_SERVICE_TYPE_MEDIA_PROJECTION);
            else startForeground(NOTIFICATION_ID,notification());
            thread=new HandlerThread("PND-Persistent-Capture");
            thread.start();
            worker=new Handler(thread.getLooper());
            // Return to the app being captured before saving the first frame.
            worker.postDelayed(()->begin(result,grant),700);
        }catch(RuntimeException e){
            stopSession("캡처 세션 시작 실패: "+e.getClass().getSimpleName());
        }
        return START_NOT_STICKY;
    }

    private void begin(int result,Intent grant){
        if(stopping.get())return;
        try{
            WindowManager wm=(WindowManager)getSystemService(WINDOW_SERVICE);
            DisplayMetrics metrics=new DisplayMetrics();
            wm.getDefaultDisplay().getRealMetrics(metrics);
            width=metrics.widthPixels;height=metrics.heightPixels;
            densityDpi=metrics.densityDpi;
            validateDimensions(width,height);
            MediaProjectionManager mgr=(MediaProjectionManager)
                getSystemService(MEDIA_PROJECTION_SERVICE);
            if(mgr==null)throw new IOException("MediaProjection 서비스 없음");
            // Android 14+ allows exactly one getMediaProjection() per consent
            // Intent and one createVirtualDisplay() per MediaProjection.
            projection=mgr.getMediaProjection(result,grant);
            if(projection==null)throw new IOException("화면 공유 세션을 만들 수 없습니다");
            projection.registerCallback(new MediaProjection.Callback(){
                @Override public void onStop(){
                    stopSession("Android에서 화면 공유가 종료됐습니다");
                }
                @Override public void onCapturedContentResize(int w,int h){
                    if(worker!=null)worker.post(()->resizeDisplay(w,h));
                }
            },worker);
            reader=createReader(width,height);
            display=projection.createVirtualDisplay("PulseNoteCapture",
                width,height,densityDpi,
                DisplayManager.VIRTUAL_DISPLAY_FLAG_AUTO_MIRROR,
                reader.getSurface(),null,worker);
            if(display==null)throw new IOException("화면 캡처 디스플레이를 만들 수 없습니다");
            pendingCaptures=1;
            ready=true;
            initializing=false;
            broadcastState();
            // If the source is initially static, ensure the first frame
            // either arrives or a clear error is shown instead of hanging.
            worker.postDelayed(()->{
                if(pendingCaptures>0 && latestFrame==null && !stopping.get()){
                    pendingCaptures=0;
                    message("화면 프레임을 받지 못했습니다. 캡처 버튼을 다시 누르세요");
                }
            },6500);
        }catch(Exception e){
            stopSession("캡처 세션 오류: "+e.getClass().getSimpleName());
        }
    }

    private static void validateDimensions(int w,int h)throws IOException{
        if(w<1||h<1||((long)w*h)>24000000L)
            throw new IOException("지원되지 않는 화면 크기");
    }

    private ImageReader createReader(int w,int h){
        ImageReader imageReader=ImageReader.newInstance(w,h,PixelFormat.RGBA_8888,3);
        imageReader.setOnImageAvailableListener(source->{
            Image image=null;
            try{
                image=source.acquireLatestImage();
                if(image==null)return;
                if(stopping.get()){image.close();return;}
                Image previous=latestFrame;
                latestFrame=image;
                image=null;
                if(previous!=null)previous.close();
                savePending();
            }catch(RuntimeException e){
                if(!stopping.get())message("화면 프레임 읽기 오류");
            }finally{
                if(image!=null)image.close();
            }
        },worker);
        return imageReader;
    }

    private void resizeDisplay(int w,int h){
        if(stopping.get()||display==null||(w==width&&h==height))return;
        try{
            validateDimensions(w,h);
            ImageReader next=createReader(w,h);
            display.resize(w,h,densityDpi);
            display.setSurface(next.getSurface());
            if(latestFrame!=null){latestFrame.close();latestFrame=null;}
            if(reader!=null)reader.close();
            reader=next;
            width=w;height=h;
        }catch(Exception e){
            stopSession("화면 크기 변경 실패 · 다시 승인해야 합니다");
        }
    }

    private void queueCapture(){
        if(stopping.get())return;
        // No additional screen-sharing consent while this display is active.
        pendingCaptures=Math.min(4,pendingCaptures+1);
        savePending();
        if(latestFrame==null){
            worker.postDelayed(()->{
                if(!stopping.get()&&latestFrame==null&&pendingCaptures>0){
                    pendingCaptures=0;
                    message("화면을 읽지 못했습니다. 화면을 켜고 다시 시도하세요");
                }
            },6500);
        }
    }

    private void savePending(){
        if(stopping.get()||latestFrame==null||pendingCaptures<=0)return;
        // Serial worker execution ensures no two PNG encoders access the same
        // Image while it is being replaced or closed by the frame listener.
        while(pendingCaptures>0 && !stopping.get()){
            pendingCaptures--;
            try{
                Image.Plane plane=latestFrame.getPlanes()[0];
                int pixelStride=plane.getPixelStride();
                int rowStride=plane.getRowStride();
                if(pixelStride!=4||rowStride<width*4||rowStride%4!=0)
                    throw new IOException("지원하지 않는 화면 픽셀 형식");
                int paddedWidth=rowStride/4;
                if(paddedWidth<width)throw new IOException("화면 너비 불일치");
                ByteBuffer buffer=plane.getBuffer().duplicate();
                buffer.rewind();
                Bitmap padded=Bitmap.createBitmap(paddedWidth,height,Bitmap.Config.ARGB_8888);
                try{
                    padded.copyPixelsFromBuffer(buffer);
                    Bitmap cropped=Bitmap.createBitmap(padded,0,0,width,height);
                    try{save(cropped);}finally{cropped.recycle();}
                }finally{padded.recycle();}
                message("스크린샷 저장 · 사진/PulseNote Deck");
            }catch(Exception e){
                message("캡처 저장 실패: "+e.getClass().getSimpleName());
                break;
            }
        }
    }

    private void save(Bitmap image) throws IOException{
        String name="PulseNote_"+new SimpleDateFormat("yyyyMMdd_HHmmss_SSS",Locale.US)
            .format(new Date())+".png";
        if(Build.VERSION.SDK_INT>=29){
            ContentValues values=new ContentValues();
            values.put(MediaStore.Images.Media.DISPLAY_NAME,name);
            values.put(MediaStore.Images.Media.MIME_TYPE,"image/png");
            values.put(MediaStore.Images.Media.RELATIVE_PATH,
                Environment.DIRECTORY_PICTURES+"/PulseNote Deck");
            values.put(MediaStore.Images.Media.IS_PENDING,1);
            ContentResolver resolver=getContentResolver();
            Uri uri=resolver.insert(MediaStore.Images.Media.EXTERNAL_CONTENT_URI,values);
            if(uri==null)throw new IOException("이미지 저장 위치를 만들 수 없습니다");
            boolean saved=false;
            try{
                try(OutputStream out=resolver.openOutputStream(uri)){
                    if(out==null||!image.compress(Bitmap.CompressFormat.PNG,100,out))
                        throw new IOException("PNG 인코딩에 실패했습니다");
                }
                ContentValues complete=new ContentValues();
                complete.put(MediaStore.Images.Media.IS_PENDING,0);
                resolver.update(uri,complete,null,null);
                saved=true;
            }finally{if(!saved)resolver.delete(uri,null,null);}
        }else{
            // Android 8-9 uses private external Pictures without storage permission.
            File dir=getExternalFilesDir(Environment.DIRECTORY_PICTURES);
            if(dir==null)throw new IOException("캡처 저장소를 사용할 수 없습니다");
            if(!dir.exists()&&!dir.mkdirs())throw new IOException("캡처 폴더를 만들 수 없습니다");
            try(OutputStream out=new FileOutputStream(new File(dir,name))){
                if(!image.compress(Bitmap.CompressFormat.PNG,100,out))
                    throw new IOException("PNG 인코딩에 실패했습니다");
            }
        }
    }


    private void stopSession(String reason){
        if(!stopping.compareAndSet(false,true))return;
        ready=false;
        initializing=false;
        broadcastState();
        if(reason!=null)message(reason);
        Handler target=worker;
        if(target==null){
            stopForeground(STOP_FOREGROUND_REMOVE);
            stopSelf();return;
        }
        target.post(()->{
            pendingCaptures=0;
            if(latestFrame!=null){latestFrame.close();latestFrame=null;}
            if(display!=null){display.release();display=null;}
            if(reader!=null){reader.close();reader=null;}
            if(projection!=null){projection.stop();projection=null;}
            main.post(()->{
                stopForeground(STOP_FOREGROUND_REMOVE);
                stopSelf();
            });
            thread.quitSafely();
        });
    }

    @Override public void onDestroy(){
        if(!stopping.get())stopSession(null);
        super.onDestroy();
    }
    @Override public IBinder onBind(Intent intent){return null;}
}
