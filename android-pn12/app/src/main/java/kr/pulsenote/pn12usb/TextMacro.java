package kr.pulsenote.pn12usb;

/** Expands a few explicit escapes in a user-programmed text macro.
 * No runtime expressions, device data, or network requests are evaluated. */
public final class TextMacro {
    private TextMacro() {}
    public static String expand(String raw) {
        if(raw==null || raw.isEmpty())return "";
        StringBuilder b=new StringBuilder(raw.length());
        for(int i=0;i<raw.length();i++) {
            char c=raw.charAt(i);
            if(c=='\\' && i+1<raw.length()) {
                char next=raw.charAt(i+1);
                if(next=='n'){b.append('\n');i++;continue;}
                if(next=='t'){b.append('\t');i++;continue;}
                if(next=='\\'){b.append('\\');i++;continue;}
            }
            if(c==0)return "";
            b.append(c);
        }
        return b.toString().replace("\r\n","\n");
    }
}
