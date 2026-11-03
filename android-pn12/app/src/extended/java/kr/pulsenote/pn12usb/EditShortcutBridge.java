package kr.pulsenote.pn12usb;

/** The optional, explicit-consent edit shortcut capability. */
public final class EditShortcutBridge {
    private EditShortcutBridge(){}
    public static final String[] PICKER_VALUES=AndroidEditShortcut.PICKER_VALUES;
    public static final String[] PICKER_LABELS=AndroidEditShortcut.PICKER_LABELS;
    public static boolean enabled(){return TextMacroAccessibilityService.enabled();}
    public static String shortcut(String value){return TextMacroAccessibilityService.shortcut(value);}
    public static int pickerIndex(String value){return AndroidEditShortcut.pickerIndex(value);}
}
