package kr.pulsenote.pn12usb;

import android.app.*;
import android.content.*;
import android.media.AudioManager;
import android.net.Uri;
import android.os.Build;
import android.os.SystemClock;
import android.view.KeyEvent;

public final class ActionDispatcher {
    private static final String CHANNEL = "pn12_shortcuts";
    private ActionDispatcher() {}

    public static String run(Context context, int key) {
        ActionStore settings = new ActionStore(context);
        int kind = settings.kind(key);
        AudioManager audio = (AudioManager) context.getSystemService(Context.AUDIO_SERVICE);
        if (kind >= 1 && kind <= 3) {
            if (audio == null) return "오디오 서비스 없음";
            int op = kind == 1 ? AudioManager.ADJUST_RAISE :
                     kind == 2 ? AudioManager.ADJUST_LOWER : AudioManager.ADJUST_TOGGLE_MUTE;
            audio.adjustStreamVolume(AudioManager.STREAM_MUSIC, op, AudioManager.FLAG_SHOW_UI);
            return ActionStore.NAMES[kind];
        }
        if (kind >= 4 && kind <= 6) {
            if (audio == null) return "오디오 서비스 없음";
            int code = kind == 4 ? KeyEvent.KEYCODE_MEDIA_PLAY_PAUSE :
                       kind == 5 ? KeyEvent.KEYCODE_MEDIA_NEXT : KeyEvent.KEYCODE_MEDIA_PREVIOUS;
            long t = SystemClock.uptimeMillis();
            audio.dispatchMediaKeyEvent(new KeyEvent(t, t, KeyEvent.ACTION_DOWN, code, 0));
            audio.dispatchMediaKeyEvent(new KeyEvent(t, t, KeyEvent.ACTION_UP, code, 0));
            return ActionStore.NAMES[kind];
        }
        if (kind == 0) return "기능 미지정";
        String value = settings.value(key).trim();
        Intent target;
        if (kind == 7) {
            Uri u = Uri.parse(value);
            if (u.getScheme() == null || u.getHost() == null ||
                !(u.getScheme().equalsIgnoreCase("https") || u.getScheme().equalsIgnoreCase("http")))
                return "올바른 URL을 입력하세요";
            target = new Intent(Intent.ACTION_VIEW, u);
        } else {
            target = context.getPackageManager().getLaunchIntentForPackage(value);
            if (target == null) return "설치된 앱을 찾지 못함";
        }
        target.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK);
        if (MainActivity.visible) {
            try { context.startActivity(target); return "실행 요청"; }
            catch (RuntimeException e) { return "실행 실패: " + e.getClass().getSimpleName(); }
        }
        NotificationManager nm = (NotificationManager) context.getSystemService(Context.NOTIFICATION_SERVICE);
        if (nm == null || (Build.VERSION.SDK_INT >= 33 && !nm.areNotificationsEnabled()))
            return "알림 권한 필요";
        nm.createNotificationChannel(new NotificationChannel(CHANNEL, "PN12 앱 실행", NotificationManager.IMPORTANCE_HIGH));
        try {
            PendingIntent launch = PendingIntent.getActivity(context, 1000 + key, target,
                PendingIntent.FLAG_UPDATE_CURRENT | PendingIntent.FLAG_IMMUTABLE);
            Notification notice = new Notification.Builder(context, CHANNEL)
              .setSmallIcon(android.R.drawable.ic_menu_view)
              .setContentTitle("PN12 " + key + "번: " + ActionStore.NAMES[kind])
              .setContentText("눌러서 실행")
              .setContentIntent(launch).setAutoCancel(true).build();
            nm.notify(1000 + key, notice);
            return "알림에서 실행";
        } catch (RuntimeException e) { return "실행 알림 실패"; }
    }
}
