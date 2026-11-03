package kr.pulsenote.pn12usb;

/**
 * Standard edition intentionally has no accessibility service, related APIs,
 * editor shortcut picker, or sensitive runtime permission.
 */
public final class EditShortcutBridge {
    private EditShortcutBridge(){}
    public static final String[] PICKER_VALUES={};
    public static final String[] PICKER_LABELS={};
    public static boolean enabled(){return false;}
    public static String shortcut(String value){return "기본 설치판에서 지원하지 않는 기능입니다";}
    public static int pickerIndex(String value){return -1;}
}
