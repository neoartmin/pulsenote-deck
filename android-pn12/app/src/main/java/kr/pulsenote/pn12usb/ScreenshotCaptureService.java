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
import java.io.*;
import java.nio.ByteBuffer;
import java.text.SimpleDateFormat;
import java.util.*;
import java.util.concurrent.atomic.AtomicBoolean;

/**
 * One consent token -> one VirtualDisplay -> one PNG saved to Pictures.
 * No content is uploaded. The media projection is stopped and released after
 * the first frame; every future capture needs new Android system consent.
 */
public final class ScreenshotCaptureService extends Service {
    public static final String START="kr.pulsenote.pn12usb.CAPTURE_START";
    public static final String RESULT_CODE="consent_result_code";
    public static final String RESULT_DATA="consent_result_data";
    private static final String CHANNEL="pnd_capture_active";
    private final AtomicBoolean closed=new AtomicBoolean(false);
    private final Handler main=new Handler(Looper.getMainLooper());
    private HandlerThread thread;
    private Handler worker;
    private MediaProjection projection;
    private VirtualDisplay display;
    private ImageReader reader;
    private int width,height;

    @Override public void onCreate(){
        super.onCreate();
        ((NotificationManager)getSystemService(NOTIFICATION_SERVICE))
            .createNotificationChannel(new NotificationChannel(CHANNEL,
                "PulseNote 캡처 중",NotificationManager.IMPORTANCE_LOW));
    }

    private Notification notification(){
        return new Notification.Builder(this,CHANNEL)
            .setSmallIcon(android.R.drawable.ic_menu_camera)
            .setContentTitle("PulseNote 화면 저장 중")
            .setContentText("사용자가 승인한 한 화면만 PNG로 저장합니다")
            .setOngoing(true).build();
    }

    @Override public int onStartCommand(Intent intent,int flags,int startId){
        if(intent==null||!START.equals(intent.getAction())||thread!=null){
            stopSelf(startId);return START_NOT_STICKY;
        }
        final int result=intent.getIntExtra(RESULT_CODE,Activity.RESULT_CANCELED);
        final Intent grant;
        if(Build.VERSION.SDK_INT>=33){
            grant=intent.getParcelableExtra(RESULT_DATA,Intent.class);
        }else grant=intent.getParcelableExtra(RESULT_DATA);
        if(result!=Activity.RESULT_OK||grant==null){
            stopSelf(startId);return START_NOT_STICKY;
        }
        try{
            if(Build.VERSION.SDK_INT>=29)
                startForeground(48,notification(),ServiceInfo.FOREGROUND_SERVICE_TYPE_MEDIA_PROJECTION);
            else startForeground(48,notification());
            thread=new HandlerThread("PND-Capture");
            thread.start();
            worker=new Handler(thread.getLooper());
            // The capture consent dialog itself has just closed; allow the
            // previous app to reappear after MainActivity.moveTaskToBack().
            main.postDelayed(()->{
                if(!closed.get())worker.post(()->begin(result,grant));
            },650);
            main.postDelayed(()->finish("화면 캡처 시간이 초과됐습니다"),12000);
        }catch(RuntimeException error){
            finish("화면 캡처 서비스를 시작할 수 없습니다");
        }
        return START_NOT_STICKY;
    }

    private void begin(int result,Intent data){
        if(closed.get())return;
        try{
            WindowManager wm=(WindowManager)getSystemService(WINDOW_SERVICE);
            DisplayMetrics metrics=new DisplayMetrics();
            wm.getDefaultDisplay().getRealMetrics(metrics);
            width=metrics.widthPixels;height=metrics.heightPixels;
            if(width<1||height<1||((long)width*height)>24000000)
                throw new IOException("캡처 화면 크기 오류");
            MediaProjectionManager mgr=(MediaProjectionManager)
                getSystemService(MEDIA_PROJECTION_SERVICE);
            projection=mgr.getMediaProjection(result,data);
            if(projection==null)throw new IOException("화면 공유 승인을 받지 못했습니다");
            projection.registerCallback(new MediaProjection.Callback(){
                @Override public void onStop(){finish("화면 공유 세션 종료");}
            },worker);
            reader=ImageReader.newInstance(width,height,PixelFormat.RGBA_8888,2);
            reader.setOnImageAvailableListener(source->{
                Image image=null;
                try{
                    image=source.acquireLatestImage();
                    if(image==null||closed.get())return;
                    Image.Plane plane=image.getPlanes()[0];
                    int pixelStride=plane.getPixelStride(),rowStride=plane.getRowStride();
                    if(pixelStride!=4||rowStride<width*4)
                        throw new IOException("지원하지 않는 이미지 형식");
                    ByteBuffer buffer=plane.getBuffer();
                    int fullWidth=rowStride/4;
                    Bitmap padded=Bitmap.createBitmap(fullWidth,height,Bitmap.Config.ARGB_8888);
                    try{
                        padded.copyPixelsFromBuffer(buffer);
                        Bitmap clean=Bitmap.createBitmap(padded,0,0,width,height);
                        try{save(clean);}finally{clean.recycle();}
                    }finally{padded.recycle();}
                    finish("스크린샷 저장 완료 · 사진/PulseNote Deck");
                }catch(Exception error){
                    finish("화면 저장 실패: "+error.getClass().getSimpleName());
                }finally{
                    if(image!=null)image.close();
                }
            },worker);
            display=projection.createVirtualDisplay("PulseNoteCapture",width,height,
                metrics.densityDpi,DisplayManager.VIRTUAL_DISPLAY_FLAG_AUTO_MIRROR,
                reader.getSurface(),null,worker);
        }catch(Exception e){finish("스크린샷 오류: "+e.getClass().getSimpleName());}
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

    private void finish(String result){
        if(!closed.compareAndSet(false,true))return;
        main.post(()->Toast.makeText(getApplicationContext(),result,Toast.LENGTH_LONG).show());
        if(worker!=null){
            worker.post(()->{
                if(display!=null){display.release();display=null;}
                if(reader!=null){reader.close();reader=null;}
                if(projection!=null){
                    projection.stop();projection=null;
                }
                if(thread!=null){thread.quitSafely();thread=null;}
                main.post(()->{
                    stopForeground(STOP_FOREGROUND_REMOVE);
                    stopSelf();
                });
            });
        }else{
            stopForeground(STOP_FOREGROUND_REMOVE);
            stopSelf();
        }
    }

    @Override public void onDestroy(){
        finish("화면 캡처 종료");
        super.onDestroy();
    }
    @Override public IBinder onBind(Intent intent){return null;}
}
