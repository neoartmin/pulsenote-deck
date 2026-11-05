'use strict';
// Canonical Windows source version. Releases remain separately verified by /api/latest.
// Do not infer a downloadable build merely because source code was updated.
const base='https://api.github.com/repos/neoartmin/pulsenote-deck/contents/';
const branch='feature/pnd-auto-click-audio-notifications-dark-theme';
async function readVersion(path){
 const uri=base+path+'?ref='+encodeURIComponent(branch);
 const r=await fetch(uri,{headers:{'User-Agent':'PulseNoteDeck-Website','Accept':'application/vnd.github+json'},signal:AbortSignal.timeout(8500)});
 if(!r.ok)throw new Error('GitHub source status '+r.status);
 const d=await r.json();
 if(!d.content||d.encoding!=='base64')throw new Error('Unsupported source response');
 const source=Buffer.from(d.content.replace(/\s/g,''),'base64').toString('utf8');
 const regex=path.endsWith('model.go')?/const\s+version\s*=\s*"(\d+\.\d+\.\d+)"/:/const\s+productVersion\s*=\s*"(\d+\.\d+\.\d+)"/;
 const match=source.match(regex);
 if(!match)throw new Error('App version not found');
 return match[1];
}
module.exports=async(req,res)=>{
 if(req.method!=='GET')return res.status(405).end();
 res.setHeader('Content-Type','application/json; charset=utf-8');
 res.setHeader('Cache-Control','public,s-maxage=300,stale-while-revalidate=300');
 try{
  const [app,installer]=await Promise.all([
   readVersion('Windows/%EC%86%8C%EC%8A%A4/app/model.go'),
   readVersion('Windows/%EC%86%8C%EC%8A%A4/installer/version.go')
  ]);
  if(app!==installer)return res.status(409).json({error:'App and installer source versions do not match'});
  return res.status(200).json({version:'v'+app,branch,origin:'Windows sources',downloadable:false,checkedAt:new Date().toISOString()});
 }catch(e){return res.status(503).json({error:'Source version temporarily unavailable'});}
};