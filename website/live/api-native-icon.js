'use strict';
const permitted=new Set(['capture','check','clipboard_history','compose_email','copy_text','folder','media_next','media_play','media_prev','none','open_alarms','open_timers','program','qr_code','sequence','shortcut','system_settings','volume_down','volume_mute','volume_up','web_search','website','window_control','auto_click','audio_output','app_logo']);
const pinned='689092bac926057d04213d6828c1610cce6baec3';
module.exports=async (req,res)=>{
 if(req.method!=='GET')return res.status(405).end();
 const name=typeof req.query.name==='string'?req.query.name:'';
 if(!permitted.has(name))return res.status(404).end();
 const generated=new Set(['auto_click','audio_output','app_logo']);
 const upstream=generated.has(name)?
 'https://github.com/neoartmin/pulsenote-deck/releases/download/pnd-windows-v7.5.22/PulseNoteDeck_'+name+'.png':
 'https://raw.githubusercontent.com/neoartmin/pulsenote-deck/'+pinned+'/Windows/%EC%86%8C%EC%8A%A4/app/assets/'+name+'.png';
 try{
   const response=await fetch(upstream,{signal:AbortSignal.timeout(9000),headers:{'User-Agent':'PulseNoteDeck-Website/1.0'}});
   if(!response.ok)throw Error('not published');
   const payload=Buffer.from(await response.arrayBuffer());
   if(payload.length<64||payload.length>500000||!payload.subarray(0,8).equals(Buffer.from([137,80,78,71,13,10,26,10])))throw Error('Not a PNG');
   res.setHeader('Content-Type','image/png');res.setHeader('Cache-Control','public,max-age=3600,s-maxage=604800,stale-while-revalidate=86400');
   return res.status(200).send(payload);
 }catch(_){return res.status(404).end();}
};