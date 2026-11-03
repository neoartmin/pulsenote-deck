package kr.pulsenote.pn12usb;

public final class TextMacroSelfTest {
    private TextMacroSelfTest(){}
    private static void eq(String name,String actual,String expected){
        if(!expected.equals(actual))
            throw new AssertionError(name+" expected="+expected+" actual="+actual);
    }
    public static void main(String[] args){
        eq("Hangul and emoji",TextMacro.expand("안녕하세요 😀"),"안녕하세요 😀");
        eq("line break",TextMacro.expand("첫째\\n둘째"),"첫째\n둘째");
        eq("tab",TextMacro.expand("a\\tb"),"a\tb");
        eq("preserve path",TextMacro.expand("C:\\Users\\Documents"),"C:\\Users\\Documents");
        eq("literal slash",TextMacro.expand("a\\\\n"),"a\\n");
        eq("invalid null",TextMacro.expand("a\u0000b"),"");
        eq("empty",TextMacro.expand(""),"");
        System.out.println("TextMacroSelfTest OK");
    }
}
