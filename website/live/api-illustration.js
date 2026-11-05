'use strict';
/* PulseNote original educational illustrations. Each image is an independently
   rendered vector concept diagram, NOT a photograph of a real manufactured unit.
   100% self-hosted SVG with bounded, allowlisted subjects; CSP-safe <img> assets. */
const scenes=new Set(['deck','keyboard','model','cad','printer','spool','switch','choices','matrix','chip','wire','usb','epdm','sound','game','office','desktop','program','profiles','mouse','screen','file','folder','browser','cloud','wallet','gear','robot','sensor','solder','assemble','case','measure','rework','workflow','note','book','history','download','firmware','files','license','shield','soundwave','light','color','repair','map','warning','timeline','test']);
const accents=['#d34a55','#8ea9d6','#5ebcae','#d0a67c','#9387d6','#d4d8e3'];
function icon(scene,a,n){
 const line=(x1,y1,x2,y2,c=a,w=7)=>'<path d="M'+x1+' '+y1+'L'+x2+' '+y2+'" stroke="'+c+'" stroke-width="'+w+'" stroke-linecap="round" fill="none"/>';
 const rect=(x,y,w,h,r,fill,stroke='#596880',sw=3)=>'<rect x="'+x+'" y="'+y+'" width="'+w+'" height="'+h+'" rx="'+r+'" fill="'+fill+'" stroke="'+stroke+'" stroke-width="'+sw+'"/>';
 const txt=(x,y,s,size=23,col='#e3e9f1')=>'<text x="'+x+'" y="'+y+'" font-family="Arial,sans-serif" font-size="'+size+'" font-weight="bold" fill="'+col+'">'+s+'</text>';
 const keygrid=(x,y,scale=1)=>{let q='<g transform="translate('+x+' '+y+') scale('+scale+')"><path d="M4 20L410 4L428 198L21 222Z" fill="#293243" stroke="#7e899e" stroke-width="5"/><path d="M21 222L428 198L429 247L24 264Z" fill="#0b0e17" stroke="#56637b" stroke-width="4"/>';for(let r=0;r<3;r++)for(let c=0;c<5;c++){const k=r*5+c;const color= [0,4,7,10,13].includes(k)?'#c94852':k%3===0?'#9ba9bf':'#d4d6df';let xx=38+c*76,yy=36+r*57;q+=rect(xx,yy,62,46,8,color,'#657085',2.5)+rect(xx+5,yy+4,50,7,2,'#ffffff35','none',0)}return q+'</g>'};
 const printer=()=>'<g>'+rect(186,96,404,345,19,'#141e2c','#718098',9)+line(215,148,567,148,'#d3d7e0',9)+rect(246,174,274,24,8,'#263a53')+rect(344,145,72,96,12,'#687a96')+'<path d="M368 246L395 246L383 279Z" fill="#ff8d71"/>'+rect(247,295,292,95,12,'#1d2737','#8b98b4',5)+rect(289,316,206,47,6,a)+line(221,419,562,419,'#737e98',8)+'</g>';
 const spool=()=>'<g><ellipse cx="387" cy="281" rx="168" ry="158" fill="#121a27" stroke="#7e8ba1" stroke-width="15"/><ellipse cx="387" cy="281" rx="134" ry="127" fill="'+a+'" stroke="#d6d7e0" stroke-width="7"/>'+Array.from({length:14},(_,i)=>'<ellipse cx="387" cy="281" rx="'+(130-i*5)+'" ry="'+(124-i*4)+'" fill="none" stroke="'+(i%2?'#282a38':'#ffffff45')+'" stroke-width="4"/>').join('')+'<ellipse cx="387" cy="281" rx="51" ry="52" fill="#141824" stroke="#b0b9c7" stroke-width="9"/></g>';
 const switchSvg=()=>'<g>'+rect(240,137,300,76,12,'#e5d6bd','#a2a6b4',6)+rect(267,192,245,146,19,a,'#c2c5d1',9)+rect(295,339,186,38,9,'#d5d8df','#8a96ac',5)+line(340,377,340,440,'#dfc897',11)+line(440,377,440,440,'#dfc897',11)+rect(351,110,67,70,10,'#d6d8e4','#687489',6)+'</g>';
 const matrix=()=>'<g>'+Array.from({length:3},(_,r)=>line(158,179+r*85,617,179+r*85,r===1?'#d05260':'#7f98bd',9)).join('')+Array.from({length:5},(_,c)=>line(204+c*97,137,204+c*97,410,'#b9c2d2',7)).join('')+Array.from({length:15},(_,i)=>{let x=204+i%5*97,y=179+Math.floor(i/5)*85;return rect(x-20,y-20,40,40,10,i%4===0?a:'#849ab6','#dee5f1',4)}).join('')+'</g>';
 const chip=()=>'<g>'+rect(253,118,270,286,25,'#17403d','#6aa798',9)+rect(309,176,157,174,19,'#142034','#a3bbd8',6)+Array.from({length:8},(_,i)=>rect(219,148+i*32,31,12,3,'#c6a871')+rect(526,148+i*32,31,12,3,'#c6a871')).join('')+Array.from({length:7},(_,i)=>line(326+i*20,165,326+i*20,355,'#426681',2)).join('')+txt(324,273,'ESP',46)+'</g>';
 const files=()=>'<g>'+rect(240,124,304,318,18,'#20314a','#8d9bb8',8)+'<path d="M449 124L449 214L544 214Z" fill="#7699bb"/>'+Array.from({length:5},(_,i)=>rect(286,247+i*32,212-i*13,9,4,i===0?a:'#8aa0b9','none',0)).join('')+'</g>';
 const screen=()=>'<g>'+rect(163,114,450,296,19,'#101827','#93a1bb',9)+rect(185,136,406,243,8,'#1f2d40','#566881',2)+Array.from({length:3},(_,r)=>Array.from({length:5},(_,c)=>rect(225+c*70,160+r*65,50,45,7,(c===r||c===4)?a:'#7b8eaa','#e8ecf240',1)).join('')).join('')+line(387,411,387,454,'#b6bed0',12)+line(318,456,458,456,'#b6bed0',12)+'</g>';
 const gear=()=>'<g><circle cx="393" cy="274" r="136" fill="#d4d9e3" stroke="#a0aec4" stroke-width="13"/>'+Array.from({length:12},(_,i)=>'<rect x="365" y="106" width="57" height="69" rx="9" fill="'+a+'" transform="rotate('+(i*30)+' 393 274)"/>').join('')+'<circle cx="393" cy="274" r="92" fill="#222e42" stroke="#c1c6d2" stroke-width="14"/><circle cx="393" cy="274" r="37" fill="#a1adc0"/></g>';
 const shield=()=>'<g><path d="M389 104L554 163V277Q551 395 389 461Q230 397 225 277V163Z" fill="#1c3048" stroke="'+a+'" stroke-width="12"/><path d="M309 278L367 336L472 208" fill="none" stroke="#f3f4f6" stroke-linecap="round" stroke-linejoin="round" stroke-width="19"/></g>';
 const wave=()=>'<g>'+Array.from({length:15},(_,i)=>{let x=155+i*32,h=20+Math.abs(Math.sin(i*1.31+n))*160;return rect(x,280-h/2,14,h,7,i%3===0?a:'#7896b4','none',0)}).join('')+'</g>';
 const robot=()=>'<g>'+rect(228,391,324,45,10,'#222f43','#8497aa',6)+rect(270,343,94,65,10,a)+line(319,344,379,258,'#b8c3d1',24)+line(379,258,495,190,'#bdc7d4',22)+'<circle cx="378" cy="258" r="31" fill="#27374d" stroke="#e5e9ed" stroke-width="9"/><circle cx="494" cy="190" r="29" fill="#27374d" stroke="#d2d9e3" stroke-width="8"/>'+line(495,190,525,255,'#d1d8df',14)+line(525,255,559,286,'#e7a19c',8)+line(525,255,491,290,'#e7a19c',8)+'</g>';
 const cost=()=>'<g>'+rect(222,146,345,259,20,'#22354d','#8c9fba',8)+rect(405,236,169,94,15,a,'#e29e9c',4)+'<circle cx="469" cy="280" r="21" fill="#f0e4d4"/>'+Array.from({length:3},(_,i)=>'<circle cx="'+(275+i*65)+'" cy="199" r="35" fill="#c7af7b" stroke="#f5dca3" stroke-width="5"/>').join('')+'</g>';
 const note=()=>'<g>'+rect(245,106,299,351,20,'#f0ede3','#98a2b5',9)+rect(345,91,103,31,10,a)+Array.from({length:6},(_,i)=>line(281,174+i*39,489-(i%3)*26,174+i*39,i===2?a:'#92a1b7',5)).join('')+'</g>';
 const usb=()=>'<g>'+rect(180,231,187,118,21,'#1a2e47','#99a9c2',8)+rect(367,254,165,74,8,'#c4ccd9','#e4e8ef',8)+rect(531,276,101,29,7,a,'#d4d5e2',5)+rect(204,266,101,12,5,'#7c92ae')+rect(204,301,101,12,5,'#7c92ae')+'</g>';
 const solder=()=>'<g>'+line(236,165,490,391,'#aebbc9',29)+line(228,170,490,396,a,7)+'<circle cx="517" cy="408" r="50" fill="#d2a87d" stroke="#f3c69f" stroke-width="10"/><path d="M390 343L510 330L593 399" fill="none" stroke="#a0b3c9" stroke-width="8" stroke-linecap="round"/><path d="M317 378Q380 342 424 405" fill="none" stroke="#8fa2bb" stroke-width="8"/></g>';
 const measure=()=>'<g>'+rect(147,195,487,141,20,'#c2a16f','#f7dfa8',7)+Array.from({length:21},(_,i)=>line(179+i*22,208,179+i*22,i%5===0?268:237,'#1d3244',i%5===0?4:2)).join('')+line(179,315,595,315,'#4f515c',7)+'</g>';
 const layer=()=>'<g>'+Array.from({length:4},(_,i)=>'<path d="M'+(215+i*9)+' '+(205+i*56)+'L'+(393+i*8)+' '+(112+i*56)+'L'+(570+i*8)+' '+(201+i*56)+'L'+(396+i*8)+' '+(288+i*56)+'Z" fill="'+(i%2?a:'#526a86')+'" fill-opacity="'+(.94-i*.1)+'" stroke="#d7dfeb" stroke-width="5"/>').join('')+'</g>';
 const base={deck:keygrid(178,133,1.02),keyboard:keygrid(178,133,1.02),model:keygrid(178,133,1.02),cad:keygrid(178,133,1.02),printer:printer(),spool:spool(),switch:switchSvg(),choices:switchSvg()+'<g transform="translate(320,0) scale(.56)">'+switchSvg()+'</g>',matrix:matrix(),chip:chip(),wire:matrix(),usb:usb(),epdm:keygrid(178,128,.95)+rect(220,420,320,34,8,'#31343b','#9eafb9',4),sound:wave(),soundwave:wave(),game:keygrid(178,133,1.02),office:screen(),desktop:screen(),program:screen(),profiles:screen(),mouse:screen(),screen:screen(),file:files(),folder:files(),browser:screen(),cloud:screen(),wallet:cost(),gear:gear(),robot:robot(),sensor:chip(),solder:solder(),assemble:layer(),case:keygrid(178,133,1.02),measure:measure(),rework:gear(),workflow:layer(),note:note(),book:note(),history:note(),download:files(),firmware:chip(),files:files(),license:shield(),shield:shield(),light:wave(),color:switchSvg(),repair:solder(),map:matrix(),warning:shield(),timeline:layer(),test:shield()};
 return base[scene]||keygrid(178,133,1.02);
}
module.exports=(req,res)=>{
 if(req.method!=='GET')return res.status(405).end();
 const topic=typeof req.query.topic==='string'?req.query.topic:'';
 if(!/^[a-z0-9-]{1,40}$/.test(topic))return res.status(400).end();
 const parts=topic.split('-'),name=parts[0];
 if(!scenes.has(name))return res.status(404).end();
 const n=Array.from(topic).reduce((sum,c)=>sum+c.charCodeAt(0),0);
 const a=accents[n%accents.length];
 const uid='p'+String(n);
 const shapes=icon(name,a,n);
 const grid=Array.from({length:9},(_,i)=>'<path d="M'+(i*95)+' 0V520" stroke="#8290a7" stroke-opacity=".08"/>').join('');
 const rings=Array.from({length:3},(_,i)=>'<circle cx="630" cy="190" r="'+(130+i*57)+'" fill="none" stroke="'+a+'" stroke-width="2" stroke-opacity="'+(.16-i*.045)+'"/>').join('');
 const deco=Array.from({length:7},(_,i)=>'<path d="M'+(40+i*97)+' 0L'+(120+i*97)+' 0L'+(40+i*97)+' 80Z" fill="'+a+'" opacity=".026"/>').join('');
 const svg='<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 780 520" preserveAspectRatio="xMidYMid meet" role="img" aria-label="PulseNote original '+name+' concept illustration">'+
 '<defs><linearGradient id="'+uid+'g" x1="0" y1="0" x2="1" y2="1"><stop offset="0%" stop-color="#19253a"/><stop offset="55%" stop-color="#0c121f"/><stop offset="100%" stop-color="#151827"/></linearGradient><radialGradient id="'+uid+'r"><stop offset="0%" stop-color="'+a+'" stop-opacity=".26"/><stop offset="100%" stop-color="'+a+'" stop-opacity="0"/></radialGradient><filter id="'+uid+'f" x="-40%" y="-40%" width="180%" height="180%"><feGaussianBlur stdDeviation="22"/></filter></defs>'+
 '<rect width="780" height="520" rx="17" fill="url(#'+uid+'g)"/>'+grid+deco+'<circle cx="600" cy="169" r="245" fill="url(#'+uid+'r)"/>'+rings+
 '<g opacity=".24"><path d="M34 440H128L153 420L179 458L205 406L235 440H330" fill="none" stroke="'+a+'" stroke-width="3" stroke-linecap="round" stroke-linejoin="round"/></g>'+
 '<g>'+shapes+'</g><path d="M34 486H746" stroke="'+a+'" stroke-opacity=".24" stroke-width="3"/>'+
 '<path d="M34 483H100" stroke="'+a+'" stroke-width="5"/><circle cx="732" cy="47" r="9" fill="'+a+'"/>'+
 '</svg>';
 res.setHeader('Content-Type','image/svg+xml; charset=utf-8');res.setHeader('X-Content-Type-Options','nosniff');res.setHeader('Cache-Control','public,max-age=86400,s-maxage=2592000,immutable');
 res.setHeader('Content-Security-Policy',"default-src 'none'; style-src 'unsafe-inline'; sandbox");
 return res.status(200).send(svg);
};