package kr.pulsenote.pn12usb;

import java.util.Locale;

/**
 * Whitelisted Android editable-field operations. These are accessibility
 * actions, not injected hardware key chords. Arbitrary Ctrl+key combinations
 * and desktop shortcuts are intentionally not advertised as supported.
 */
public final class AndroidEditShortcut {
    public enum Action { COPY, PASTE, CUT, SELECT_ALL }
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
