package kr.pulsenote.pn12usb;

import android.content.Context;
import android.content.Intent;
import android.graphics.Color;
import android.graphics.PixelFormat;
import android.graphics.drawable.GradientDrawable;
import android.os.Build;
import android.provider.Settings;
import android.view.Gravity;
import android.view.View;
import android.view.WindowManager;
import android.widget.TextView;

/** A small, genuinely visible status bubble for the optional direct-launch mode.
 * The user explicitly enables it and grants Android's overlay permission.
 * This does NOT guarantee that every Android version permits background launches.
 */
public final class DirectLaunchOverlay {
    private final Context context;
    private final WindowManager manager;
    private View bubble;
    private static volatile boolean visible=false;
    public DirectLaunchOverlay(Context context) {
        this.context=context;
        this.manager=(WindowManager)context.getSystemService(Context.WINDOW_SERVICE);
    }
    public static boolean isVisible(){return visible;}
    public static boolean allowed(Context ctx) {
        return Settings.canDrawOverlays(ctx) &&
            ctx.getSharedPreferences("pn12_permissions",Context.MODE_PRIVATE)
               .getBoolean("direct_launch_enabled",false);
    }
    public void refresh(boolean connected){
        if(!connected||!allowed(context)){remove();return;}
        if(bubble!=null)return;
        try {
            TextView badge=new TextView(context);
            badge.setText("PN");
            badge.setTextSize(11f);badge.setTextColor(Color.WHITE);
            badge.setGravity(Gravity.CENTER);
            GradientDrawable background=new GradientDrawable();
            background.setColor(0xe12b3e81);
            background.setCornerRadius(dp(12));
            background.setStroke(dp(1),0xff8794ff);
            badge.setBackground(background);
            badge.setContentDescription("PN12 직접 실행 활성화. 눌러서 PulseNote Deck 열기");
            badge.setOnClickListener(v->{
                Intent i=new Intent(context,MainActivity.class);
                i.addFlags(Intent.FLAG_ACTIVITY_NEW_TASK|Intent.FLAG_ACTIVITY_SINGLE_TOP);
                context.startActivity(i);
            });
            WindowManager.LayoutParams params=new WindowManager.LayoutParams(
                dp(36),dp(36),WindowManager.LayoutParams.TYPE_APPLICATION_OVERLAY,
                WindowManager.LayoutParams.FLAG_NOT_FOCUSABLE
                   |WindowManager.LayoutParams.FLAG_NOT_TOUCH_MODAL
                   |WindowManager.LayoutParams.FLAG_LAYOUT_IN_SCREEN,
                PixelFormat.TRANSLUCENT);
            params.gravity=Gravity.TOP|Gravity.END;
            params.y=dp(115);
            params.x=dp(8);
            manager.addView(badge,params);
            bubble=badge;
            visible=true;
        }catch(RuntimeException ex){bubble=null;visible=false;}
    }
    public void remove(){
        visible=false;
        if(bubble!=null){
            try{manager.removeView(bubble);}catch(RuntimeException ignored){}
            bubble=null;
        }
    }
    private int dp(int n){return (int)(n*context.getResources().getDisplayMetrics().density+.5f);}
}
