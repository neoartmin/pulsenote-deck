package kr.pulsenote.pn12usb;

/** Immutable snapshot of the current device profile before creating a copy. */
public final class ProfileSnapshot {
    public static final int MAX_PROFILES = 8;

    public static final class Key {
        public final int kind;
        public final String name;
        public final String value;
        public final int color;

        public Key(int kind, String name, String value, int color) {
            this.kind = kind;
            this.name = name;
            this.value = value;
            this.color = color;
        }
    }

    private final Key[] keys;

    public ProfileSnapshot(Key[] source) {
        if (source == null || (source.length != 12 && source.length != 15))
            throw new IllegalArgumentException("Unsupported device key count");
        keys = source.clone();
        for (Key key : keys) if (key == null)
            throw new IllegalArgumentException("Missing key configuration");
    }

    public int size() { return keys.length; }
    public Key key(int zeroBasedIndex) { return keys[zeroBasedIndex]; }
}
