package kr.pulsenote.pn12usb;

import java.net.URLEncoder;
import java.nio.charset.StandardCharsets;
import java.util.ArrayList;
import java.util.List;

/** Shared, testable specs for additions not covered by existing URL/app shortcuts. */
public final class QuickActions {
    public static final int WEB_SEARCH=16;
    public static final int COMPOSE_EMAIL=17;
    public static final int QR_CODE=18;
    public static final int SEQUENCE=19;
    public static final int SHARE_TEXT=20;

    public static final int MAX_VALUE_LENGTH=1600;
    public static final int MAX_SEQUENCE_STEPS=8;
    private QuickActions(){}

    public static String webSearch(String raw){
        String value=raw==null?"":raw.trim();
        String engine="google";
        int pipe=value.indexOf('|');
        if(pipe>=0){
            engine=value.substring(0,pipe).trim().toLowerCase(java.util.Locale.ROOT);
            value=value.substring(pipe+1).trim();
        }
        if(value.isEmpty())throw new IllegalArgumentException("검색어를 입력하세요");
        if(value.length()>MAX_VALUE_LENGTH)throw new IllegalArgumentException("검색어가 너무 깁니다");
        String q=URLEncoder.encode(value,StandardCharsets.UTF_8);
        switch(engine){
            case "naver":return "https://search.naver.com/search.naver?query="+q;
            case "youtube":return "https://www.youtube.com/results?search_query="+q;
            case "google":return "https://www.google.com/search?q="+q;
            default:throw new IllegalArgumentException("검색 엔진은 google, naver, youtube만 지원합니다");
        }
    }
    public static String searchEngine(String raw){
        if(raw!=null&&raw.contains("|")){
            String id=raw.substring(0,raw.indexOf('|')).trim().toLowerCase(java.util.Locale.ROOT);
            if(id.equals("google")||id.equals("naver")||id.equals("youtube"))return id;
        }
        return "google";
    }
    public static String searchTerm(String raw){
        if(raw==null)return "";
        int sep=raw.indexOf('|');
        if(sep>=0&&raw.substring(0,sep).trim().equalsIgnoreCase(searchEngine(raw)))
            return raw.substring(sep+1);
        return raw;
    }
    public static String encodeSearch(String engine,String term){
        return engine+"|"+(term==null?"":term);
    }

    public static final class EmailDraft {
        public final String to,subject,body;
        EmailDraft(String a,String s,String b){to=a;subject=s;body=b;}
    }
    // Internal form storage protects pipes, newlines and emoji; old
    // "to|subject|body" values in existing profiles are still supported.
    public static String encodeEmail(String to,String subject,String body){
        return "mail2:"+java.net.URLEncoder.encode(to==null?"":to,StandardCharsets.UTF_8)
            +"|"+java.net.URLEncoder.encode(subject==null?"":subject,StandardCharsets.UTF_8)
            +"|"+java.net.URLEncoder.encode(body==null?"":body,StandardCharsets.UTF_8);
    }
    public static EmailDraft email(String raw){
        String[] data;
        if(raw!=null&&raw.startsWith("mail2:")){
            data=raw.substring(6).split("\\|",-1);
            if(data.length!=3)throw new IllegalArgumentException("이메일 데이터가 손상됐습니다");
            try{
                for(int i=0;i<data.length;i++)data[i]=java.net.URLDecoder.decode(
                    data[i],StandardCharsets.UTF_8);
            }catch(IllegalArgumentException error){
                throw new IllegalArgumentException("이메일 데이터 디코딩 실패",error);
            }
        }else data=(raw==null?"":raw).split("\\|",3);
        String recipient=data.length>0?data[0].trim():"";
        String subject=data.length>1?data[1]:"";
        String body=data.length>2?data[2]:"";
        if(raw==null||raw.trim().isEmpty())throw new IllegalArgumentException("받는 사람|제목|본문을 입력하세요");
        if(raw.length()>4096)throw new IllegalArgumentException("이메일 초안이 너무 깁니다");
        // Prevent URI/header injection through the recipient address.
        if(recipient.contains("\n")||recipient.contains("\r")||recipient.contains("?")||
           recipient.contains("#")||recipient.contains("&")||recipient.contains(" "))
            throw new IllegalArgumentException("이메일 주소 형식이 올바르지 않습니다");
        if(!recipient.isEmpty()&&(!recipient.contains("@")||recipient.contains(";")))
            throw new IllegalArgumentException("받는 사람은 단일 이메일 주소만 입력하세요");
        return new EmailDraft(recipient,subject,body);
    }

    public static final class Step {
        public final int key,delayMs;
        private Step(int k,int delay){key=k;delayMs=delay;}
        public boolean isWait(){return delayMs>=0;}
    }
    /**
     * Entries separated by commas or newlines, for example "2,wait:250,4".
     * Steps can reference other physical keys (in the SAME active profile)
     * but must never reference this sequence itself or another sequence.
     * No shell commands, hidden permissions, or recursive execution.
     */
    public static List<Step> sequence(String raw,int keys,int self){
        String[] tokens=(raw==null?"":raw).trim().split("[,\\n]+");
        List<Step> result=new ArrayList<>();
        if(raw==null||raw.trim().isEmpty())throw new IllegalArgumentException("실행할 키 번호를 입력하세요");
        if(tokens.length>MAX_SEQUENCE_STEPS)
            throw new IllegalArgumentException("연속 실행은 최대 "+MAX_SEQUENCE_STEPS+"단계입니다");
        for(String token:tokens){
            token=token.trim();
            if(token.isEmpty())throw new IllegalArgumentException("빈 단계가 있습니다");
            if(token.toLowerCase(java.util.Locale.ROOT).startsWith("wait:")){
                int pause=parseNumber(token.substring(5),"대기 시간");
                if(pause<0||pause>1000)
                    throw new IllegalArgumentException("대기는 0~1000ms만 허용합니다");
                result.add(new Step(-1,pause));
            }else{
                int key=parseNumber(token,"키 번호");
                if(key<1||key>keys||key==self)
                    throw new IllegalArgumentException("유효한 다른 키 번호(1~"+keys+")를 입력하세요");
                result.add(new Step(key,-1));
            }
        }
        return result;
    }
    private static int parseNumber(String raw,String label){
        try{return Integer.parseInt(raw.trim());}
        catch(NumberFormatException e){throw new IllegalArgumentException(label+"가 숫자가 아닙니다");}
    }
}
