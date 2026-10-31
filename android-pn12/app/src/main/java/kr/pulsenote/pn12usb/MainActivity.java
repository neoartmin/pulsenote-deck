package kr.pulsenote.pn12usb;

import android.Manifest;
import android.app.*;
import android.content.*;
import android.content.pm.PackageManager;
import android.content.pm.ResolveInfo;
import android.graphics.*;
import android.graphics.drawable.*;
import android.hardware.usb.*;
import android.os.*;
import android.provider.Settings;
import android.graphics.Color;
import android.text.*;
import android.view.*;
import android.widget.*;
import java.util.*;

public final class MainActivity extends Activity {
    public static volatile boolean visible=false;
    private static final String PERMISSION="kr.pulsenote.pn12usb.GRANT";
    private static final int FOLDER_PICK=41;
    private static final int BG=0xff080f1f, SIDE=0xff0f1c31, PANEL=0xff15243d,
      BORDER=0xff263e60, MUTED=0xffa3b8d8, WHITE=0xfff4f7ff,
      BLUE=0xff5769fa, ACTIVE=0xff283a78, INNER=0xff091526;
    private final List<UsbDevice> candidates=new ArrayList<>();
    private final View[] keyViews=new View[12];
    private final TextView[] nav=new TextView[3];
    private UsbManager usb;
    private ActionStore store;
    private BroadcastReceiver receiver;
    private LinearLayout pageHost,rightHost,keysHost,sideHost;
    private TextView statusHeader,statusSide,statusFooter,keyHeader,profileCaption,devicePickerLabel;
    private Spinner chooser;
    private int page=0,selectedKey=7;
    private boolean programmatic=false;

    @Override public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        store=new ActionStore(this);
        usb=(UsbManager)getSystemService(USB_SERVICE);
        getWindow().setStatusBarColor(BG);
        getWindow().setNavigationBarColor(BG);
        getWindow().setSoftInputMode(WindowManager.LayoutParams.SOFT_INPUT_ADJUST_RESIZE);
        buildShell();
        receiver=new BroadcastReceiver(){
           @Override public void onReceive(Context context,Intent intent){
              String a=intent.getAction();if(a==null)return;
              if(PERMISSION.equals(a)){
                 if(intent.getBooleanExtra(UsbManager.EXTRA_PERMISSION_GRANTED,false))startChosen();
                 else updateStatus("USB 접근 권한 거부");
              }else if(Pn12UsbService.STATUS.equals(a))updateStatus(intent.getStringExtra(Pn12UsbService.TEXT));
              else if(Pn12UsbService.KEY.equals(a)){
                 int number=intent.getIntExtra(Pn12UsbService.NUMBER,0);
                 String result=intent.getStringExtra(Pn12UsbService.TEXT);
                 if(statusFooter!=null)statusFooter.setText("KEY: "+number+" · "+result);
              }else if(UsbManager.ACTION_USB_DEVICE_ATTACHED.equals(a)||UsbManager.ACTION_USB_DEVICE_DETACHED.equals(a))scan();
           }
        };
        IntentFilter filter=new IntentFilter();
        for(String a:new String[]{PERMISSION,Pn12UsbService.STATUS,Pn12UsbService.KEY,UsbManager.ACTION_USB_DEVICE_ATTACHED,UsbManager.ACTION_USB_DEVICE_DETACHED})filter.addAction(a);
        if(Build.VERSION.SDK_INT>=33)registerReceiver(receiver,filter,Context.RECEIVER_NOT_EXPORTED);
        else registerReceiver(receiver,filter);
        scan();
        if(Build.VERSION.SDK_INT>=33&&checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS)!=PackageManager.PERMISSION_GRANTED)
            requestPermissions(new String[]{Manifest.permission.POST_NOTIFICATIONS},8);
    }

    private int dp(float n){return (int)(getResources().getDisplayMetrics().density*n+.5f);}
    private GradientDrawable shape(int fill,int radius,int stroke) {
        GradientDrawable d=new GradientDrawable();d.setColor(fill);d.setCornerRadius(dp(radius));
        if(stroke!=0)d.setStroke(dp(1),stroke);return d;
    }
    private GradientDrawable grad(int top,int bottom,int radius,int stroke){
        GradientDrawable d=new GradientDrawable(GradientDrawable.Orientation.TOP_BOTTOM,new int[]{top,bottom});
        d.setCornerRadius(dp(radius));if(stroke!=0)d.setStroke(dp(1),stroke);return d;
    }
    private void pad(View v,int x,int y){v.setPadding(dp(x),dp(y),dp(x),dp(y));}
    private TextView txt(String s,float size,int color,boolean bold){
        TextView t=new TextView(this);t.setText(s);t.setTextSize(size);t.setTextColor(color);
        t.setGravity(Gravity.CENTER_VERTICAL);t.setFontFeatureSettings("kern");
        if(bold)t.setTypeface(android.graphics.Typeface.DEFAULT,android.graphics.Typeface.BOLD);
        return t;
    }
    private LinearLayout vertical(){LinearLayout l=new LinearLayout(this);l.setOrientation(LinearLayout.VERTICAL);return l;}
    private LinearLayout horizontal(){LinearLayout l=new LinearLayout(this);l.setOrientation(LinearLayout.HORIZONTAL);l.setGravity(Gravity.CENTER_VERTICAL);return l;}
    private LinearLayout.LayoutParams lp(int w,int h){return new LinearLayout.LayoutParams(w<0?w:dp(w),h<0?h:dp(h));}
    private void add(LinearLayout parent,View view,int w,int h){parent.addView(view,lp(w,h));}
    private void space(LinearLayout parent,int px){add(parent,new View(this),1,px);}
    private LinearLayout card(){
        LinearLayout l=vertical();l.setBackground(grad(PANEL,0xff102037,16,BORDER));pad(l,16,15);return l;
    }
    private TextView pill(String s,boolean selected){
        TextView t=txt(s,12,WHITE,selected);t.setGravity(Gravity.CENTER);
        t.setBackground(shape(selected?ACTIVE:SIDE,10,selected?BLUE:BORDER));pad(t,12,9);return t;
    }
    private TextView click(String s,Runnable r){
        TextView t=pill(s,false);t.setOnClickListener(v->r.run());return t;
    }
    private void buildShell(){
        ScrollView vscroll=new ScrollView(this);vscroll.setFillViewport(true);vscroll.setBackgroundColor(BG);
        HorizontalScrollView horizontal=new HorizontalScrollView(this);
        horizontal.setFillViewport(true);horizontal.setHorizontalScrollBarEnabled(false);
        vscroll.addView(horizontal);
        LinearLayout shell=horizontal();shell.setGravity(Gravity.TOP);shell.setMinimumHeight(dp(690));horizontal.addView(shell);
        sideHost=vertical();sideHost.setBackground(grad(SIDE,0xff0b172b,0,0));pad(sideHost,14,24);
        add(shell,sideHost,216,-1);
        buildSidebar();
        LinearLayout content=vertical();content.setPadding(dp(18),dp(26),dp(20),dp(12));
        add(shell,content,1015,-2);
        LinearLayout top=horizontal();top.setGravity(Gravity.CENTER_VERTICAL);
        LinearLayout titleBox=vertical();
        TextView header=txt("버튼 설정",29,WHITE,true);add(titleBox,header,-1,48);
        add(titleBox,txt("화면에서 키를 선택하고 기능과 색상을 지정하세요.",13,MUTED,false),-1,25);
        top.addView(titleBox,new LinearLayout.LayoutParams(0,dp(79),1f));
        statusHeader=pill("●  USB 미연결",false);
        statusHeader.setOnClickListener(v->{page=2;refreshNavigation();renderPage();});
        add(top,statusHeader,174,42);space(top,12);
        add(top,pill("PN12 · 12키 전용",false),185,42);
        add(content,top,-1,87);
        pageHost=vertical();add(content,pageHost,-1,-2);
        space(content,10);
        LinearLayout footer=horizontal();pad(footer,12,5);
        footer.setBackground(shape(0xff11213a,9,BORDER));
        statusFooter=txt("●  하드웨어 연결 확인 대기",11,MUTED,false);
        footer.addView(statusFooter,new LinearLayout.LayoutParams(0,dp(26),1));
        TextView version=txt("v1.1.0  ·  ANDROID",10,BLUE,false);add(footer,version,-2,26);
        add(content,footer,-1,38);
        setContentView(vscroll);
        renderPage();
    }
    private void buildSidebar(){
        LinearLayout brand=horizontal();brand.setGravity(Gravity.CENTER_VERTICAL);
        View logo=new LogoView();add(brand,logo,53,53);space(brand,10);
        LinearLayout names=vertical();add(names,txt("PulseNote Deck",17,WHITE,true),-1,27);
        add(names,txt("CONTROL CENTER",10,MUTED,false),-1,18);add(brand,names,-2,-2);
        add(sideHost,brand,-1,65);space(sideHost,35);
        add(sideHost,txt("WORKSPACE",11,0xff83a0ca,false),-1,32);
        nav[0]=navItem("▦   버튼 설정",0);add(sideHost,nav[0],-1,53);space(sideHost,9);
        nav[1]=navItem("▣   프로필 관리",1);add(sideHost,nav[1],-1,53);
        space(sideHost,24);add(sideHost,txt("SYSTEM",11,0xff83a0ca,false),-1,30);
        nav[2]=navItem("⚙   장치 및 설정",2);add(sideHost,nav[2],-1,53);
        space(sideHost,140);
        statusSide=txt("●  USB 미연결\n\nPN12 연결 대기",12,WHITE,false);pad(statusSide,12,13);
        statusSide.setBackground(shape(0xff0d203a,12,BORDER));add(sideHost,statusSide,-1,92);
        space(sideHost,20);
        add(sideHost,txt("v1.1.0  ·  ANDROID",11,0xff829bc5,false),-1,25);
        refreshNavigation();
    }
    private TextView navItem(String name,int index){
        TextView t=txt(name,15,WHITE,true);t.setGravity(Gravity.CENTER_VERTICAL);
        pad(t,14,6);t.setOnClickListener(v->{page=index;refreshNavigation();renderPage();});
        return t;
    }
    private void refreshNavigation(){
        for(int i=0;i<nav.length;i++)if(nav[i]!=null)
            nav[i].setBackground(shape(page==i?ACTIVE:0xff0c182b,13,page==i?BLUE:BORDER));
    }
    private void renderPage(){
        if(pageHost==null)return;
        pageHost.removeAllViews();
        if(page==0)buildButtonPage();
        else if(page==1)buildProfilePage();
        else buildDevicePage();
    }
    private void buildButtonPage(){
        LinearLayout row=horizontal();row.setGravity(Gravity.TOP);
        LinearLayout center=card();add(row,center,515,-2);space(row,12);
        LinearLayout right=card();add(row,right,464,-2);
        add(pageHost,row,-1,-2);
        LinearLayout first=horizontal();first.addView(txt("나의 12키 컨트롤러",19,WHITE,true),new LinearLayout.LayoutParams(0,dp(32),1));
        add(first,pill("PN12 전용",true),145,37);add(center,first,-1,41);
        add(center,txt("가로 3개 × 세로 4개 · 키별 자동 저장",11,MUTED,false),-1,29);
        LinearLayout frame=vertical();frame.setGravity(Gravity.CENTER);
        frame.setBackground(shape(INNER,15,0xff2c4364));pad(frame,14,11);
        keysHost=vertical();keysHost.setGravity(Gravity.CENTER);
        add(frame,keysHost,-1,-2);
        add(center,frame,-1,-2);
        renderKeys();
        space(center,13);
        LinearLayout bottom=horizontal();
        profileCaption=txt("프로필 · "+store.profileName(store.activeProfile()),11,MUTED,false);
        bottom.addView(profileCaption,new LinearLayout.LayoutParams(0,dp(40),1));
        add(bottom,click("▷   동작 테스트",()->testSelected()),150,39);
        add(center,bottom,-1,42);
        rightHost=right;
        renderRight();
    }
    private void renderKeys(){
        if(keysHost==null || page!=0)return;
        keysHost.removeAllViews();
        for(int r=0;r<4;r++){
            LinearLayout row=horizontal();row.setGravity(Gravity.CENTER);
            for(int c=0;c<3;c++){
                int k=r*3+c+1;
                LinearLayout key=vertical();key.setGravity(Gravity.CENTER_HORIZONTAL);pad(key,7,4);
                int color=ActionStore.COLORS[store.color(k)];
                key.setBackground(grad(adjust(color,0.19f),color,14,k==selectedKey?BLUE:0xffa3bdd7));
                TextView n=txt(String.format(Locale.US,"%02d",k),11,0xff202e45,true);
                n.setGravity(Gravity.LEFT);add(key,n,-1,17);
                View icon=createSymbol(store.kind(k),39,k);
                add(key,icon,39,39);
                TextView name=txt(store.name(k),11,0xff162133,true);
                name.setGravity(Gravity.CENTER);name.setSingleLine(true);name.setEllipsize(android.text.TextUtils.TruncateAt.END);
                add(key,name,-1,21);
                TextView type=txt(ActionStore.NAMES[store.kind(k)],9,0xff41536b,false);
                type.setGravity(Gravity.CENTER);add(key,type,-1,15);
                key.setOnClickListener(v->{selectedKey=k;renderKeys();renderRight();});
                LinearLayout.LayoutParams kp=lp(120,109);kp.setMargins(dp(5),dp(4),dp(5),dp(4));row.addView(key,kp);keyViews[k-1]=key;
            }
            add(keysHost,row,-1,-2);
        }
    }
    private int adjust(int color,float factor){
        return Color.rgb(Math.min(255,(int)(Color.red(color)+(255-Color.red(color))*factor)),
            Math.min(255,(int)(Color.green(color)+(255-Color.green(color))*factor)),
            Math.min(255,(int)(Color.blue(color)+(255-Color.blue(color))*factor)));
    }
    private View createSymbol(int kind,int size,int key){
        return KeyArtwork.create(this,kind,store.value(key),size);
    }
    private void renderRight(){
        if(rightHost==null||page!=0)return;
        rightHost.removeAllViews();
        int k=selectedKey,kind=store.kind(k);
        add(rightHost,txt(String.format(Locale.US,"키 %02d 설정",k),19,WHITE,true),-1,32);
        add(rightHost,txt("기능·이름·색상을 원하는 대로 지정하세요.",11,MUTED,false),-1,27);
        LinearLayout preview=horizontal();pad(preview,13,10);preview.setBackground(shape(0xff0d1b32,11,BORDER));
        add(preview,createSymbol(kind,44,k),44,44);space(preview,11);
        LinearLayout text=vertical();add(text,txt(store.name(k),14,WHITE,true),-1,23);
        add(text,txt(ActionStore.NAMES[kind],10,MUTED,false),-1,19);add(preview,text,-1,-1);
        add(rightHost,preview,-1,64);space(rightHost,13);
        add(rightHost,txt("기능",13,WHITE,true),-1,27);
        int[] types={9,8,7,10,1,2,3,4,5,6,0};
        for(int r=0;r<4;r++){
            LinearLayout line=horizontal();
            for(int col=0;col<3;col++){
                int pos=r*3+col;
                if(pos>=types.length)break;
                int type=types[pos];
                TextView b=txt(ActionStore.SYMBOLS[type]+"  "+ActionStore.NAMES[type],10,WHITE,type==kind);
                b.setGravity(Gravity.CENTER);b.setSingleLine(true);
                b.setBackground(shape(type==kind?ACTIVE:0xff0e1d34,10,type==kind?BLUE:BORDER));
                b.setOnClickListener(v->{
                    store.setKind(k,type);
                    renderKeys();renderRight();
                });
                LinearLayout.LayoutParams pp=new LinearLayout.LayoutParams(0,dp(39),1);pp.setMargins(dp(3),dp(3),dp(3),dp(3));
                line.addView(b,pp);
            }
            add(rightHost,line,-1,45);
        }
        space(rightHost,6);
        add(rightHost,txt("버튼 이름",12,WHITE,true),-1,27);
        EditText name=edit(store.name(k),"이름 입력");add(rightHost,name,-1,39);
        name.addTextChangedListener(watcher(s->{
            if(programmatic)return;store.setName(k,s);renderKeys();
        }));
        String hint=null;
        if(kind==9)hint="키보드 단축키 (예: CTRL+C)";
        else if(kind==7)hint="웹사이트 주소 (https://...)";
        else if(kind==8)hint="Android 앱 패키지";
        else if(kind==10)hint="저장된 폴더 주소 (URI)";
        if(hint!=null){
            space(rightHost,8);add(rightHost,txt(hint,12,WHITE,true),-1,27);
            EditText value=edit(store.value(k),hint);add(rightHost,value,-1,39);
            value.addTextChangedListener(watcher(s->store.setValue(k,s)));
            if(kind==8){
                space(rightHost,5);add(rightHost,click("▦  설치된 앱에서 선택",()->pickApp(k)),208,35);
            }
            if(kind==10){
                space(rightHost,5);add(rightHost,click("▣  Android 폴더 선택",()->pickFolder()),208,35);
            }
            if(kind==9){
                add(rightHost,txt("복사·붙여넣기·잘라내기 등은 접근성 권한 필요. 모든 Windows 단축키를 지원하지는 않습니다.",10,MUTED,false),-1,43);
                add(rightHost,click("접근성 단축키 설정 열기",()->startActivity(new Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS))),220,34);
            }
        }
        space(rightHost,9);
        add(rightHost,txt("키캡 색상",12,WHITE,true),-1,27);
        LinearLayout swatches=horizontal();
        for(int c=0;c<ActionStore.COLORS.length;c++){
            final int colorIndex=c;
            TextView swatch=txt(store.color(k)==c?"✓":"",15,0xff18233a,true);swatch.setGravity(Gravity.CENTER);
            swatch.setBackground(shape(ActionStore.COLORS[c],9,store.color(k)==c?BLUE:0xff394965));
            swatch.setOnClickListener(v->{store.setColor(k,colorIndex);renderKeys();renderRight();});
            LinearLayout.LayoutParams w=new LinearLayout.LayoutParams(0,dp(30),1);w.setMargins(dp(3),0,dp(3),0);swatches.addView(swatch,w);
        }
        add(rightHost,swatches,-1,35);
        space(rightHost,8);
        if(kind==8) add(rightHost,txt("참고: Windows .exe 실행은 Android에서 지원되지 않습니다.",10,MUTED,false),-1,26);
        if(kind==10) add(rightHost,txt("폴더는 Android 문서 선택기를 통해 엽니다.",10,MUTED,false),-1,24);
    }
    private EditText edit(String value,String hint){
        EditText e=new EditText(this);e.setText(value);e.setSingleLine(true);e.setTextSize(12);
        e.setSelectAllOnFocus(false);e.setTextColor(WHITE);e.setHintTextColor(MUTED);e.setHint(hint);
        e.setBackground(shape(0xff0b182d,5,0xff475979));pad(e,8,2);
        return e;
    }
    private interface TextChanged{void changed(String value);}
    private TextWatcher watcher(TextChanged listener){
        return new TextWatcher(){public void beforeTextChanged(CharSequence s,int start,int count,int after){}
          public void onTextChanged(CharSequence s,int start,int before,int count){listener.changed(s.toString());}
          public void afterTextChanged(Editable e){}};
    }
    private void testSelected(){
        try{String r=ActionDispatcher.run(this,selectedKey);if(statusFooter!=null)statusFooter.setText("KEY: "+selectedKey+" · "+r);}
        catch(Exception ex){if(statusFooter!=null)statusFooter.setText("테스트 오류: "+ex.getMessage());}
    }
    private void buildProfilePage(){
        LinearLayout panel=card();add(pageHost,panel,-1,-2);
        add(panel,txt("프로필 관리",22,WHITE,true),-1,43);
        add(panel,txt("프로필마다 12개 버튼의 기능, 이름, 색상을 별도로 저장합니다.",13,MUTED,false),-1,40);
        for(int id:store.profiles()){
            LinearLayout line=horizontal();pad(line,12,9);
            line.setBackground(shape(id==store.activeProfile()?ACTIVE:INNER,12,id==store.activeProfile()?BLUE:BORDER));
            TextView name=txt((id==store.activeProfile()?"●  ":"○  ")+store.profileName(id),15,WHITE,true);
            line.addView(name,new LinearLayout.LayoutParams(0,dp(50),1));
            add(line,click("적용",()->{store.setProfile(id);selectedKey=7;renderPage();}),94,37);
            LinearLayout.LayoutParams p=lp(-1,68);p.bottomMargin=dp(11);panel.addView(line,p);
        }
        add(panel,click("+  새 프로필 만들기",()->profileDialog(false)),213,43);space(panel,9);
        add(panel,click("선택한 프로필 이름 변경",()->profileDialog(true)),213,42);space(panel,9);
        add(panel,click("선택한 프로필 삭제",()->new AlertDialog.Builder(this).setTitle("프로필 삭제")
            .setMessage("현재 프로필을 목록에서 삭제하시겠습니까? 기본 프로필은 삭제할 수 없습니다.")
            .setNegativeButton("취소",null).setPositiveButton("삭제",(d,w)->{store.deleteCurrent();renderPage();}).show()),213,42);
        add(panel,txt("기본 프로필은 삭제되지 않습니다. 다른 프로필은 별도 저장되며 Windows PC 프로그램의 프로필 파일에는 영향을 주지 않습니다.",11,MUTED,false),-1,66);
    }
    private void profileDialog(boolean rename){
        EditText input=edit(rename?store.profileName(store.activeProfile()):"","프로필 이름");
        LinearLayout wrapper=vertical();pad(wrapper,20,14);add(wrapper,input,-1,48);
        new AlertDialog.Builder(this).setTitle(rename?"프로필 이름 변경":"새 프로필")
          .setView(wrapper).setNegativeButton("취소",null)
          .setPositiveButton("저장",(d,w)->{
             String name=input.getText().toString().trim();
             if(name.isEmpty())return;
             if(rename)store.renameProfile(store.activeProfile(),name);else store.newProfile(name);
             renderPage();
          }).show();
    }
    private void buildDevicePage(){
        LinearLayout card=card();add(pageHost,card,-1,-2);
        add(card,txt("장치 및 설정",23,WHITE,true),-1,50);
        add(card,txt("PulseNote PN12 · ESP32-C3 · USB CDC 115200bps",14,MUTED,false),-1,38);
        devicePickerLabel=txt("USB-C 데이터 케이블을 연결한 후 장치를 검색하세요.",12,MUTED,false);
        add(card,devicePickerLabel,-1,48);
        chooser=new Spinner(this);chooser.setBackground(shape(0xff0d1b33,11,BORDER));add(card,chooser,650,53);
        space(card,14);
        LinearLayout row=horizontal();
        add(row,click("⟳  USB 장치 검색",()->scan()),180,48);space(row,12);
        add(row,click("●  PN12 USB 연결",()->connect()),200,48);space(row,12);
        add(row,click("■  연결 종료",()->startService(new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.STOP))),170,48);
        add(card,row,-1,55);space(card,15);
        add(card,txt("현재 상태",14,WHITE,true),-1,29);
        TextView state=txt(Pn12UsbService.status,14,0xff3edba8,true);add(card,state,-1,41);
        add(card,txt("연결 후 상태는 상단의 USB 표시에도 반영됩니다. Android에서는 COM 포트 번호 대신 USB Host 장치 경로를 사용합니다.",12,MUTED,false),-1,55);
        add(card,click("Android 접근성 단축키 설정",()->startActivity(new Intent(Settings.ACTION_ACCESSIBILITY_SETTINGS))),300,46);
        space(card,14);
        add(card,txt("Windows용 기존 PN12 펌웨어는 그대로 사용합니다. 단축키는 Android 보안 정책과 접근성 서비스 지원 범위에 따라 제한될 수 있으며 앱 실행은 백그라운드 상태에서 알림을 통해 진행됩니다.",12,MUTED,false),-1,84);
        scan();
    }
    private void pickApp(int key){
        Intent query=new Intent(Intent.ACTION_MAIN);query.addCategory(Intent.CATEGORY_LAUNCHER);
        List<ResolveInfo> info=getPackageManager().queryIntentActivities(query,0);
        info.sort((a,b)->String.CASE_INSENSITIVE_ORDER.compare(String.valueOf(a.loadLabel(getPackageManager())),String.valueOf(b.loadLabel(getPackageManager()))));
        if(info.isEmpty()){new AlertDialog.Builder(this).setMessage("선택 가능한 앱이 없습니다. 패키지 이름을 직접 입력하세요.").setPositiveButton("확인",null).show();return;}
        String[] names=new String[info.size()];
        for(int i=0;i<info.size();i++)names[i]=String.valueOf(info.get(i).loadLabel(getPackageManager()));
        new AlertDialog.Builder(this).setTitle("실행할 앱 선택").setItems(names,(d,which)->{
            store.setValue(key,info.get(which).activityInfo.packageName);
            store.setName(key,names[which]);renderKeys();renderRight();
        }).setNegativeButton("취소",null).show();
    }
    private void pickFolder(){
        try{Intent intent=new Intent(Intent.ACTION_OPEN_DOCUMENT_TREE);
            intent.addFlags(Intent.FLAG_GRANT_READ_URI_PERMISSION|Intent.FLAG_GRANT_PERSISTABLE_URI_PERMISSION);
            startActivityForResult(intent,FOLDER_PICK);
        }catch(Exception e){updateStatus("폴더 선택기 실행 실패");}
    }
    @Override protected void onActivityResult(int request,int result,Intent data){
        super.onActivityResult(request,result,data);
        if(request==FOLDER_PICK&&result==RESULT_OK&&data!=null&&data.getData()!=null){
            try{getContentResolver().takePersistableUriPermission(data.getData(),Intent.FLAG_GRANT_READ_URI_PERMISSION);}
            catch(Exception ignored){}
            store.setValue(selectedKey,data.getData().toString());renderRight();
        }
    }
    private void scan(){
        if(usb==null)return;candidates.clear();
        ArrayList<String> labels=new ArrayList<>();
        for(UsbDevice d:usb.getDeviceList().values())if(Pn12UsbService.candidate(d)){
            candidates.add(d);labels.add("USB CDC  ·  VID "+d.getVendorId()+" / PID "+d.getProductId());
        }
        if(labels.isEmpty())labels.add("PN12 USB CDC 장치 없음");
        if(chooser!=null)chooser.setAdapter(new ArrayAdapter<>(this,android.R.layout.simple_spinner_dropdown_item,labels));
        if(devicePickerLabel!=null)devicePickerLabel.setText(candidates.isEmpty()?"장치를 찾지 못했습니다. USB-C 데이터 케이블과 USB Host 지원을 확인하세요.":"호환되는 USB CDC 장치 "+candidates.size()+"개 발견");
    }
    private UsbDevice chosen(){return candidates.isEmpty()?null:candidates.get(Math.max(0,Math.min(candidates.size()-1,chooser==null?0:chooser.getSelectedItemPosition())));}
    private void connect(){
        scan();UsbDevice d=chosen();
        if(d==null){updateStatus("USB CDC 장치 없음");return;}
        if(usb.hasPermission(d))startChosen();
        else {
            PendingIntent p=PendingIntent.getBroadcast(this,1,new Intent(PERMISSION).setPackage(getPackageName()),PendingIntent.FLAG_MUTABLE|PendingIntent.FLAG_UPDATE_CURRENT);
            usb.requestPermission(d,p);
        }
    }
    private void startChosen(){
        UsbDevice d=chosen();if(d==null)return;
        Intent intent=new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.START);
        intent.putExtra(Pn12UsbService.DEVICE,d.getDeviceName());
        if(Build.VERSION.SDK_INT>=26)startForegroundService(intent);else startService(intent);
        updateStatus("USB 연결 중");
    }
    private void updateStatus(String message){
        if(message==null)message="상태 확인 중";
        boolean okay=message.startsWith("연결 완료");
        if(statusHeader!=null)statusHeader.setText(okay?"●  USB 연결됨":"●  "+message);
        if(statusSide!=null)statusSide.setText((okay?"●  USB 연결됨":"●  USB 상태")+"\n\n"+message);
        if(statusFooter!=null)statusFooter.setText("●  "+message);
        if(page==2&&devicePickerLabel!=null)devicePickerLabel.setText(message);
    }
    @Override protected void onResume(){super.onResume();visible=true;updateStatus(Pn12UsbService.status);}
    @Override protected void onPause(){visible=false;super.onPause();}
    @Override protected void onDestroy(){if(receiver!=null)unregisterReceiver(receiver);super.onDestroy();}
    private final class LogoView extends View{
        private final Paint paint=new Paint(3);
        LogoView(){super(MainActivity.this);}
        @Override protected void onDraw(Canvas canvas){
           float w=getWidth(),h=getHeight(),r=dp(13);
           paint.setColor(0xff566afa);canvas.drawRoundRect(0,0,w,h,r,r,paint);
           paint.setColor(WHITE);float s=w*.17f;
           for(int y=0;y<2;y++)for(int x=0;x<2;x++){
               float px=w*(.31f+x*.30f),py=h*(.31f+y*.30f);
               canvas.drawRoundRect(px,py,px+s,py+s,s*.25f,s*.25f,paint);
           }
        }
    }
}
