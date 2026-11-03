package kr.pulsenote.pn12usb;

import java.util.Locale;

/**
 * Whitelisted Android editable-field operations. These are accessibility
 * actions, not injected hardware key chords. Arbitrary Ctrl+key combinations
 * and desktop shortcuts are intentionally not advertised as supported.
 */
public final class AndroidEditShortcut {
    public enum Action { COPY, PASTE, CUT, SELECT_ALL }
    public static final String[] PICKER_VALUES = {"CTRL+C","CTRL+V","CTRL+X","CTRL+A"};
    public static final String[] PICKER_LABELS = {
        "복사   ·   Ctrl+C", "붙여넣기   ·   Ctrl+V",
        "잘라내기   ·   Ctrl+X", "전체 선택   ·   Ctrl+A"
    };

    public static int pickerIndex(String existing) {
        Action action=parse(existing);
        if(action==null)return -1;
        for(int i=0;i<PICKER_VALUES.length;i++){
            if(parse(PICKER_VALUES[i])==action)return i;
        }
        return -1;
    }
    private AndroidEditShortcut() {}

    public static Action parse(String raw) {
        if (raw == null) return null;
        String key = raw.toUpperCase(Locale.ROOT).replace(" ", "").replace("\t", "");
        if (key.startsWith("CONTROL+")) key = "CTRL+" + key.substring(8);
        if (key.equals("CTRL+C")) return Action.COPY;
        if (key.equals("CTRL+V")) return Action.PASTE;
        if (key.equals("CTRL+X")) return Action.CUT;
        if (key.equals("CTRL+A")) return Action.SELECT_ALL;
        return null;
    }
}
