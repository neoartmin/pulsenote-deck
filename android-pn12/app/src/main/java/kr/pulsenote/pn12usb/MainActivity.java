package kr.pulsenote.pn12usb;
import android.Manifest;
import android.app.*;
import android.content.*;
import android.content.pm.PackageManager;
import android.hardware.usb.*;
import android.os.*;
import android.graphics.Color;
import android.view.*;
import android.widget.*;
import java.util.*;
public final class MainActivity extends Activity {
 public static volatile boolean visible=false;
 private static final String PERMISSION="kr.pulsenote.pn12usb.GRANT";
 private UsbManager usb;
 private LinearLayout root;
 private TextView status;
 private final List<UsbDevice> candidates=new ArrayList<>();
 private final Button[] buttons=new Button[12];
 private Spinner chooser;
 private BroadcastReceiver receiver;
 private int bg=Color.rgb(14,22,35),panel=Color.rgb(28,41,62);
 @Override public void onCreate(Bundle s){
   super.onCreate(s); usb=(UsbManager)getSystemService(USB_SERVICE);
   getWindow().setStatusBarColor(bg);getWindow().setNavigationBarColor(bg);
   ScrollView scroll=new ScrollView(this);scroll.setFillViewport(true);
   root=new LinearLayout(this);root.setOrientation(1);root.setPadding(24,24,24,24);root.setBackgroundColor(bg);scroll.addView(root);setContentView(scroll);
   TextView title=label("PulseNote  PN12",26);title.setTypeface(null,1);root.addView(title);
   root.addView(label("Android USB · 기존 펌웨어 유지",14));
   status=label(Pn12UsbService.status,15);root.addView(status);
   chooser=new Spinner(this);root.addView(chooser);
   Button scan=button("USB 장치 새로고침");scan.setOnClickListener(v->scan());root.addView(scan);
   Button connect=button("PN12 USB 연결");connect.setOnClickListener(v->connect());root.addView(connect);
   Button stop=button("연결 종료");stop.setOnClickListener(v->startService(new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.STOP)));root.addView(stop);
   root.addView(label("버튼을 눌러 기능을 지정하세요",16));
   for(int row=0;row<4;row++){
     LinearLayout line=new LinearLayout(this);line.setOrientation(LinearLayout.HORIZONTAL);
     for(int col=0;col<3;col++){
       final int key=row*3+col+1;
       Button b=button("");buttons[key-1]=b;
       line.addView(b,new LinearLayout.LayoutParams(0,110,1));
       b.setOnClickListener(v->edit(key));
     }
     root.addView(line);
   }
   refresh();
   receiver=new BroadcastReceiver(){@Override public void onReceive(Context c,Intent i){
     String a=i.getAction();
     if(PERMISSION.equals(a)){
       if(i.getBooleanExtra(UsbManager.EXTRA_PERMISSION_GRANTED,false))startChosen();
       else status.setText("USB 접근 권한 거부");
     }else if(Pn12UsbService.STATUS.equals(a))status.setText(i.getStringExtra(Pn12UsbService.TEXT));
     else if(Pn12UsbService.KEY.equals(a)) status.setText("KEY:"+i.getIntExtra(Pn12UsbService.NUMBER,0)+" · "+i.getStringExtra(Pn12UsbService.TEXT));
     else if(UsbManager.ACTION_USB_DEVICE_ATTACHED.equals(a)||UsbManager.ACTION_USB_DEVICE_DETACHED.equals(a))scan();
   }};
   IntentFilter f=new IntentFilter();f.addAction(PERMISSION);f.addAction(Pn12UsbService.STATUS);f.addAction(Pn12UsbService.KEY);f.addAction(UsbManager.ACTION_USB_DEVICE_ATTACHED);f.addAction(UsbManager.ACTION_USB_DEVICE_DETACHED);
   if(Build.VERSION.SDK_INT>=33)registerReceiver(receiver,f,Context.RECEIVER_NOT_EXPORTED);else registerReceiver(receiver,f);
   scan();
   if(Build.VERSION.SDK_INT>=33&&checkSelfPermission(Manifest.permission.POST_NOTIFICATIONS)!=PackageManager.PERMISSION_GRANTED)requestPermissions(new String[]{Manifest.permission.POST_NOTIFICATIONS},9);
 }
 private TextView label(String s,int size){TextView t=new TextView(this);t.setText(s);t.setTextColor(Color.WHITE);t.setTextSize(size);t.setPadding(10,14,10,14);return t;}
 private Button button(String s){Button b=new Button(this);b.setText(s);b.setTextColor(Color.WHITE);b.setBackgroundTintList(android.content.res.ColorStateList.valueOf(panel));return b;}
 private void refresh(){ActionStore data=new ActionStore(this);for(int i=0;i<12;i++)buttons[i].setText((i+1)+"\n"+data.label(i+1));}
 private void scan(){
   candidates.clear();List<String> labels=new ArrayList<>();
   for(UsbDevice d:usb.getDeviceList().values())if(Pn12UsbService.candidate(d)){candidates.add(d);labels.add(d.getDeviceName()+" (VID "+d.getVendorId()+")");}
   if(labels.isEmpty())labels.add("USB CDC 장치를 찾지 못함");
   chooser.setAdapter(new ArrayAdapter<>(this,android.R.layout.simple_spinner_dropdown_item,labels));
 }
 private UsbDevice chosen(){int n=chooser.getSelectedItemPosition();return n>=0&&n<candidates.size()?candidates.get(n):null;}
 private void connect(){
   UsbDevice d=chosen();if(d==null){status.setText("PN12 USB-C 데이터 케이블을 연결하세요");return;}
   if(usb.hasPermission(d))startChosen();
   else {
     PendingIntent p=PendingIntent.getBroadcast(this,1,new Intent(PERMISSION).setPackage(getPackageName()),PendingIntent.FLAG_MUTABLE|PendingIntent.FLAG_UPDATE_CURRENT);
     usb.requestPermission(d,p);
   }
 }
 private void startChosen(){UsbDevice d=chosen();if(d==null)return;Intent i=new Intent(this,Pn12UsbService.class).setAction(Pn12UsbService.START);i.putExtra(Pn12UsbService.DEVICE,d.getDeviceName());if(Build.VERSION.SDK_INT>=26)startForegroundService(i);else startService(i);}
 private void edit(int key){
   ActionStore store=new ActionStore(this);
   LinearLayout content=new LinearLayout(this);content.setPadding(24,12,24,10);content.setOrientation(1);
   Spinner spinner=new Spinner(this);spinner.setAdapter(new ArrayAdapter<>(this,android.R.layout.simple_spinner_dropdown_item,ActionStore.NAMES));spinner.setSelection(store.kind(key));content.addView(spinner);
   EditText value=new EditText(this);value.setSingleLine(true);value.setHint("웹 URL 또는 앱 패키지 이름");value.setText(store.value(key));content.addView(value);
   new AlertDialog.Builder(this).setTitle("버튼 "+key+" 설정").setView(content).setNegativeButton("취소",null).setPositiveButton("저장",(d,w)->{store.set(key,spinner.getSelectedItemPosition(),value.getText().toString());refresh();}).show();
 }
 @Override protected void onResume(){super.onResume();visible=true;if(status!=null)status.setText(Pn12UsbService.status);}
 @Override protected void onPause(){visible=false;super.onPause();}
 @Override protected void onDestroy(){if(receiver!=null)unregisterReceiver(receiver);super.onDestroy();}
}
