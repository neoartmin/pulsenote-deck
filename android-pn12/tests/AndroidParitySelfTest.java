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
        check(AndroidEditShortcut.PICKER_VALUES.length==4,"only four supported Android shortcuts");
        check(AndroidEditShortcut.PICKER_LABELS.length==AndroidEditShortcut.PICKER_VALUES.length,
            "wheel labels and actions differ");
        for(int i=0;i<AndroidEditShortcut.PICKER_VALUES.length;i++)
            check(AndroidEditShortcut.pickerIndex(AndroidEditShortcut.PICKER_VALUES[i])==i,
                "wheel index round-trip "+i);
        check(AndroidEditShortcut.pickerIndex("CTRL+S")==-1,"unsupported shortcut should stay unselected");
        check(UsbAutoConnectPolicy.shouldAttempt(true,false,false,false,-1,1000),
            "initial USB auto connect");
        check(!UsbAutoConnectPolicy.shouldAttempt(false,false,false,false,-1,1000),
            "disabled auto connect must be respected");
        check(!UsbAutoConnectPolicy.shouldAttempt(true,true,false,false,-1,1000),
            "never prompt twice");
        check(!UsbAutoConnectPolicy.shouldAttempt(true,false,true,false,-1,1000),
            "already connected");
        check(!UsbAutoConnectPolicy.shouldAttempt(true,false,false,true,-1,1000),
            "user rejected permission");
        check(!UsbAutoConnectPolicy.shouldAttempt(true,false,false,false,1000,5000),
            "USB handshake retry is too early");
        check(UsbAutoConnectPolicy.shouldAttempt(true,false,false,false,1000,
            1000+UsbAutoConnectPolicy.RETRY_AFTER_MILLIS),"USB retry after timeout");

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
