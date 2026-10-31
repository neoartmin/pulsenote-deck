package kr.pulsenote.pn12usb;

import android.content.Context;
import android.graphics.*;
import android.graphics.drawable.*;
import android.net.Uri;
import android.view.Gravity;
import android.view.View;
import android.widget.*;
import java.io.*;
import java.net.*;
import java.util.*;
import java.util.concurrent.*;

public final class KeyArtwork {
    private static final ExecutorService work=Executors.newFixedThreadPool(2);
    private static final Map<String,Bitmap> cache=Collections.synchronizedMap(new LinkedHashMap<String,Bitmap>(40,0.75f,true){
        @Override protected boolean removeEldestEntry(Map.Entry<String,Bitmap> eldest){return size()>40;}
    });
    private KeyArtwork(){}
    private static int dp(Context c,int value){return Math.round(value*c.getResources().getDisplayMetrics().density);}
    private static GradientDrawable background(int fill,Context ctx){
        GradientDrawable d=new GradientDrawable(GradientDrawable.Orientation.TOP_BOTTOM,
            new int[]{lighten(fill),fill});
        d.setCornerRadius(dp(ctx,10));return d;
    }
    private static int lighten(int c){
        return Color.rgb(Math.min(255,Color.red(c)+32),Math.min(255,Color.green(c)+32),Math.min(255,Color.blue(c)+32));
    }
    public static View create(Context ctx,int kind,String value,int size){
        FrameLayout frame=new FrameLayout(ctx);
        int[] colors={0xff894fe8,0xffed8c44,0xff009fb8,0xffe23b79,0xff43b96b,0xff468be6,0xffa34bd9};
        frame.setBackground(background(colors[Math.floorMod(kind,colors.length)],ctx));
        String glyph=ActionStore.SYMBOLS[Math.max(0,Math.min(kind,ActionStore.SYMBOLS.length-1))];
        TextView fallback=new TextView(ctx);fallback.setGravity(Gravity.CENTER);
        fallback.setText(glyph);fallback.setTextSize(size>=42?25:21);fallback.setTextColor(Color.WHITE);
        fallback.setTypeface(null,1);
        frame.addView(fallback,new FrameLayout.LayoutParams(-1,-1));
        if(kind==8){
            try {
                Drawable icon=ctx.getPackageManager().getApplicationIcon(value);
                ImageView iv=image(ctx);iv.setImageDrawable(icon);
                frame.addView(iv,new FrameLayout.LayoutParams(-1,-1));
                fallback.setVisibility(View.INVISIBLE);
            }catch(Exception ignored){}
        }else if(kind==7&&value!=null){
            try{
                Uri parsed=Uri.parse(value);
                String host=parsed.getHost();
                if(host!=null&&host.matches("[a-zA-Z0-9.-]{1,180}") &&
                   ("https".equalsIgnoreCase(parsed.getScheme())||"http".equalsIgnoreCase(parsed.getScheme()))){
                    ImageView iv=image(ctx);frame.addView(iv,new FrameLayout.LayoutParams(-1,-1));
                    Bitmap existing=cache.get(host);
                    if(existing!=null){iv.setImageBitmap(existing);fallback.setVisibility(View.INVISIBLE);}
                    else{
                        work.execute(()->{
                            Bitmap bitmap=null;HttpURLConnection con=null;
                            try{
                                URL url=new URL("https://www.google.com/s2/favicons?sz=64&domain="+URLEncoder.encode(host,"UTF-8"));
                                con=(HttpURLConnection)url.openConnection();con.setConnectTimeout(2500);con.setReadTimeout(2500);
                                con.setInstanceFollowRedirects(false);
                                if(con.getResponseCode()==200)try(InputStream input=con.getInputStream(); ByteArrayOutputStream out=new ByteArrayOutputStream()){
                                    byte[] buf=new byte[4096];int n;while((n=input.read(buf))>0){
                                        if(out.size()+n>131072)break;out.write(buf,0,n);
                                    }
                                    bitmap=BitmapFactory.decodeByteArray(out.toByteArray(),0,out.size());
                                }
                            }catch(Exception ignored){}finally{if(con!=null)con.disconnect();}
                            if(bitmap!=null){
                                cache.put(host,bitmap);
                                Bitmap result=bitmap;
                                frame.post(()->{iv.setImageBitmap(result);fallback.setVisibility(View.INVISIBLE);});
                            }
                        });
                    }
                }
            }catch(Exception ignored){}
        }
        return frame;
    }
    private static ImageView image(Context ctx){
        ImageView image=new ImageView(ctx);image.setScaleType(ImageView.ScaleType.FIT_CENTER);
        int padding=dp(ctx,5);image.setPadding(padding,padding,padding,padding);return image;
    }
}
