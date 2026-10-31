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
        if(kind==9)return ShortcutService.execute(s.value(key));
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
        if(MainActivity.visible || DirectLaunchOverlay.isVisible()) {
            try {
                context.startActivity(target);
                return MainActivity.visible ? "실행 요청" : "백그라운드 실행 요청";
            }catch(RuntimeException ex){
                // Android can deny background activity launches; keep the
                // user-interaction notification fallback when direct launch fails.
            }
        }
        NotificationManager nm=(NotificationManager)context.getSystemService(Context.NOTIFICATION_SERVICE);
        if(nm==null||(Build.VERSION.SDK_INT>=33&&!nm.areNotificationsEnabled()))return "앱 실행용 알림 권한 필요";
        nm.createNotificationChannel(new NotificationChannel(CHANNEL,"PN12 앱 열기",NotificationManager.IMPORTANCE_HIGH));
        try {
            PendingIntent pi=PendingIntent.getActivity(context,1000+key,target,PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE);
            nm.notify(1000+key,new Notification.Builder(context,CHANNEL)
              .setSmallIcon(android.R.drawable.ic_menu_view).setContentTitle("PN12 · "+s.name(key))
              .setContentText("탭하여 "+ActionStore.NAMES[kind]+" 실행")
              .setContentIntent(pi).setAutoCancel(true).build());
            return "알림을 눌러 실행";
        }catch(Exception ex){return "실행 알림 실패";}
    }
}
