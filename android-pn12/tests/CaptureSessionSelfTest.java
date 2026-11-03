package kr.pulsenote.pn12usb;

public final class CaptureSessionSelfTest {
    private static void check(boolean ok,String message){
        if(!ok)throw new AssertionError(message);
    }
    public static void main(String[] args){
        // First press after a fresh start requires one OS consent dialog.
        check(CaptureSessionPolicy.decide(false,false,true,true)==
            CaptureSessionPolicy.Route.PROMPT,"first foreground press");
        check(CaptureSessionPolicy.decide(false,false,false,true)==
            CaptureSessionPolicy.Route.NOTIFY,"first background press");
        check(CaptureSessionPolicy.decide(false,false,false,false)==
            CaptureSessionPolicy.Route.NOTIFICATION_PERMISSION_REQUIRED,
            "background initial consent must not start silently");

        // One approved projection permits any number of subsequent captures,
        // even when MainActivity is not visible and notifications are denied.
        for(int i=0;i<100;i++){
            check(CaptureSessionPolicy.decide(true,false,false,false)==
                CaptureSessionPolicy.Route.CAPTURE,"active session #"+i);
            check(CaptureSessionPolicy.decide(true,false,true,true)==
                CaptureSessionPolicy.Route.CAPTURE,"visible active session #"+i);
        }
        check(CaptureSessionPolicy.decide(false,true,false,true)==
            CaptureSessionPolicy.Route.CAPTURE,"initialization must not re-prompt");
        // On stop or system projection revocation the next press must consent
        // again, never reusing an old projection result Intent.
        check(CaptureSessionPolicy.decide(false,false,true,true)==
            CaptureSessionPolicy.Route.PROMPT,"after stopping projection");
        System.out.println("CaptureSessionSelfTest OK");
    }
}
