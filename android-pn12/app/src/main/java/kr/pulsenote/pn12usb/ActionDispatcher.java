package kr.pulsenote.pn12usb;

import android.app.*;
import android.content.*;
import android.media.AudioManager;
import android.net.Uri;
import android.os.Build;
import android.os.SystemClock;
import android.provider.DocumentsContract;
import android.view.KeyEvent;

public final class ActionDispatcher {
    private static final String CHANNEL="pn12_shortcuts";
    private ActionDispatcher(){}

    public static String run(Context context,int key) {
        ActionStore s=new ActionStore(context);
        int kind=s.kind(key);
        if(kind==0)return "기능 없음";
        AudioManager audio=(AudioManager)context.getSystemService(Context.AUDIO_SERVICE);
        if(kind>=1&&kind<=3) {
            if(audio==null)return "오디오 서비스 없음";
            int dir=kind==1?AudioManager.ADJUST_RAISE:kind==2?AudioManager.ADJUST_LOWER:AudioManager.ADJUST_TOGGLE_MUTE;
            audio.adjustStreamVolume(AudioManager.STREAM_MUSIC,dir,AudioManager.FLAG_SHOW_UI);
            return ActionStore.NAMES[kind];
        }
        if(kind>=4&&kind<=6) {
            if(audio==null)return "오디오 서비스 없음";
            int code=kind==4?KeyEvent.KEYCODE_MEDIA_PLAY_PAUSE:kind==5?KeyEvent.KEYCODE_MEDIA_NEXT:KeyEvent.KEYCODE_MEDIA_PREVIOUS;
            long now=SystemClock.uptimeMillis();
            audio.dispatchMediaKeyEvent(new KeyEvent(now,now,KeyEvent.ACTION_DOWN,code,0));
            audio.dispatchMediaKeyEvent(new KeyEvent(now,now,KeyEvent.ACTION_UP,code,0));
            return ActionStore.NAMES[kind];
        }
        if(kind==9)return "시스템 단축키는 별도 입력 방식이 필요합니다";
        String value=s.value(key).trim();
        Intent target=null;
        if(kind==7) {
            Uri uri=Uri.parse(value);
            if(uri.getHost()==null||uri.getScheme()==null||!(uri.getScheme().equalsIgnoreCase("https")||uri.getScheme().equalsIgnoreCase("http")))return "올바른 URL을 설정하세요";
            target=new Intent(Intent.ACTION_VIEW,uri);
        }else if(kind==8) {
            if(value.isEmpty())return "실행 앱을 선택하세요";
            target=context.getPackageManager().getLaunchIntentForPackage(value);
            if(target==null)return "설치된 Android 앱을 찾지 못함";
        }else if(kind==10) {
            target=new Intent(Intent.ACTION_OPEN_DOCUMENT_TREE);
            if(!value.isEmpty())try{target.putExtra(DocumentsContract.EXTRA_INITIAL_URI,Uri.parse(value));}catch(Exception ignored){}
            target.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION|Intent.FLAG_GRANT_WRITE_URI_PERMISSION);
        }
        if(target==null)return "지원되지 않는 동작";
        target.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
        if(MainActivity.visible) {
            try{context.startActivity(target);return "직접 실행 요청";}
            catch(RuntimeException ex){return "앱 실행 실패: "+ex.getClass().getSimpleName();}
        }
        // The PN12 hardware key is the user action. In direct mode the user
        // has explicitly opted in and the app shows a visible overlay window.
        // Send an activity PendingIntent with BAL opt-in on Android 14+.
        // OS versions can still refuse to surface the activity.
        boolean overlayAllowed=DirectLaunchOverlay.allowed(context);
        if(overlayAllowed && DirectLaunchOverlay.isVisible()){
            try{
                // Android 16 (API 36) introduces ALLOW_ALWAYS (value 3) for
                // explicit connected-device actions. compileSdk is currently 35,
                // so use its documented numeric value until upgrading the SDK.
                // Earlier versions retain the API 34 allowed mode.
                int balMode=Build.VERSION.SDK_INT>=36 ? 3 :
                            ActivityOptions.MODE_BACKGROUND_ACTIVITY_START_ALLOWED;
                android.os.Bundle creatorOptions=null;
                if(Build.VERSION.SDK_INT>=35){
                    creatorOptions=ActivityOptions.makeBasic()
                        .setPendingIntentCreatorBackgroundActivityStartMode(balMode).toBundle();
                }
                PendingIntent pending=PendingIntent.getActivity(context,2000+key,target,
                    PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE,
                    creatorOptions);
                if(Build.VERSION.SDK_INT>=34){
                    ActivityOptions options=ActivityOptions.makeBasic();
                    options.setPendingIntentBackgroundActivityStartMode(balMode);
                    pending.send(context,0,null,null,null,null,options.toBundle());
                }else pending.send();
                return "백그라운드 직접 실행 전송";
            }catch(PendingIntent.CanceledException|RuntimeException ignored){
                // Fallback to user-facing notification.
            }
        }
        NotificationManager nm=(NotificationManager)context.getSystemService(Context.NOTIFICATION_SERVICE);
        if(nm==null||(Build.VERSION.SDK_INT>=33&&!nm.areNotificationsEnabled()))return "앱 실행용 알림 권한 필요";
        nm.createNotificationChannel(new NotificationChannel(CHANNEL,"PN12 앱 열기",NotificationManager.IMPORTANCE_HIGH));
        try {
            PendingIntent pi=PendingIntent.getActivity(context,1000+key,target,PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE);
            String reason=!android.provider.Settings.canDrawOverlays(context)?
                    "화면 위 표시 권한이 없어 직접 실행을 대체합니다":
                    !overlayAllowed?"직접 실행 모드가 꺼져 있습니다":
                    "직접 실행을 사용할 수 없어 알림으로 대체합니다";
            nm.notify(1000+key,new Notification.Builder(context,CHANNEL)
              .setSmallIcon(android.R.drawable.ic_menu_view).setContentTitle("PulseNote · "+s.name(key))
              .setContentText(reason+" · 탭하여 실행")
              .setContentIntent(pi).setAutoCancel(true).build());
            return "알림을 눌러 실행";
        }catch(Exception ex){return "실행 알림 실패";}
    }
}
