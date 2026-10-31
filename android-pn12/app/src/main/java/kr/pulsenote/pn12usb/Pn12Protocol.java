package kr.pulsenote.pn12usb;

import java.nio.charset.StandardCharsets;

public final class Pn12Protocol {
    public interface Listener {
        void onHandshake();
        void onKey(int key);
    }
    public static final byte[] HELLO = "PND:HELLO\n".getBytes(StandardCharsets.US_ASCII);
    private final Listener listener;
    private final StringBuilder line = new StringBuilder();
    private boolean ready, info, announced, dropping;

    public Pn12Protocol(Listener callback) { listener = callback; }
    public boolean isVerified() { return ready && info; }
    public void receive(byte[] data, int size) {
        for (int i = 0; i < Math.max(0, Math.min(size, data.length)); i++) {
            int c = data[i] & 0xff;
            if (c == '\r') continue;
            if (c == '\n') {
                if (!dropping) parse(line.toString());
                line.setLength(0); dropping = false;
            } else if (!dropping && c >= 32 && c <= 126) {
                if (line.length() < 120) line.append((char)c);
                else { dropping = true; line.setLength(0); }
            }
        }
    }
    private void parse(String s) {
        if ("PND:READY:2.1".equals(s)) ready = true;
        else if ("PND:INFO:12:C3".equals(s)) info = true;
        else if (s.startsWith("KEY:") && isVerified()) {
            try {
                String digits = s.substring(4);
                if (!digits.matches("(?:[1-9]|1[0-2])")) return;
                listener.onKey(Integer.parseInt(digits));
            } catch (NumberFormatException ignored) {}
        }
        if (isVerified() && !announced) { announced = true; listener.onHandshake(); }
    }
}
