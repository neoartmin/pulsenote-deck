package kr.pulsenote.pn12usb;

import android.content.Context;
import android.content.SharedPreferences;

public final class ActionStore {
    public static final String[] NAMES = {
        "기능 없음", "볼륨 올리기", "볼륨 내리기", "음소거 전환",
        "재생 / 일시정지", "다음 곡", "이전 곡", "웹사이트 열기", "앱 실행"
    };
    private static final int[] DEFAULTS = {4, 5, 6, 1, 2, 3, 7, 0, 0, 0, 0, 0};
    private final SharedPreferences prefs;
    public ActionStore(Context context) {
        prefs = context.getSharedPreferences("pn12_actions", Context.MODE_PRIVATE);
    }
    public int kind(int key) {
        return prefs.getInt("type_" + key, DEFAULTS[Math.max(0, Math.min(11, key - 1))]);
    }
    public String value(int key) { return prefs.getString("value_" + key, ""); }
    public void set(int key, int kind, String value) {
        if (key < 1 || key > 12 || kind < 0 || kind >= NAMES.length) return;
        prefs.edit().putInt("type_" + key, kind).putString("value_" + key, value).apply();
    }
    public String label(int key) {
        int k = kind(key);
        String v = value(key);
        return NAMES[k] + (k >= 7 && !v.isEmpty() ? "\n" + v : "");
    }
}
