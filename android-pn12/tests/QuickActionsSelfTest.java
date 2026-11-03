package kr.pulsenote.pn12usb;

import java.util.List;

public final class QuickActionsSelfTest {
    private static void ok(boolean condition,String message){
        if(!condition)throw new AssertionError(message);
    }
    private static void fail(Runnable r){
        try{r.run();throw new AssertionError("expected input rejection");}
        catch(IllegalArgumentException expected){}
    }
    public static void main(String[] args){
        ok(QuickActions.WEB_SEARCH==16&&QuickActions.COMPOSE_EMAIL==17&&
            QuickActions.QR_CODE==18&&QuickActions.SEQUENCE==19&&
            QuickActions.SHARE_TEXT==20,"numeric kinds appended");
        ok(QuickActions.webSearch("google|저항과 회로").startsWith("https://www.google.com/search?q="),
            "Google encoded query");
        ok(QuickActions.webSearch("naver|날씨").startsWith("https://search.naver.com/"),
            "Naver query");
        ok(QuickActions.webSearch("youtube|회로 강의").contains("search_query="),
            "YouTube search");
        fail(()->QuickActions.webSearch("none|keyword"));
        fail(()->QuickActions.webSearch(""));
        QuickActions.EmailDraft email=QuickActions.email("user@example.com|제목|본문\n2번째 줄");
        ok("user@example.com".equals(email.to),"recipient");
        ok("제목".equals(email.subject),"subject");
        ok(email.body.endsWith("2번째 줄"),"multiline email body");
        fail(()->QuickActions.email("u@example.com\nBCC:bad|제목|본문"));
        fail(()->QuickActions.email(""));
        List<QuickActions.Step> steps=QuickActions.sequence("2,wait:250,4",15,1);
        ok(steps.size()==3&&steps.get(0).key==2,"three steps");
        ok(steps.get(1).isWait()&&steps.get(1).delayMs==250,"pause step");
        ok(steps.get(2).key==4,"last step");
        fail(()->QuickActions.sequence("1",15,1)); // No self recursion.
        fail(()->QuickActions.sequence("16",15,1)); // No out-of-range.
        fail(()->QuickActions.sequence("wait:1001",15,1));
        fail(()->QuickActions.sequence("2,3,4,5,6,7,8,9,10",15,1));
        System.out.println("QuickActionsSelfTest OK");
    }
}
