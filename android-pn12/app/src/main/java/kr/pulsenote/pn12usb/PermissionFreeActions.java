package kr.pulsenote.pn12usb;

/**
 * Android actions that need no new manifest, AccessibilityService, overlay,
 * or dangerous runtime permission. Existing USB device access is separate.
 *
 * Keep numeric action IDs stable: all prior profiles use IDs 0..11.
 */
public final class PermissionFreeActions {
    public static final int COPY_TEXT=12;
    public static final int SYSTEM_SETTINGS=13;
    public static final int OPEN_TIMERS=14;
    public static final int OPEN_ALARMS=15;

    public static final int MAX_CLIPBOARD_CHARS=4096;

    public static final String[] SETTINGS_VALUES={
        "wifi","bluetooth","display","sound","battery","general"
    };
    public static final String[] SETTINGS_LABELS={
        "Wi-Fi 설정","블루투스 설정","화면 설정","소리 설정",
        "배터리 절약 설정","전체 설정"
    };

    private PermissionFreeActions(){}

    public static int settingsIndex(String value){
        if(value!=null)for(int i=0;i<SETTINGS_VALUES.length;i++)
            if(SETTINGS_VALUES[i].equals(value))return i;
        return SETTINGS_VALUES.length-1;
    }

    public static String settingsIntentAction(String value){
        switch(SETTINGS_VALUES[settingsIndex(value)]){
            case "wifi":return "android.settings.WIFI_SETTINGS";
            case "bluetooth":return "android.settings.BLUETOOTH_SETTINGS";
            case "display":return "android.settings.DISPLAY_SETTINGS";
            case "sound":return "android.settings.SOUND_SETTINGS";
            case "battery":return "android.settings.BATTERY_SAVER_SETTINGS";
            default:return "android.settings.SETTINGS";
        }
    }

    public static String settingsLabel(String value){
        return SETTINGS_LABELS[settingsIndex(value)];
    }

    public static String clipboardError(String text){
        if(text==null||text.isEmpty())return "클립보드에 복사할 문구를 입력하세요";
        if(text.length()>MAX_CLIPBOARD_CHARS)
            return "복사할 문구는 "+MAX_CLIPBOARD_CHARS+"자 이하로 입력하세요";
        return null;
    }
}
