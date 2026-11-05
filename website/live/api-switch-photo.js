'use strict';
// Fixed-source images of the named commercial products, proxied through
// our own origin because website CSP allows only same-origin images.
// A source and credit link accompanies each image in the hardware page.
const sources={
 'cream-yellow':'https://down-th.img.susercontent.com/file/cn-11134207-7r98o-lnd9cdwrggij62',
 'sea-salt':'https://www.lemokey.com/cdn/shop/files/Keychron-Sea-Salt-Linear-Switch-2.jpg?v=1731381214&width=1445',
 'peach':'https://bizweb.dktcdn.net/100/484/752/products/outemu-peach-v3-1728306613019.jpg?v=1728306617000',
 'lime':'https://bizweb.dktcdn.net/100/484/752/products/outemu-lime-v3-1728306603558.jpg?v=1728306610947'
};
module.exports=async(req,res)=>{
 if(req.method!=='GET')return res.status(405).end();
 const name=typeof req.query.name==='string'?req.query.name:'';
 const uri=sources[name];if(!uri)return res.status(404).end();
 try{
  const response=await fetch(uri,{signal:AbortSignal.timeout(10000),redirect:'follow',headers:{'Accept':'image/avif,image/webp,image/jpeg,image/png,*/*','User-Agent':'Mozilla/5.0 (compatible; PulseNoteDeck-ProductOverview/1.0)'}});
  if(!response.ok)throw new Error('Retailer image '+response.status);
  const type=(response.headers.get('content-type')||'').toLowerCase().split(';')[0];
  const supported=['image/jpeg','image/png','image/webp','image/avif'].includes(type);
  const length=Number(response.headers.get('content-length')||0);
  if(!supported||length>1500000)throw new Error('Unexpected image response');
  const bytes=Buffer.from(await response.arrayBuffer());
  const jpeg=bytes[0]===255&&bytes[1]===216&&bytes[2]===255;
  const png=bytes.subarray(0,8).equals(Buffer.from([137,80,78,71,13,10,26,10]));
  const webp=bytes.toString('ascii',0,4)==='RIFF'&&bytes.toString('ascii',8,12)==='WEBP';
  const avif=bytes.toString('ascii',4,12).includes('ftypavif');
  if(bytes.length<512||bytes.length>1500000||!((type==='image/jpeg'&&jpeg)||(type==='image/png'&&png)||(type==='image/webp'&&webp)||(type==='image/avif'&&avif)))throw new Error('Invalid photo');
  res.setHeader('Content-Type',type);
  res.setHeader('Cache-Control','public,max-age=86400,s-maxage=604800,stale-while-revalidate=86400');
  res.setHeader('X-Content-Type-Options','nosniff');
  return res.status(200).send(bytes);
 }catch(e){res.setHeader('Cache-Control','public,max-age=60,s-maxage=120');return res.status(502).end();}
};