package kr.pulsenote.pn12usb;

import android.Manifest;
import android.app.*;
import android.content.*;
import android.content.pm.PackageManager;
import android.os.Build;

/**
 * Hardware key entry point. Reuse the active MediaProjection screen session
 * for each press. Only the FIRST press after a session ends requests OS consent.
 */
public final class ScreenshotCapture {
    public static final String REQUEST="kr.pulsenote.pn12usb.SCREENSHOT_REQUEST";
    private static final String CHANNEL="pnd_capture";
    private ScreenshotCapture(){}

    public static String request(Context context){
        boolean notifications=Build.VERSION.SDK_INT<33 ||
            context.checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS)==PackageManager.PERMISSION_GRANTED;
        CaptureSessionPolicy.Route route=CaptureSessionPolicy.decide(
            ScreenshotCaptureService.isSessionReady(),
            ScreenshotCaptureService.isSessionStarting(),
            MainActivity.visible,notifications);
        if(route==CaptureSessionPolicy.Route.CAPTURE){
            try{
                context.startService(new Intent(context,ScreenshotCaptureService.class)
                    .setAction(ScreenshotCaptureService.CAPTURE));
                return "캡처 세션에서 PNG 저장 요청";
            }catch(RuntimeException error){
                return "캡처 세션에 연결하지 못했습니다: "+error.getClass().getSimpleName();
            }
        }
        if(route==CaptureSessionPolicy.Route.PROMPT){
            context.sendBroadcast(new Intent(REQUEST).setPackage(context.getPackageName()));
            return "화면 공유 최초 승인 요청";
        }
        if(route==CaptureSessionPolicy.Route.NOTIFICATION_PERMISSION_REQUIRED)
            return "화면 캡처 최초 승인 알림을 위해 Android 알림 권한을 허용하고 앱을 여세요";
        try{
            NotificationManager nm=(NotificationManager)context.getSystemService(Context.NOTIFICATION_SERVICE);
            if(nm==null)return "알림 서비스를 사용할 수 없습니다";
            nm.createNotificationChannel(new NotificationChannel(CHANNEL,
                "PulseNote 화면 캡처 승인",NotificationManager.IMPORTANCE_HIGH));
            Intent open=new Intent(context,MainActivity.class).setAction(REQUEST);
            open.addFlags(Intent.FLAG_ACTIVITY_CLEAR_TOP|Intent.FLAG_ACTIVITY_SINGLE_TOP);
            PendingIntent click=PendingIntent.getActivity(context,77,open,
                PendingIntent.FLAG_UPDATE_CURRENT|PendingIntent.FLAG_IMMUTABLE);
            nm.notify(77,new Notification.Builder(context,CHANNEL)
                .setSmallIcon(android.R.drawable.ic_menu_camera)
                .setContentTitle("PulseNote · 캡처 모드 시작")
                .setContentText("한 번 화면 공유를 승인하면 캡처 버튼을 반복 사용할 수 있습니다")
                .setContentIntent(click).setAutoCancel(true).build());
            return "첫 캡처: 승인 알림을 눌러 시작";
        }catch(RuntimeException e){
            return "캡처 승인 알림 실패: "+e.getClass().getSimpleName();
        }
    }
}
