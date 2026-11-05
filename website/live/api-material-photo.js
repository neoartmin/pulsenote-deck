'use strict';
// Product-photo references. Displayed with original source attribution, not as
// photographs taken by PulseNote; product identity is not inferred from other images.
const images={
  petg:'https://down-vn.img.susercontent.com/file/vn-11134207-820l4-mg4rvf1dbz7w9b',
  epdm:'https://item-shopping.c.yimg.jp/i/n/okayasurubber_4573213470619_1'
};
module.exports=async(req,res)=>{
  if(req.method!=='GET')return res.status(405).end();
  const name=typeof req.query.name==='string'?req.query.name:'';
  const url=images[name];
  if(!url)return res.status(404).end();
  try{
    const response=await fetch(url,{signal:AbortSignal.timeout(12000),redirect:'follow',headers:{
      'Accept':'image/avif,image/webp,image/png,image/jpeg,*/*',
      'User-Agent':'Mozilla/5.0 (compatible; PulseNoteDeck-MaterialReference/1.0)'
    }});
    if(!response.ok)throw new Error('Image source HTTP '+response.status);
    const type=(response.headers.get('content-type')||'').split(';')[0].toLowerCase();
    if(!['image/jpeg','image/png','image/webp','image/avif'].includes(type))throw new Error('Unexpected content type');
    const contentLength=Number(response.headers.get('content-length')||0);
    if(contentLength>2500000)throw new Error('Image too large');
    const b=Buffer.from(await response.arrayBuffer());
    if(b.length<300||b.length>2500000)throw new Error('Invalid file size');
    const isJpeg=b[0]===255&&b[1]===216&&b[2]===255;
    const isPng=b.subarray(0,8).equals(Buffer.from([137,80,78,71,13,10,26,10]));
    const isWebp=b.toString('ascii',0,4)==='RIFF'&&b.toString('ascii',8,12)==='WEBP';
    const isAvif=b.toString('ascii',4,12).includes('ftypavif');
    if(!(isJpeg||isPng||isWebp||isAvif))throw new Error('Invalid image header');
    res.setHeader('Content-Type',type);
    res.setHeader('Cache-Control','public,max-age=86400,s-maxage=604800,stale-while-revalidate=86400');
    res.setHeader('X-Content-Type-Options','nosniff');
    return res.status(200).send(b);
  }catch(_){
    res.setHeader('Cache-Control','public,max-age=60,s-maxage=120');
    return res.status(502).end();
  }
};