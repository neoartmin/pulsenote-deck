package kr.pulsenote.pn12usb;

import android.content.Context;
import android.content.SharedPreferences;
import java.util.ArrayList;
import java.util.List;

public final class ActionStore {
    // Keep v1.0.1 numeric types 0..8 and its default-profile preference keys.
    public static final String[] NAMES = {
       "기능 없음", "볼륨 +", "볼륨 -", "음소거", "재생 / 정지",
       "다음 곡", "이전 곡", "웹사이트", "프로그램", "단축키", "폴더"
    };
    public static final String[] SYMBOLS = {"＋","◖","◗","×","▶","⏭","⏮","◎","▦","⌨","▣"};
    public static final int[] COLORS = {0xffa8c9e9,0xffdf9fa9,0xffe2e9f5,0xffc4aedc,0xffedbb95,0xffa8d4bf,0xffe4d99e};
    private static final int[] DEFAULT_TYPES = {7,7,4,7,7,0,9,9,0,6,5,10};
    private static final int[] DEFAULT_COLORS = {1,1,2,2,2,1,2,2,1,2,2,1};
    private static final String[] DEFAULT_NAMES = {
       "제미나이","챗GPT","버튼 03","한밭대 이러닝","유튜브","Autodesk …",
       "CTRL+C","CTRL+V","Bambu Studio","버튼 10","버튼 11","File Explorer"
    };
    private static final String[] DEFAULT_VALUES = {
       "https://gemini.google.com/","https://chatgpt.com/","","","https://www.youtube.com/","",
       "CTRL+C","CTRL+V","","","",""
    };
    private final SharedPreferences prefs;

    public ActionStore(Context context) {
        prefs=context.getSharedPreferences("pn12_actions",Context.MODE_PRIVATE);
    }
    public int activeProfile() { return prefs.getInt("active_profile",0); }
    public void setProfile(int id) { if(profiles().contains(id)) prefs.edit().putInt("active_profile",id).apply(); }
    public List<Integer> profiles() {
        ArrayList<Integer> result=new ArrayList<>(); result.add(0);
        String raw=prefs.getString("profile_ids","");
        for(String s:raw.split(",")) if(!s.isEmpty()) try{
            int id=Integer.parseInt(s); if(id>0 && !result.contains(id))result.add(id);
        }catch(NumberFormatException ignored){}
        return result;
    }
    public String profileName(int id) { return id==0 ? prefs.getString("profile_name_0","기본 프로필") : prefs.getString("profile_name_"+id,"프로필 "+id); }
    public void renameProfile(int id,String name) {
        String safe=name.trim();if(!safe.isEmpty() && profiles().contains(id))prefs.edit().putString("profile_name_"+id,safe).apply();
    }
    public int newProfile(String name) {
        int next=prefs.getInt("next_profile",1);List<Integer> list=profiles();list.add(next);
        StringBuilder all=new StringBuilder();for(int id:list)if(id>0){if(all.length()>0)all.append(',');all.append(id);}
        prefs.edit().putString("profile_ids",all.toString()).putString("profile_name_"+next,name.trim().isEmpty()?"새 프로필":name.trim()).putInt("active_profile",next).putInt("next_profile",next+1).apply();
        return next;
    }
    public boolean deleteCurrent() {
        int current=activeProfile();if(current==0)return false;
        StringBuilder all=new StringBuilder();for(int id:profiles())if(id>0&&id!=current){if(all.length()>0)all.append(',');all.append(id);}
        prefs.edit().putString("profile_ids",all.toString()).putInt("active_profile",0).apply();return true;
    }
    private String idKey(String field,int key) { return (activeProfile()==0?"":("p"+activeProfile()+"_"))+field+"_"+key; }
    private int ix(int key){return Math.max(0,Math.min(11,key-1));}
    public int kind(int key) {
       int k=prefs.getInt(idKey("type",key),activeProfile()==0?DEFAULT_TYPES[ix(key)]:0);
       return k>=0&&k<NAMES.length?k:0;
    }
    public String name(int key) {
       return prefs.getString(idKey("name",key),activeProfile()==0?DEFAULT_NAMES[ix(key)]:String.format(java.util.Locale.US,"버튼 %02d",key));
    }
    public String value(int key) {
       return prefs.getString(idKey("value",key),activeProfile()==0?DEFAULT_VALUES[ix(key)]:"");
    }
    public int color(int key) {
       int c=prefs.getInt(idKey("color",key),activeProfile()==0?DEFAULT_COLORS[ix(key)]:2);
       return c>=0&&c<COLORS.length?c:2;
    }
    public void setKind(int key,int kind) {
       if(key>=1&&key<=12&&kind>=0&&kind<NAMES.length)prefs.edit().putInt(idKey("type",key),kind).apply();
    }
    public void setName(int key,String name) {
       if(key>=1&&key<=12)prefs.edit().putString(idKey("name",key),name).apply();
    }
    public void setValue(int key,String value) {
       if(key>=1&&key<=12)prefs.edit().putString(idKey("value",key),value).apply();
    }
    public void setColor(int key,int color) {
       if(key>=1&&key<=12&&color>=0&&color<COLORS.length)prefs.edit().putInt(idKey("color",key),color).apply();
    }
    public String label(int key) { return name(key); }
}
