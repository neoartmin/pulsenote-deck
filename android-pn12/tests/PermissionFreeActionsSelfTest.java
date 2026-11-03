package kr.pulsenote.pn12usb;

/** Standalone JVM regression tests requiring neither Android runtime nor USB. */
public final class PermissionFreeActionsSelfTest {
    private static void check(boolean ok,String detail){
        if(!ok)throw new AssertionError(detail);
    }
    public static void main(String[] ignored){
        check(PermissionFreeActions.COPY_TEXT==12,"keep saved 0..11 action IDs");
        check(PermissionFreeActions.SYSTEM_SETTINGS==13,"settings ID");
        check(PermissionFreeActions.OPEN_TIMERS==14,"timers ID");
        check(PermissionFreeActions.OPEN_ALARMS==15,"alarms ID");
        check(PermissionFreeActions.SETTINGS_VALUES.length==6,"six settings shortcuts");
        check(PermissionFreeActions.SETTINGS_LABELS.length==6,"labels are 1:1");
        String[] expected={
            "android.settings.WIFI_SETTINGS",
            "android.settings.BLUETOOTH_SETTINGS",
            "android.settings.DISPLAY_SETTINGS",
            "android.settings.SOUND_SETTINGS",
            "android.settings.BATTERY_SAVER_SETTINGS",
            "android.settings.SETTINGS"
        };
        for(int i=0;i<expected.length;i++){
            String saved=PermissionFreeActions.SETTINGS_VALUES[i];
            check(PermissionFreeActions.settingsIndex(saved)==i,"roundtrip "+i);
            check(PermissionFreeActions.settingsIntentAction(saved).equals(expected[i]),
                "target action "+i);
            check(!PermissionFreeActions.settingsLabel(saved).isEmpty(),
                "label "+i);
        }
        check(PermissionFreeActions.settingsIndex("unknown")==5,"safe settings fallback");
        check(PermissionFreeActions.settingsIndex(null)==5,"null-safe settings fallback");
        check(PermissionFreeActions.clipboardError(null)!=null,"null text");
        check(PermissionFreeActions.clipboardError("")!=null,"empty text");
        check(PermissionFreeActions.clipboardError("두 줄\n두 번째 ✅")==null,
            "multiline and unicode clipboard");
        check(PermissionFreeActions.clipboardError(" ".repeat(3))==null,
            "preserve intentional whitespace");
        check(PermissionFreeActions.clipboardError("A".repeat(4096))==null,
            "boundary 4096");
        check(PermissionFreeActions.clipboardError("A".repeat(4097))!=null,
            "reject over-long clipboard text");
        System.out.println("PermissionFreeActionsSelfTest OK");
    }
}
