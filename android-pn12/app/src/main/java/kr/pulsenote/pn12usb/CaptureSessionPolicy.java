package kr.pulsenote.pn12usb;

/**
 * Decides what to do when a hardware screenshot key is pressed.
 * A running MediaProjection session always takes precedence over another
 * consent dialog; a fresh consent is requested only after that session ends.
 */
public final class CaptureSessionPolicy {
    public enum Route { CAPTURE, PROMPT, NOTIFY, NOTIFICATION_PERMISSION_REQUIRED }
    private CaptureSessionPolicy() {}

    public static Route decide(boolean running, boolean starting,
                               boolean activityVisible, boolean notificationsAllowed) {
        if (running || starting) return Route.CAPTURE;
        if (activityVisible) return Route.PROMPT;
        return notificationsAllowed ? Route.NOTIFY : Route.NOTIFICATION_PERMISSION_REQUIRED;
    }
}
