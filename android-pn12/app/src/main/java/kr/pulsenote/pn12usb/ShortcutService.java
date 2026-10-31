package kr.pulsenote.pn12usb;

import android.accessibilityservice.AccessibilityService;
import android.view.accessibility.AccessibilityEvent;
import android.view.accessibility.AccessibilityNodeInfo;
import android.os.Bundle;
import android.os.Handler;
import android.os.Looper;
import java.util.Locale;

public final class ShortcutService extends AccessibilityService {
    private static volatile ShortcutService current;
    @Override protected void onServiceConnected() { super.onServiceConnected();current=this; }
    @Override public void onAccessibilityEvent(AccessibilityEvent e) {}
    @Override public void onInterrupt() {}
    @Override public void onDestroy(){if(current==this)current=null;super.onDestroy();}
    public static boolean enabled(){return current!=null;}
    public static String execute(String value) {
        final ShortcutService svc=current;
        if(svc==null)return "단축키: 접근성 서비스 활성화 필요";
        String shortcut=value.trim().toUpperCase(Locale.ROOT).replace(" ","").replace("CONTROL","CTRL");
        final int action;
        if(shortcut.equals("CTRL+C"))action=AccessibilityNodeInfo.ACTION_COPY;
        else if(shortcut.equals("CTRL+V"))action=AccessibilityNodeInfo.ACTION_PASTE;
        else if(shortcut.equals("CTRL+X"))action=AccessibilityNodeInfo.ACTION_CUT;
        else if(shortcut.equals("CTRL+A"))action=AccessibilityNodeInfo.ACTION_SET_SELECTION;
        else return "Android 미지원 단축키: "+value;
        new Handler(Looper.getMainLooper()).post(()->{
            AccessibilityNodeInfo root=svc.getRootInActiveWindow();
            if(root==null)return;
            AccessibilityNodeInfo focused=root.findFocus(AccessibilityNodeInfo.FOCUS_INPUT);
            if(focused!=null){
                if(action==AccessibilityNodeInfo.ACTION_SET_SELECTION){
                    CharSequence currentText=focused.getText();
                    if(currentText!=null){
                        Bundle args=new Bundle();
                        args.putInt(AccessibilityNodeInfo.ACTION_ARGUMENT_SELECTION_START_INT,0);
                        args.putInt(AccessibilityNodeInfo.ACTION_ARGUMENT_SELECTION_END_INT,currentText.length());
                        focused.performAction(action,args);
                    }
                }else focused.performAction(action);
            }
        });
        return "단축키 요청 · "+shortcut;
    }
}
