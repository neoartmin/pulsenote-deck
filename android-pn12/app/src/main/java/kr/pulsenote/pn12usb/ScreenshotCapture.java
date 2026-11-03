package kr.pulsenote.pn12usb;

import android.Manifest;
import android.app.*;
import android.content.*;
import android.content.pm.PackageManager;
import android.os.Build;

/**
 * Trigger only in response to an assigned physical key press.
 * Android screen consent always belongs to the OS, never to this helper.
 */
public final class ScreenshotCapture {
    public static final String REQUEST="kr.pulsenote.pn12usb.SCREENSHOT_REQUEST";
    private static final String CHANNEL="pnd_capture";
    private ScreenshotCapture(){}

    public static String request(Context context){
        if(MainActivity.visible){
            context.sendBroadcast(new Intent(REQUEST).setPackage(context.getPackageName()));
            return "화면 캡처 승인 요청";
        }
        if(Build.VERSION.SDK_INT>=33 &&
                context.checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS)!=PackageManager.PERMISSION_GRANTED)
            return "캡처 알림을 표시하려면 Android 알림 권한을 허용하고 앱을 여세요";
        try{
            NotificationManager nm=(NotificationManager)context.getSystemService(Context.NOTIFICATION_SERVICE);
            if(nm==null)return "알림 서비스를 사용할 수 없습니다";
            nm.createNotificationChannel(new NotificationChannel(CHANNEL,"PulseNote 화면 캡처",
                NotificationManager.IMPORTANCE_HIGH));
            Intent open=new Intent(context,MainActivity.class).setAction(REQUEST);
            open.addFlags(Intent.FLAG_ACTIVITY_CLEAR_TOP|Intent.FLAG_ACTIVITY_SINGLE_TOP);
            PendingIntent click=PendingIntent.getActivity(context,77,open,
                PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE);
            nm.notify(77,new Notification.Builder(context,CHANNEL)
                .setSmallIcon(android.R.drawable.ic_menu_camera)
                .setContentTitle("PulseNote 캡처 승인")
                .setContentText("화면을 저장하려면 눌러서 Android 화면 공유를 승인하세요")
                .setContentIntent(click).setAutoCancel(true).build());
            return "캡처 승인 알림을 눌러 계속하세요";
        }catch(RuntimeException e){return "캡처 요청 알림 실패: "+e.getClass().getSimpleName();}
    }
}
