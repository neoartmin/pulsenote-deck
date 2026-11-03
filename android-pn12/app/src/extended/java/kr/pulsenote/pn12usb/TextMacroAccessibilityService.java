package kr.pulsenote.pn12usb;

import android.accessibilityservice.AccessibilityService;
import android.os.Bundle;
import android.view.accessibility.AccessibilityEvent;
import android.view.accessibility.AccessibilityNodeInfo;

/**
 * Optional, user-enabled service. It never monitors or stores screen content.
 * It only performs user-assigned hardware-button actions on focused editable
 * fields: text insertion and a small whitelist of Android edit operations.
 * Text is never stored or transmitted over the network.
 */
public final class TextMacroAccessibilityService extends AccessibilityService {
    private static volatile TextMacroAccessibilityService active;

    public static String insert(String value) {
        TextMacroAccessibilityService service=active;
        if(service==null)return "텍스트 자동 입력을 사용하려면 장치 및 설정에서 접근성 서비스를 활성화하세요";
        return service.insertNow(value);
    }
    public static String shortcut(String value) {
        AndroidEditShortcut.Action operation=AndroidEditShortcut.parse(value);
        if(operation==null)return "Android 편집 단축키는 CTRL+C / CTRL+V / CTRL+X / CTRL+A만 지원합니다";
        TextMacroAccessibilityService service=active;
        if(service==null)return "단축키를 실행하려면 장치 및 설정에서 접근성 서비스를 활성화하세요";
        return service.shortcutNow(operation);
    }
    public static boolean enabled(){return active!=null;}

    @Override protected void onServiceConnected() {
        super.onServiceConnected();
        active=this;
    }
    @Override public void onAccessibilityEvent(AccessibilityEvent event) {
        // Intentionally do not inspect, collect, log, or transmit events.
    }
    @Override public void onInterrupt() {}
    @Override public void onDestroy() {
        if(active==this)active=null;
        super.onDestroy();
    }

    private String shortcutNow(AndroidEditShortcut.Action operation) {
        AccessibilityNodeInfo root=getRootInActiveWindow();
        if(root==null)return "현재 앱의 입력창을 찾을 수 없습니다";
        try{
            AccessibilityNodeInfo input=root.findFocus(AccessibilityNodeInfo.FOCUS_INPUT);
            if(input==null)return "일반 텍스트 입력창을 선택한 뒤 사용하세요";
            try{
                if(!input.isEditable() || input.isPassword())
                    return "일반 편집 가능한 텍스트 입력창에서만 지원합니다 (암호 필드 제외)";
                CharSequence owner=input.getPackageName();
                if(owner!=null&&getPackageName().contentEquals(owner))
                    return "PulseNote 앱 자체의 편집창에서는 단축키를 실행하지 않습니다";
                int action;
                Bundle args=null;
                switch(operation){
                    case COPY:
                    case CUT:
                        int start=input.getTextSelectionStart();
                        int end=input.getTextSelectionEnd();
                        if(start<0||end<0||start==end)return "복사·잘라내기는 먼저 텍스트를 선택하세요";
                        action=operation==AndroidEditShortcut.Action.COPY?
                            AccessibilityNodeInfo.ACTION_COPY:AccessibilityNodeInfo.ACTION_CUT;
                        break;
                    case PASTE:
                        action=AccessibilityNodeInfo.ACTION_PASTE;
                        break;
                    case SELECT_ALL:
                        CharSequence text=input.getText();
                        if(text==null)return "현재 입력창의 텍스트를 선택할 수 없습니다";
                        action=AccessibilityNodeInfo.ACTION_SET_SELECTION;
                        args=new Bundle();
                        args.putInt(AccessibilityNodeInfo.ACTION_ARGUMENT_SELECTION_START_INT,0);
                        args.putInt(AccessibilityNodeInfo.ACTION_ARGUMENT_SELECTION_END_INT,text.length());
                        break;
                    default:
                        return "지원되지 않는 편집 동작입니다";
                }
                if(!input.performAction(action,args))
                    return "현재 앱 입력창이 선택한 편집 단축키를 지원하지 않습니다";
                switch(operation){
                    case COPY:return "선택한 텍스트 복사";
                    case CUT:return "선택한 텍스트 잘라내기";
                    case PASTE:return "클립보드 내용 붙여넣기";
                    case SELECT_ALL:return "입력창 전체 선택";
                    default:return "편집 완료";
                }
            }finally{input.recycle();}
        }finally{root.recycle();}
    }

    private String insertNow(String raw) {
        String macro=TextMacro.expand(raw);
        if(macro.isEmpty())return "입력할 텍스트가 없습니다";
        if(macro.getBytes(java.nio.charset.StandardCharsets.UTF_8).length>4096)
            return "텍스트 매크로는 4096바이트 이하만 지원합니다";
        AccessibilityNodeInfo root=getRootInActiveWindow();
        if(root==null)return "활성 입력창을 찾지 못했습니다";
        try {
            AccessibilityNodeInfo input=root.findFocus(AccessibilityNodeInfo.FOCUS_INPUT);
            if(input==null)return "텍스트 입력창을 선택한 뒤 키를 누르세요";
            try{
                if(!input.isEditable() || input.isPassword())
                    return "일반 텍스트 입력창에서만 사용할 수 있습니다 (암호 필드 제외)";
                CharSequence owner=input.getPackageName();
                if(owner!=null&&getPackageName().contentEquals(owner))
                    return "다른 앱의 입력창을 선택한 뒤 사용하세요";
                CharSequence current=input.getText();
                // ACTION_SET_TEXT replaces the entire content. Preserve the
                // existing text and insert at the selection/caret instead.
                if(current==null)return "이 입력창은 기존 텍스트를 읽을 수 없어 안전하게 삽입할 수 없습니다";
                String existing=current.toString();
                int start=input.getTextSelectionStart();
                int end=input.getTextSelectionEnd();
                if(start<0||end<0||start>existing.length()||end>existing.length()){
                    start=existing.length();end=start;
                }
                if(start>end){int tmp=start;start=end;end=tmp;}
                String combined=existing.substring(0,start)+macro+existing.substring(end);
                Bundle args=new Bundle();
                args.putCharSequence(AccessibilityNodeInfo.ACTION_ARGUMENT_SET_TEXT_CHARSEQUENCE,combined);
                if(!input.performAction(AccessibilityNodeInfo.ACTION_SET_TEXT,args))
                    return "현재 앱의 입력창에서 자동 텍스트 삽입을 지원하지 않습니다";
                // ACTION_SET_TEXT is not a physical Enter key. Some chat apps
                // accept it but silently remove LF. Detect an observable
                // mismatch rather than claiming the newline is guaranteed.
                if(macro.indexOf('\n')>=0 && input.refresh()){
                    CharSequence observed=input.getText();
                    if(observed!=null && !combined.contentEquals(observed))
                        return "텍스트 삽입 요청 완료 · 대상 앱에서 줄바꿈 유지 여부를 확인하세요";
                }
                return "텍스트 입력 완료";
            }finally{input.recycle();}
        }finally{root.recycle();}
    }
}
