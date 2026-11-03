package kr.pulsenote.pn12usb;

import android.app.Activity;
import android.content.*;
import android.graphics.Bitmap;
import android.graphics.Color;
import android.net.Uri;
import android.os.Bundle;
import android.view.Gravity;
import android.view.View;
import android.widget.*;
import com.google.zxing.BarcodeFormat;
import com.google.zxing.MultiFormatWriter;
import com.google.zxing.common.BitMatrix;
import java.io.*;
import java.nio.charset.StandardCharsets;

/**
 * QR content is encoded locally with ZXing. No camera, internet, storage or
 * accessibility permissions are requested. Saving invokes Android's Storage
 * Access Framework document picker and writes only to the user-chosen URI.
 */
public final class QrDisplayActivity extends Activity {
    private static final int SAVE_QR=74;
    private Bitmap qr;
    private String original;

    @Override public void onCreate(Bundle state){
        super.onCreate(state);
        original=getIntent().getStringExtra("qr_text");
        if(original==null)original="";
        LinearLayout root=new LinearLayout(this);
        root.setOrientation(LinearLayout.VERTICAL);
        root.setPadding(24,30,24,24);
        root.setGravity(Gravity.CENTER_HORIZONTAL);
        root.setBackgroundColor(Color.rgb(17,27,47));
        TextView title=new TextView(this);
        title.setText("PulseNote · QR코드");
        title.setTextColor(Color.WHITE);
        title.setTextSize(22);
        title.setGravity(Gravity.CENTER);
        root.addView(title,new LinearLayout.LayoutParams(-1,60));
        ImageView image=new ImageView(this);
        image.setAdjustViewBounds(true);
        root.addView(image,new LinearLayout.LayoutParams(-1,0,1));
        TextView info=new TextView(this);
        info.setTextColor(Color.LTGRAY);
        info.setTextSize(13);
        info.setGravity(Gravity.CENTER);
        info.setMaxLines(3);
        info.setText(original);
        root.addView(info,new LinearLayout.LayoutParams(-1,74));
        Button save=new Button(this);
        save.setText("PNG로 저장할 위치 선택");
        root.addView(save,new LinearLayout.LayoutParams(-1,58));
        save.setOnClickListener(v->saveWithPicker());
        try{
            if(original.isEmpty()||original.getBytes(StandardCharsets.UTF_8).length>900)
                throw new IllegalArgumentException("QR 내용은 UTF-8 900바이트 이하여야 합니다");
            BitMatrix bits=new MultiFormatWriter().encode(
                original,BarcodeFormat.QR_CODE,512,512);
            qr=Bitmap.createBitmap(512,512,Bitmap.Config.ARGB_8888);
            for(int y=0;y<512;y++)for(int x=0;x<512;x++)
                qr.setPixel(x,y,bits.get(x,y)?Color.BLACK:Color.WHITE);
            image.setImageBitmap(qr);
        }catch(Exception error){
            save.setEnabled(false);
            info.setText("QR 생성 실패: "+error.getMessage());
        }
        setContentView(root);
    }

    private void saveWithPicker(){
        if(qr==null)return;
        try{
            Intent save=new Intent(Intent.ACTION_CREATE_DOCUMENT)
                .addCategory(Intent.CATEGORY_OPENABLE)
                .setType("image/png")
                .putExtra(Intent.EXTRA_TITLE,"PulseNote_QR.png");
            startActivityForResult(save,SAVE_QR);
        }catch(RuntimeException error){
            Toast.makeText(this,"문서 저장 화면을 열 수 없습니다",Toast.LENGTH_LONG).show();
        }
    }

    @Override protected void onActivityResult(int request,int result,Intent data){
        super.onActivityResult(request,result,data);
        if(request!=SAVE_QR||result!=RESULT_OK||data==null||data.getData()==null||qr==null)
            return;
        Uri target=data.getData();
        try(OutputStream out=getContentResolver().openOutputStream(target,"w")){
            if(out==null||!qr.compress(Bitmap.CompressFormat.PNG,100,out))
                throw new IOException("PNG 저장 실패");
            Toast.makeText(this,"QR코드 PNG 저장 완료",Toast.LENGTH_SHORT).show();
        }catch(Exception error){
            Toast.makeText(this,"QR 저장 실패: "+error.getMessage(),Toast.LENGTH_LONG).show();
        }
    }

    @Override protected void onDestroy(){
        if(qr!=null){qr.recycle();qr=null;}
        super.onDestroy();
    }
}
