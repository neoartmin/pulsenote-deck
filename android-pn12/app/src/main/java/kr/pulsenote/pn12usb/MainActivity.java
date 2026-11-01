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
import android.net.Uri;
import android.graphics.Color;
import android.text.*;
import android.view.*;
import android.widget.*;
import java.util.*;

public final class MainActivity extends Activity {
    public static volatile boolean visible=false;
    private static final String PERMISSION="kr.pulsenote.pn12usb.GRANT";
    private static final int FOLDER_PICK=41;
    private static final int NOTIFICATION_REQUEST=8;
    private static final int BG=0xff080f1f, SIDE=0xff0f1c31, PANEL=0xff15243d,
      BORDER=0xff263e60, MUTED=0xffa3b8d8, WHITE=0xfff4f7ff,
      BLUE=0xff5769fa, ACTIVE=0xff283a78, INNER=0xff091526;
    private final List<UsbDevice> candidates=new ArrayList<>();
    private final View[] keyViews=new View[15];
    private final TextView[] nav=new TextView[3];
    private UsbManager usb;
    private ActionStore store;
    private BroadcastReceiver receiver;
    private LinearLayout pageHost,rightHost,keysHost,sideHost;
    private TextView statusHeader,statusSide,statusFooter,keyHeader,profileCaption,devicePickerLabel;
    private Spinner chooser;
    private int page=0,selectedKey=7,activeKeyCount=15;
    private String activeSide="R";
    private boolean programmatic=false;
    // All measurements are Android density-independent pixels (dp).
    // Recomputed after rotation, display scaling, or split-screen resize.
    private int screenWidthDp=800,contentWidthDp=620,panelWidthDp=620;
    private int centerWidthDp=515,rightWidthDp=445;
    private boolean compactNav=false,splitPanels=false;
    private String pendingUsbDevice=null;
    private boolean permissionPromptActive=false;
    private boolean showingConnectionHome=false;

    @Override public void onCreate(Bundle savedInstanceState) {
        super.onCreate(savedInstanceState);
        if(savedInstanceState!=null){
            page=Math.max(0,Math.min(2,savedInstanceState.getInt("pn12.page",0)));
            selectedKey=Math.max(1,Math.min(15,savedInstanceState.getInt("pn12.selectedKey",7)));
        }
        store=new ActionStore(this);
        activeKeyCount=store.keyCount();
        activeSide=store.deviceSide();
        selectedKey=Math.min(selectedKey,activeKeyCount);
        usb=(UsbManager)getSystemService(USB_SERVICE);
        getWindow().setStatusBarColor(BG);
        getWindow().setNavigationBarColor(BG);
        getWindow().setSoftInputMode(WindowManager.LayoutParams.SOFT_INPUT_ADJUST_RESIZE);
        buildShell();
        receiver=new BroadcastReceiver(){
           @Override public void onReceive(Context context,Intent intent){
              String a=intent.getAction();if(a==null)return;
              if(PERMISSION.equals(a)){
                 permissionPromptActive=false;
                 UsbDevice device=usbFromIntent(intent);
                 if(device==null&&pendingUsbDevice!=null)device=usb.getDeviceList().get(pendingUsbDevice);
                 pendingUsbDevice=null;
                 if(intent.getBooleanExtra(UsbManager.EXTRA_PERMISSION_GRANTED,false)&&device!=null)startDevice(device);
                 else updateStatus("USB 접근 권한 거부");
                  new Handler(Looper.getMainLooper()).post(()->maybeAutoConnect());
              }else if(Pn12UsbService.STATUS.equals(a)){
                 String status=intent.getStringExtra(Pn12UsbService.TEXT);
                 updateStatus(status);
                  int reportedModel=intent.getIntExtra(Pn12UsbService.MODEL,0);
                  if(reportedModel==12||reportedModel==15)
                      switchDetectedModel(reportedModel,intent.getStringExtra(Pn12UsbService.SIDE));
                 if(showingConnectionHome&&status!=null&&status.startsWith("연결 완료")){
                     showingConnectionHome=false;
                     showControllerHome();
                 }
              }
              else if(Pn12UsbService.KEY.equals(a)){
                 int number=intent.getIntExtra(Pn12UsbService.NUMBER,0);
                 String result=intent.getStringExtra(Pn12UsbService.TEXT);
                  int eventModel=intent.getIntExtra(Pn12UsbService.MODEL,0);
                  if(eventModel==12||eventModel==15)
                      switchDetectedModel(eventModel,intent.getStringExtra(Pn12UsbService.SIDE));
                 if(statusFooter!=null)statusFooter.setText("KEY: "+number+" · "+result);
              }else if(UsbManager.ACTION_USB_DEVICE_ATTACHED.equals(a)){
                  showControllerHome();
                  scan();
                  UsbDevice newlyAttached=usbFromIntent(intent);
                  if(newlyAttached!=null)maybeAutoConnect(newlyAttached);
                  else maybeAutoConnect();
              }else if(UsbManager.ACTION_USB_DEVICE_DETACHED.equals(a)){
                  permissionPromptActive=false;pendingUsbDevice=null;showingConnectionHome=false;
                  scan();
              }
           }
        };
        IntentFilter filter=new IntentFilter();
        for(String a:new String[]{PERMISSION,Pn12UsbService.STATUS,Pn12UsbService.KEY,UsbManager.ACTION_USB_DEVICE_ATTACHED,UsbManager.ACTION_USB_DEVICE_DETACHED})filter.addAction(a);
        if(Build.VERSION.SDK_INT>=33)registerReceiver(receiver,filter,Context.RECEIVER_NOT_EXPORTED);
        else registerReceiver(receiver,filter);
        scan();
        UsbDevice attached=usbFromIntent(getIntent());
        if(attached!=null)maybeAutoConnect(attached);
        else maybeAutoConnect();
        if(Build.VERSION.SDK_INT>=33&&checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS)!=PackageManager.PERMISSION_GRANTED)
            requestPermissions(new String[]{Manifest.permission.POST_NOTIFICATIONS},NOTIFICATION_REQUEST);
        boolean shown=getSharedPreferences("pn12_permissions",MODE_PRIVATE).getBoolean("permission_guide_111",false);
        if(!shown){
            getSharedPreferences("pn12_permissions",MODE_PRIVATE).edit().putBoolean("permission_guide_111",true).apply();
            new Handler(Looper.getMainLooper()).postDelayed(()->{
                if(isFinishing()||isDestroyed())return;
                new AlertDialog.Builder(this).setTitle("PulseNote USB 사용 권한 안내")
                    .setMessage("USB 연결 권한은 PN12/PN15 연결 시 요청합니다. 앱을 벗어나서 다른 앱을 직접 실행하려면 '다른 앱 위에 표시' 허용과 직접 실행 모드를 켜세요. 접근성 서비스는 설치 차단을 줄이기 위해 제외했습니다. 일부 Android 버전에서는 백그라운드 직접 실행이 제한됩니다.")
                    .setNegativeButton("나중에",null)
                    .setPositiveButton("권한 설정", (d,w)->{page=2;refreshNavigation();renderPage();})
                    .show();
            },1100);
        }
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
    private int screenWidth() {
        int reported=getResources().getConfiguration().screenWidthDp;
        if(reported>0)return reported;
        return (int)(getResources().getDisplayMetrics().widthPixels/getResources().getDisplayMetrics().density);
    }
    private void measureLayout(){
        screenWidthDp=screenWidth();
        compactNav=screenWidthDp<700;
        splitPanels=screenWidthDp>=1040;
        int sideWidth=compactNav?0:(screenWidthDp>=1050?196:168);
        int shellPadding=compactNav?20:28;
        contentWidthDp=Math.max(280,screenWidthDp-sideWidth-shellPadding);
        int rowWidth=contentWidthDp-12;
        if(splitPanels){
            centerWidthDp=Math.round(rowWidth*0.525f);
            rightWidthDp=rowWidth-centerWidthDp;
            panelWidthDp=centerWidthDp;
        }else{
            centerWidthDp=contentWidthDp;
            rightWidthDp=contentWidthDp;
            panelWidthDp=contentWidthDp;
        }
    }
    private void buildShell(){
        measureLayout();
        ScrollView scroller=new ScrollView(this);
        scroller.setFillViewport(true);
        scroller.setVerticalScrollBarEnabled(true);
        scroller.setBackgroundColor(BG);
        LinearLayout container=vertical();
        container.setBackgroundColor(BG);
        scroller.addView(container,new ScrollView.LayoutParams(-1,-2));
        LinearLayout shell=horizontal();
        shell.setGravity(Gravity.TOP);
        if(compactNav){
            buildCompactNavigation(container);
            add(container,shell,-1,-2);
        }else{
            add(container,shell,-1,-2);
            sideHost=vertical();
            sideHost.setBackground(grad(SIDE,0xff0b172b,0,0));
            pad(sideHost,12,18);
            add(shell,sideHost,screenWidthDp>=1050?196:168,-1);
            buildSidebar();
        }
        LinearLayout content=vertical();
        content.setPadding(dp(compactNav?10:14),dp(12),dp(compactNav?10:14),dp(10));
        shell.addView(content,new LinearLayout.LayoutParams(0,-2,1));
        if(contentWidthDp<770){
            LinearLayout headings=vertical();
            add(headings,txt("버튼 설정",compactNav?23:26,WHITE,true),-1,41);
            TextView detail=txt("화면에서 키를 선택하고 기능과 색상을 지정하세요.",12,MUTED,false);
            add(headings,detail,-1,-2);
            add(content,headings,-1,-2);
            space(content,8);
            LinearLayout statuses=horizontal();
            statusHeader=pill("●  USB 미연결",false);
            statusHeader.setOnClickListener(v->{page=2;refreshNavigation();renderPage();});
            statusHeader.setSingleLine(true);
            statusHeader.setEllipsize(android.text.TextUtils.TruncateAt.END);
            LinearLayout.LayoutParams st=new LinearLayout.LayoutParams(0,dp(39),1);
            statuses.addView(statusHeader,st);
            space(statuses,8);
            TextView device=pill("PN12 / PN15 · USB",false);
            device.setSingleLine(true);
            statuses.addView(device,new LinearLayout.LayoutParams(0,dp(39),1));
            add(content,statuses,-1,45);
        }else{
            LinearLayout top=horizontal();
            top.setGravity(Gravity.CENTER_VERTICAL);
            LinearLayout titleBox=vertical();
            add(titleBox,txt("버튼 설정",29,WHITE,true),-1,46);
            add(titleBox,txt("화면에서 키를 선택하고 기능과 색상을 지정하세요.",13,MUTED,false),-1,24);
            top.addView(titleBox,new LinearLayout.LayoutParams(0,dp(75),1));
            statusHeader=pill("●  USB 미연결",false);
            statusHeader.setSingleLine(true);
            statusHeader.setEllipsize(android.text.TextUtils.TruncateAt.END);
            statusHeader.setOnClickListener(v->{page=2;refreshNavigation();renderPage();});
            add(top,statusHeader,163,39);
            space(top,9);
            add(top,pill("PN12 / PN15 · USB",false),161,39);
            add(content,top,-1,82);
        }
        space(content,9);
        pageHost=vertical();
        add(content,pageHost,-1,-2);
        space(content,10);
        LinearLayout footer=horizontal();
        pad(footer,10,5);
        footer.setBackground(shape(0xff11213a,9,BORDER));
        statusFooter=txt("●  하드웨어 연결 확인 대기",11,MUTED,false);
        statusFooter.setSingleLine(true);
        statusFooter.setEllipsize(android.text.TextUtils.TruncateAt.END);
        footer.addView(statusFooter,new LinearLayout.LayoutParams(0,dp(26),1));
        add(footer,txt("v1.3.0 · ANDROID",10,BLUE,false),-2,26);
        add(content,footer,-1,37);
        setContentView(scroller);
        renderPage();
        updateStatus(Pn12UsbService.status);
    }
    private void buildCompactNavigation(LinearLayout parent){
        LinearLayout navRoot=vertical();
        navRoot.setBackground(grad(SIDE,0xff0b172b,0,0));
        navRoot.setPadding(dp(10),dp(10),dp(10),dp(9));
        LinearLayout brand=horizontal();
        add(brand,new LogoView(),37,37);
        space(brand,8);
        brand.addView(txt("PulseNote Deck",17,WHITE,true),new LinearLayout.LayoutParams(0,dp(39),1));
        add(brand,txt("PN12·15",11,MUTED,true),-2,30);
        add(navRoot,brand,-1,40);
        space(navRoot,10);
        LinearLayout navRow=horizontal();
        String[] labels={"▦  버튼 설정","▣  프로필","⚙  장치"};
        for(int i=0;i<3;i++){
            final int index=i;
            nav[i]=navItem(labels[i],i);
            nav[i].setGravity(Gravity.CENTER);
            nav[i].setTextSize(11);
            LinearLayout.LayoutParams p=new LinearLayout.LayoutParams(0,dp(42),1);
            p.setMargins(dp(2),0,dp(2),0);
            navRow.addView(nav[i],p);
        }
        add(navRoot,navRow,-1,44);
        add(parent,navRoot,-1,-2);
        statusSide=null;
        refreshNavigation();
    }
    private void buildSidebar(){
        LinearLayout brand=horizontal();brand.setGravity(Gravity.CENTER_VERTICAL);
        add(brand,new LogoView(),43,43);space(brand,8);
        LinearLayout names=vertical();
        TextView name=txt("PulseNote Deck",15,WHITE,true);
        name.setSingleLine(true);
        name.setEllipsize(android.text.TextUtils.TruncateAt.END);
        add(names,name,-1,23);
        add(names,txt("CONTROL CENTER",9,MUTED,false),-1,16);
        brand.addView(names,new LinearLayout.LayoutParams(0,-2,1));
        add(sideHost,brand,-1,62);
        space(sideHost,27);
        add(sideHost,txt("WORKSPACE",11,0xff83a0ca,false),-1,28);
        nav[0]=navItem("▦  버튼 설정",0);add(sideHost,nav[0],-1,50);space(sideHost,7);
        nav[1]=navItem("▣  프로필 관리",1);add(sideHost,nav[1],-1,50);
        space(sideHost,24);
        add(sideHost,txt("SYSTEM",11,0xff83a0ca,false),-1,28);
        nav[2]=navItem("⚙  장치 및 설정",2);add(sideHost,nav[2],-1,50);
        space(sideHost,Math.max(20,getResources().getConfiguration().screenHeightDp-570));
        statusSide=txt("●  USB 미연결\n\nPN12 연결 대기",11,WHITE,false);
        pad(statusSide,8,10);
        statusSide.setBackground(shape(0xff0d203a,12,BORDER));
        add(sideHost,statusSide,-1,85);
        space(sideHost,12);
        add(sideHost,txt("v1.2.2 · ANDROID",10,0xff829bc5,false),-1,23);
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
        // Use a side-by-side editor only when both panes fit without horizontal scrolling.
        LinearLayout wrapper=splitPanels?horizontal():vertical();
        wrapper.setGravity(Gravity.TOP);
        LinearLayout center=card();
        if(splitPanels){
            add(wrapper,center,centerWidthDp,-2);
            View gap=new View(this);
            add(wrapper,gap,12,1);
        }else{
            add(wrapper,center,-1,-2);
            space(wrapper,12);
        }
        LinearLayout right=card();
        if(splitPanels)add(wrapper,right,rightWidthDp,-2);
        else add(wrapper,right,-1,-2);
        add(pageHost,wrapper,-1,-2);
        LinearLayout first=horizontal();
        TextView h=txt(activeKeyCount==15?"나의 15키 컨트롤러":"나의 12키 컨트롤러",Math.min(19,screenWidthDp<400?16:19),WHITE,true);
        h.setSingleLine(true);
        h.setEllipsize(android.text.TextUtils.TruncateAt.END);
        first.addView(h,new LinearLayout.LayoutParams(0,dp(33),1));
        TextView modelButton=pill(activeKeyCount==15?("PN15 "+("L".equals(activeSide)?"LEFT":"RIGHT")):"PN12",true);
        // Firmware dictates model/hand: no manual model selector.
        add(first,modelButton,screenWidthDp<400?115:135,34);
        add(center,first,-1,40);
        add(center,txt(activeKeyCount==15?("5열 × 3행 · "+("L".equals(activeSide)?"왼쪽":"오른쪽")+" 하강 대각선 배열"):"가로 3개 × 세로 4개 · 키별 자동 저장",11,MUTED,false),-1,28);
        LinearLayout frame=vertical();
        frame.setGravity(Gravity.CENTER);
        frame.setBackground(shape(INNER,15,0xff2c4364));
        pad(frame,12,11);
        keysHost=vertical();
        keysHost.setGravity(Gravity.CENTER);
        add(frame,keysHost,-1,-2);
        int frameWidth=Math.min(485,Math.max(245,centerWidthDp-44));
        add(center,frame,frameWidth,-2);
        renderKeys();
        space(center,12);
        LinearLayout bottom=horizontal();
        profileCaption=txt("프로필 · "+store.profileName(store.activeProfile()),11,MUTED,false);
        profileCaption.setSingleLine(true);
        profileCaption.setEllipsize(android.text.TextUtils.TruncateAt.END);
        bottom.addView(profileCaption,new LinearLayout.LayoutParams(0,dp(39),1));
        add(bottom,click("▷  동작 테스트",()->testSelected()),screenWidthDp<400?118:144,39);
        add(center,bottom,-1,40);
        rightHost=right;
        renderRight();
    }
    private void renderKeys(){
        if(keysHost==null||page!=0)return;
        keysHost.removeAllViews();
        if(activeKeyCount==15) { renderPn15Keys(); return; }
        int frameWidth=Math.min(485,Math.max(245,centerWidthDp-44));
        int realKeyWidth=(frameWidth-24-18)/3;
        int keyHeight=Math.max(84,Math.min(113,Math.round(realKeyWidth*.98f)));
        int iconSize=Math.min(39,Math.max(25,keyHeight-48));
        boolean narrow=keyHeight<96;
        for(int r=0;r<4;r++){
            LinearLayout row=horizontal();
            row.setGravity(Gravity.CENTER);
            for(int c=0;c<3;c++){
                final int k=r*3+c+1;
                LinearLayout key=vertical();
                key.setGravity(Gravity.CENTER_HORIZONTAL);
                key.setPadding(dp(narrow?3:5),dp(3),dp(narrow?3:5),dp(3));
                int color=ActionStore.COLORS[store.color(k)];
                key.setBackground(grad(adjust(color,.19f),color,12,k==selectedKey?BLUE:0xffa3bdd7));
                TextView number=txt(String.format(Locale.US,"%02d",k),narrow?9:10,0xff202e45,true);
                add(key,number,-1,narrow?13:15);
                View icon=createSymbol(store.kind(k),iconSize,k);
                add(key,icon,iconSize,iconSize);
                TextView name=txt(store.name(k),narrow?9:10,0xff162133,true);
                name.setGravity(Gravity.CENTER);
                name.setSingleLine(true);
                name.setEllipsize(android.text.TextUtils.TruncateAt.END);
                add(key,name,-1,narrow?17:19);
                TextView type=txt(ActionStore.NAMES[store.kind(k)],narrow?8:9,0xff41536b,false);
                type.setGravity(Gravity.CENTER);
                type.setSingleLine(true);
                type.setEllipsize(android.text.TextUtils.TruncateAt.END);
                add(key,type,-1,narrow?13:15);
                key.setOnClickListener(v->{selectedKey=k;renderKeys();renderRight();});
                LinearLayout.LayoutParams kp=new LinearLayout.LayoutParams(0,dp(keyHeight),1);
                kp.setMargins(dp(3),dp(3),dp(3),dp(3));
                row.addView(key,kp);
                keyViews[k-1]=key;
            }
            add(keysHost,row,-1,-2);
        }
    }

    // PN15 CAD layout: 19.04 mm horizontal pitch, 5.93 mm downward per
    // column, and 19.0 mm downward per row. Keep key numbering row-major.
    private void renderPn15Keys(){
        int frameWidth=Math.min(485,Math.max(245,centerWidthDp-44));
        int gap=6;
        int keyWidth=Math.max(32,(frameWidth-38-gap*4)/5);
        int keyHeight=Math.max(49,Math.round(keyWidth*1.04f));
        int rowPitch=keyHeight+gap;
        int colSlant=Math.round(rowPitch*(5.93f/19.0f));
        int height=3*keyHeight+2*gap+4*colSlant+8;
        FrameLayout diagonal=new FrameLayout(this);
        add(keysHost,diagonal,-1,height);
        for(int row=0;row<3;row++){
            for(int col=0;col<5;col++){
                final int k=row*5+col+1;
                LinearLayout key=vertical();
                key.setGravity(Gravity.CENTER_HORIZONTAL);
                key.setPadding(dp(2),dp(2),dp(2),dp(1));
                int color=ActionStore.COLORS[store.color(k)];
                key.setBackground(grad(adjust(color,.19f),color,10,k==selectedKey?BLUE:0xffa3bdd7));
                TextView number=txt(String.format(Locale.US,"%02d",k),keyWidth<57?9:10,0xff202e45,true);
                number.setGravity(Gravity.CENTER);
                add(key,number,-1,keyWidth<57?12:16);
                int iconSize=Math.max(17,Math.min(34,keyWidth/2-2));
                View artwork=createSymbol(store.kind(k),iconSize,k);
                add(key,artwork,iconSize,iconSize);
                if(keyWidth>=57){
                    TextView label=txt(store.name(k),9,0xff162133,true);
                    label.setGravity(Gravity.CENTER);
                    label.setSingleLine(true);
                    label.setEllipsize(android.text.TextUtils.TruncateAt.END);
                    add(key,label,-1,18);
                }
                key.setOnClickListener(v->{selectedKey=k;renderKeys();renderRight();});
                FrameLayout.LayoutParams p=new FrameLayout.LayoutParams(dp(keyWidth),dp(keyHeight));
                p.leftMargin=dp(col*(keyWidth+gap));
                p.topMargin=dp(row*rowPitch+("L".equals(activeSide)?4-col:col)*colSlant);
                diagonal.addView(key,p);
                keyViews[k-1]=key;
            }
        }
    }
    private void switchDetectedModel(int count,String side){
        if(count!=12 && count!=15)return;
        String normalized=count==15&&"L".equals(side)?"L":"R";
        if(activeKeyCount==count && activeSide.equals(normalized))return;
        activeKeyCount=count;
        activeSide=normalized;
        store.setDeviceModel(count,normalized);
        selectedKey=Math.min(selectedKey,count);
        // Header, key positions and model label all depend on handedness.
        buildShell();
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
        int cols=rightWidthDp<395?2:3;
        for(int r=0;r<(types.length+cols-1)/cols;r++){
            LinearLayout line=horizontal();
            for(int col=0;col<cols;col++){
                int pos=r*cols+col;
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
        if(kind==9)hint="단축키 (안전 설치판에서는 실행 제한)";
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
                add(rightHost,txt("외부 앱의 복사·붙여넣기 기능은 현재 버전에서 지원하지 않습니다. 접근성 서비스를 제거해 설치 차단 위험을 줄였습니다.",10,MUTED,false),-1,63);
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
        add(panel,txt("프로필마다 "+activeKeyCount+"개 버튼의 기능, 이름, 색상을 별도로 저장합니다.",13,MUTED,false),-1,40);
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
        add(card,txt("PulseNote PN12 / PN15 · ESP32-C3 · USB CDC 115200bps",14,MUTED,false),-1,38);
        add(card,click(autoEnabled()?"●  USB 자동 연결: 켜짐":"○  USB 자동 연결: 꺼짐",()->{
            boolean next=!autoEnabled();
            getSharedPreferences("pn12_permissions",MODE_PRIVATE).edit().putBoolean("auto_connect",next).apply();
            renderPage();
            if(next)maybeAutoConnect();
        }),Math.min(300,contentWidthDp-44),43);
        space(card,10);
        add(card,txt("앱을 열거나 PulseNote USB 케이블을 연결하면 자동으로 접속합니다. 최초 연결은 Android의 USB 허용 창에서 승인해야 합니다.",11,MUTED,false),-1,58);
        devicePickerLabel=txt("USB-C 데이터 케이블을 연결한 후 장치를 검색하세요.",12,MUTED,false);
        add(card,devicePickerLabel,-1,48);
        chooser=new Spinner(this);
        chooser.setBackground(shape(0xff0d1b33,11,BORDER));
        add(card,chooser,Math.min(650,contentWidthDp-44),53);
        space(card,14);
        LinearLayout row=contentWidthDp<680?vertical():horizontal();
        add(row,click("⟳  USB 장치 검색",()->scan()),contentWidthDp<680?-1:180,48);
        if(contentWidthDp<680)space(row,7);else space(row,12);
        add(row,click("●  PulseNote USB 연결",()->connect()),contentWidthDp<680?-1:200,48);
        if(contentWidthDp<680)space(row,7);else space(row,12);
        add(row,click("■  연결 종료",()->{
            // An explicit user disconnect must not be immediately reversed
            // by a later Activity.onResume callback.
            getSharedPreferences("pn12_permissions",MODE_PRIVATE).edit().putBoolean("auto_connect",false).apply();
            startService(new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.STOP));
            renderPage();
        }),contentWidthDp<680?-1:170,48);
        add(card,row,-1,-2);
        space(card,15);
        add(card,txt("현재 상태",14,WHITE,true),-1,29);
        TextView state=txt(Pn12UsbService.status,14,0xff3edba8,true);add(card,state,-1,41);
        add(card,txt("연결 후 상태는 상단의 USB 표시에도 반영됩니다. Android에서는 COM 포트 번호 대신 USB Host 장치 경로를 사용합니다.",12,MUTED,false),-1,55);
        add(card,txt("USB 통신은 화면 밖에서도 유지됩니다. 직접 실행을 켜면 PulseNote 연결 중 작은 PN 상태 표시가 나타나며, 앱·웹사이트를 곧바로 여는 기능을 시도합니다.",12,MUTED,false),-1,69);
        space(card,16);
        buildPermissionSection(card);
        space(card,14);
        add(card,txt("PN12에는 기존 펌웨어를, PN15에는 15키 펌웨어를 사용합니다. 단축키는 Android 보안 정책과 접근성 서비스 지원 범위에 따라 제한될 수 있으며 앱 실행은 백그라운드 상태에서 알림을 통해 진행됩니다.",12,MUTED,false),-1,84);
        scan();
    }

    private void buildPermissionSection(LinearLayout parent){
        add(parent,txt("권한 및 직접 실행",19,WHITE,true),-1,39);
        add(parent,txt("PulseNote 장치는 화면 밖에서 USB 연결과 버튼 입력을 유지합니다. 백그라운드 앱 실행을 위해 '다른 앱 위에 표시' 권한을 이용하며, 선택적으로 작은 PN 상태 표시를 띄웁니다. 일부 Android 버전은 직접 실행을 제한할 수 있습니다.",12,MUTED,false),-1,88);

        boolean granted=Settings.canDrawOverlays(this);
        boolean toggled=getSharedPreferences("pn12_permissions",MODE_PRIVATE).getBoolean("direct_launch_enabled",false);
        add(parent,txt("백그라운드 직접 실행  ·  "+(granted&&toggled?"활성화됨":granted?"실행 모드 꺼짐":"권한 필요"),13,
           granted&&toggled?0xff3edba8:WHITE,true),-1,32);
        add(parent,txt(granted&&toggled?
             "PulseNote USB 연결 중 우측에 작은 PN 표시가 보이면 직접 실행 준비가 된 상태입니다.":
             "권한을 허용하면 직접 실행 모드가 자동으로 켜집니다. 언제든 끌 수 있습니다.",
             11,MUTED,false),-1,52);
        add(parent,click(granted?"권한 설정 확인":"다른 앱 위에 표시 권한 허용",()->requestOverlayPermission()),310,45);
        space(parent,8);
        add(parent,click(toggled?"●  직접 실행 끄기":"▷  직접 실행 켜기",()->{
            if(toggled){
                setDirectLaunch(false);
            }else if(!Settings.canDrawOverlays(this)){
                new AlertDialog.Builder(this).setTitle("직접 실행 권한")
                  .setMessage("앱을 떠나서도 PulseNote 버튼으로 프로그램을 열려면 화면 위 표시 권한을 허용해야 합니다. 권한이 허용되면 PulseNote 연결 중 화면 우측에 작은 PN 버튼이 표시됩니다. 사용하지 않을 때는 언제든 끌 수 있습니다.")
                  .setNegativeButton("취소",null)
                  .setPositiveButton("설정 열기",(d,w)->{setDirectLaunch(true);requestOverlayPermission();})
                  .show();
            }else setDirectLaunch(true);
        }),300,43);
        space(parent,13);
        boolean notificationGranted=Build.VERSION.SDK_INT<33 ||
          checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS)==PackageManager.PERMISSION_GRANTED;
        add(parent,txt("알림  ·  "+(notificationGranted?"허용됨":"허용 필요"),13,notificationGranted?0xff3edba8:WHITE,true),-1,32);
        add(parent,click(notificationGranted?"알림 설정 확인":"알림 권한 요청",()->requestNotificationPermission()),245,43);
        space(parent,13);
        add(parent,txt("시스템 단축키  ·  지원 제한 (접근성 서비스 제거)",12,MUTED,false),-1,40);
        add(parent,txt("USB 접근 권한은 'PulseNote USB 연결' 버튼을 누를 때 Android 시스템 창에서 요청합니다.",11,MUTED,false),-1,47);
        add(parent,txt("볼륨과 미디어 제어는 별도 특수 권한 없이 작동합니다. 백그라운드 앱 실행을 위해 화면 위 표시 권한을 요청하며, Android 보안 정책에 따라 일부 실행이 제한될 수 있습니다.",11,MUTED,false),-1,65);
    }
    private void setDirectLaunch(boolean enabled) {
        getSharedPreferences("pn12_permissions",MODE_PRIVATE).edit()
            .putBoolean("direct_launch_enabled",enabled)
            .putBoolean("direct_launch_explicit_off",!enabled)
            .apply();
        if(Pn12UsbService.status.startsWith("연결 완료")){
           Intent i=new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.REFRESH);
           startService(i);
        }
        if(page==2)renderPage();
    }
    private void syncOverlayPermission(){
        if(!Settings.canDrawOverlays(this))return;
        android.content.SharedPreferences prefs=getSharedPreferences("pn12_permissions",MODE_PRIVATE);
        if(!prefs.getBoolean("direct_launch_explicit_off",false) &&
                !prefs.getBoolean("direct_launch_enabled",false)){
            prefs.edit().putBoolean("direct_launch_enabled",true).apply();
        }
        if(Pn12UsbService.status.startsWith("연결 완료")){
            try{startService(new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.REFRESH));}
            catch(RuntimeException ignored){}
        }
    }
    private void requestOverlayPermission() {
        if(Settings.canDrawOverlays(this)){
            setDirectLaunch(true);
            return;
        }
        getSharedPreferences("pn12_permissions",MODE_PRIVATE).edit()
            .putBoolean("direct_launch_explicit_off",false).apply();
        try {
            Intent intent=new Intent(Settings.ACTION_MANAGE_OVERLAY_PERMISSION,
                  Uri.parse("package:"+getPackageName()));
            startActivity(intent);
        }catch(Exception error) {
            try{startActivity(new Intent(Settings.ACTION_APPLICATION_DETAILS_SETTINGS,
                   Uri.parse("package:"+getPackageName())));}
            catch(Exception ignored){updateStatus("권한 설정을 열 수 없습니다");}
        }
    }
    private void requestNotificationPermission() {
        if(Build.VERSION.SDK_INT>=33&&checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS)!=PackageManager.PERMISSION_GRANTED){
            requestPermissions(new String[]{Manifest.permission.POST_NOTIFICATIONS},NOTIFICATION_REQUEST);
        }else{
            try{startActivity(new Intent(Settings.ACTION_APP_NOTIFICATION_SETTINGS)
                 .putExtra(Settings.EXTRA_APP_PACKAGE,getPackageName()));}
            catch(Exception ignored){updateStatus("알림 설정 실행 실패");}
        }
    }
    @Override public void onRequestPermissionsResult(int request,String[] permissions,int[] results) {
        super.onRequestPermissionsResult(request,permissions,results);
        if(page==2)renderPage();
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
        if(labels.isEmpty())labels.add("PulseNote USB CDC 장치 없음");
        if(chooser!=null)chooser.setAdapter(new ArrayAdapter<>(this,android.R.layout.simple_spinner_dropdown_item,labels));
        if(devicePickerLabel!=null)devicePickerLabel.setText(candidates.isEmpty()?"장치를 찾지 못했습니다. USB-C 데이터 케이블과 USB Host 지원을 확인하세요.":"호환되는 USB CDC 장치 "+candidates.size()+"개 발견");
    }
    private UsbDevice chosen(){return candidates.isEmpty()?null:candidates.get(Math.max(0,Math.min(candidates.size()-1,chooser==null?0:chooser.getSelectedItemPosition())));}
    private boolean autoEnabled(){
        return getSharedPreferences("pn12_permissions",MODE_PRIVATE).getBoolean("auto_connect",true);
    }
    private UsbDevice usbFromIntent(Intent intent){
        if(intent==null)return null;
        if(Build.VERSION.SDK_INT>=33)
            return intent.getParcelableExtra(UsbManager.EXTRA_DEVICE,UsbDevice.class);
        return intent.getParcelableExtra(UsbManager.EXTRA_DEVICE);
    }
    private void maybeAutoConnect(){
        if(!autoEnabled()||usb==null)return;
        for(UsbDevice device:usb.getDeviceList().values()){
            if(!Pn12UsbService.candidate(device))continue;
            if(Pn12UsbService.isConnected(device.getDeviceName()))continue;
            if(usb.hasPermission(device))startDevice(device);
            else if(!permissionPromptActive)requestOrStart(device);
            if(permissionPromptActive)break;
        }
    }
    private void maybeAutoConnect(UsbDevice device){
        if(!autoEnabled()||usb==null||device==null||!Pn12UsbService.candidate(device))return;
        if(Pn12UsbService.isConnected(device.getDeviceName()))return;
        if(permissionPromptActive)return;
        requestOrStart(device);
    }
    private void connect(){
        scan();
        UsbDevice device=chosen();
        if(device==null){updateStatus("USB CDC 장치 없음");return;}
        requestOrStart(device);
    }
    private void requestOrStart(UsbDevice device){
        if(device==null||!Pn12UsbService.candidate(device))return;
        if(usb.hasPermission(device)){startDevice(device);return;}
        if(permissionPromptActive)return;
        pendingUsbDevice=device.getDeviceName();
        permissionPromptActive=true;
        try {
            Intent response=new Intent(PERMISSION).setPackage(getPackageName());
            PendingIntent callback=PendingIntent.getBroadcast(this,1,response,
                PendingIntent.FLAG_MUTABLE|PendingIntent.FLAG_UPDATE_CURRENT);
            usb.requestPermission(device,callback);
            updateStatus("PulseNote USB 권한 승인 대기");
        }catch(RuntimeException error){
            permissionPromptActive=false;pendingUsbDevice=null;
            updateStatus("USB 권한 요청 실패: "+error.getClass().getSimpleName());
        }
    }
    private void showControllerHome(){
        if(page==0 || isFinishing() || isDestroyed())return;
        page=0;
        refreshNavigation();
        renderPage();
    }
    private void startDevice(UsbDevice device){
        if(device==null||!usb.hasPermission(device))return;
        // Transition to the 12-key controller immediately after USB permission,
        // not only after the subsequent serial handshake has finished.
        showControllerHome();
        showingConnectionHome=true;
        Intent intent=new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.START);
        intent.putExtra(Pn12UsbService.DEVICE,device.getDeviceName());
        try {
            if(Build.VERSION.SDK_INT>=26)startForegroundService(intent);
            else startService(intent);
            updateStatus("PulseNote USB 자동 연결 중");
        }catch(RuntimeException error){
            updateStatus("USB 서비스 실행 제한: "+error.getClass().getSimpleName());
        }
    }
    @Override protected void onNewIntent(Intent intent){
        super.onNewIntent(intent);
        setIntent(intent);
        if(UsbManager.ACTION_USB_DEVICE_ATTACHED.equals(intent.getAction())){
            showControllerHome();
            scan();
            UsbDevice device=usbFromIntent(intent);
            if(device!=null)maybeAutoConnect(device);else maybeAutoConnect();
        }
    }
    private void updateStatus(String message){
        if(message==null)message="상태 확인 중";
        boolean okay=message.startsWith("연결 완료");
        // Models are chosen by verified per-device broadcast metadata.
        if(statusHeader!=null)statusHeader.setText(okay?"●  USB 연결됨":"●  "+message);
        if(statusSide!=null)statusSide.setText((okay?"●  USB 연결됨":"●  USB 상태")+"\n\n"+message);
        if(statusFooter!=null)statusFooter.setText("●  "+message);
        if(page==2&&devicePickerLabel!=null)devicePickerLabel.setText(message);
    }
    @Override protected void onSaveInstanceState(Bundle state){
        state.putInt("pn12.page",page);
        state.putInt("pn12.selectedKey",selectedKey);
        super.onSaveInstanceState(state);
    }
    @Override protected void onResume(){
        super.onResume();visible=true;syncOverlayPermission();updateStatus(Pn12UsbService.status);
        if(UsbManager.ACTION_USB_DEVICE_ATTACHED.equals(getIntent()==null?null:getIntent().getAction())&&
                usbFromIntent(getIntent())!=null)showControllerHome();
        if(page==2&&pageHost!=null)renderPage();
        if(Pn12UsbService.status.startsWith("연결 완료")){
            try{startService(new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.REFRESH));}
            catch(RuntimeException ignored){}
        }
        if(autoEnabled()&&!permissionPromptActive)maybeAutoConnect();
    }
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
