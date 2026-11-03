package kr.pulsenote.pn12usb;

public final class AndroidParitySelfTest {
    private static void check(boolean ok, String message) {
        if (!ok) throw new AssertionError(message);
    }
    public static void main(String[] args) {
        check(AndroidEditShortcut.parse("CTRL+C") == AndroidEditShortcut.Action.COPY, "copy");
        check(AndroidEditShortcut.parse("control + v") == AndroidEditShortcut.Action.PASTE, "paste");
        check(AndroidEditShortcut.parse("Ctrl+X") == AndroidEditShortcut.Action.CUT, "cut");
        check(AndroidEditShortcut.parse(" cTrL + a ") == AndroidEditShortcut.Action.SELECT_ALL, "select all");
        for (String unsupported : new String[]{"CTRL+S","ALT+F4","CTRL+Z","CTRL+C+V","C","", "SHIFT+ENTER"}) {
            check(AndroidEditShortcut.parse(unsupported) == null, "unsafe shortcut accepted: "+unsupported);
        }
        check(AndroidEditShortcut.parse(null) == null, "null shortcut accepted");
        for (int count : new int[]{12,15}) {
            ProfileSnapshot.Key[] source = new ProfileSnapshot.Key[count];
            for (int i=0;i<count;i++) {
                source[i] = new ProfileSnapshot.Key(i%12, "키 "+i,
                    "줄 1\n두 번째 줄 😀 "+i, i%7);
            }
            ProfileSnapshot saved = new ProfileSnapshot(source);
            source[0] = new ProfileSnapshot.Key(0, "다른 값", "", 0);
            check(saved.size()==count, "profile size");
            check(saved.key(0).name.equals("키 0"), "snapshot changed after source mutation");
            check(saved.key(count-1).value.contains("\n"), "multiline macro lost");
            check(saved.key(count-1).color==(count-1)%7, "key color lost");
        }
        check(ProfileSnapshot.MAX_PROFILES==8, "must match Windows profile limit");
        System.out.println("AndroidParitySelfTest OK");
    }
}
