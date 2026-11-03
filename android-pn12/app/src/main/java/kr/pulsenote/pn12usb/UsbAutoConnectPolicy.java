package kr.pulsenote.pn12usb;

/** Decision-only logic for foreground USB auto-connect retries.
 * No permission dialogs are repeated after the user denies a device. */
public final class UsbAutoConnectPolicy {
    public static final long RETRY_AFTER_MILLIS=15000L;
    private UsbAutoConnectPolicy(){}

    public static boolean shouldAttempt(boolean enabled,boolean promptInProgress,
            boolean connected,boolean rejected,long lastAttempt,long now){
        return enabled && !promptInProgress && !connected && !rejected &&
            (lastAttempt<0 || now-lastAttempt>=RETRY_AFTER_MILLIS);
    }
}
