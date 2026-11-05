'use strict';
const $=q=>document.querySelector(q); const $$=q=>[...document.querySelectorAll(q)];
const design=[
{src:'assets/design/housing-top.png',alt:'PN15 검정색 상단 하우징 3D 렌더링',label:'01 / UPPER HOUSING'},
{src:'assets/design/housing-port.png',alt:'PN15 USB-C 포트와 측면 하우징 3D 렌더링',label:'02 / USB-C PORT'},
{src:'assets/design/housing-angle.png',alt:'PN15 외곽 실루엣 3D 렌더링',label:'03 / BODY PROFILE'},
{src:'assets/design/housing-variant.png',alt:'PN15 다른 상판 변형 디자인 3D 렌더링',label:'04 / TOP SHELL VARIANT'}];
$$('[data-design]').forEach(btn=>btn.addEventListener('click',()=>{let i=Number(btn.dataset.design);$('#design-image').src=design[i].src;$('#design-image').alt=design[i].alt;$('#design-label').textContent=design[i].label;$$('[data-design]').forEach(b=>{let sel=b===btn;b.classList.toggle('active',sel);b.setAttribute('aria-pressed',String(sel));});}));
const photoData=[
{src:'assets/complete/hero.jpg',alt:'PN15 완조립 사선 사진',label:'THREE-QUARTER VIEW',description:'각인된 키캡과 전체 배치를 한 장에서 확인할 수 있는 대표 사진입니다.'},
{src:'assets/complete/front-left.jpg',alt:'PN15 완조립 좌측 정면',label:'FRONT LEFT',description:'전면 하우징과 키 배열, 키캡 구성의 균형을 보여주는 각도입니다.'},
{src:'assets/complete/front-right.jpg',alt:'PN15 완조립 우측 정면',label:'FRONT RIGHT',description:'우측에서 본 키캡 높이와 케이스 외형을 확인할 수 있습니다.'},
{src:'assets/complete/port-angle.jpg',alt:'PN15 USB-C 포트 사선',label:'PORT ANGLE',description:'USB-C 포트 접근성과 측면 마감 상태를 확인하기 좋은 사진입니다.'},
{src:'assets/complete/port-front.jpg',alt:'PN15 USB-C 포트 정면',label:'USB-C FRONT',description:'포트 전면부와 스위치 장착 상태를 보다 가까이 보여줍니다.'},
{src:'assets/complete/side-angle.jpg',alt:'PN15 완조립 측면',label:'SIDE PROFILE',description:'측면 경사와 하우징 높이, 전체적인 볼륨감을 확인할 수 있습니다.'}];
let photoIndex=0;let lastFocus=null;const lightbox=$('#lightbox');
function showPhoto(i){photoIndex=(i+photoData.length)%photoData.length;const p=photoData[photoIndex];$('#large-photo').src=p.src;$('#large-photo').alt=p.alt;$('#photo-count').textContent=String(photoIndex+1).padStart(2,'0')+' / '+String(photoData.length).padStart(2,'0')+' · ORIGINAL PHOTO';$('#photo-label').textContent=p.label;}
function openLightbox(i,sourceEl){lastFocus=sourceEl||document.activeElement;showPhoto(i);lightbox.showModal();document.body.classList.add('modal');$('#photo-close').focus();}
$$('[data-photo]').forEach(b=>b.addEventListener('click',()=>openLightbox(Number(b.dataset.photo),b)));
$('#photo-close').addEventListener('click',()=>lightbox.close());$('#photo-prev').addEventListener('click',()=>showPhoto(photoIndex-1));$('#photo-next').addEventListener('click',()=>showPhoto(photoIndex+1));lightbox.addEventListener('click',e=>{if(e.target===lightbox)lightbox.close()});lightbox.addEventListener('close',()=>{document.body.classList.remove('modal');if(lastFocus?.focus)lastFocus.focus()});document.addEventListener('keydown',e=>{if(!lightbox.open)return;if(e.key==='ArrowRight')showPhoto(photoIndex+1);if(e.key==='ArrowLeft')showPhoto(photoIndex-1)});

const galleryImage=$('#gallery-image'),galleryTitle=$('#gallery-title'),galleryDescription=$('#gallery-description'),galleryCount=$('#gallery-count'),galleryOpen=$('#gallery-open');
const galleryThumbs=$$('[data-gallery-index]');let galleryIndex=0;
function setGallery(i){galleryIndex=(i+photoData.length)%photoData.length;const p=photoData[galleryIndex];galleryImage.src=p.src;galleryImage.alt=p.alt;galleryTitle.textContent=p.label;galleryDescription.textContent=p.description;galleryCount.textContent=String(galleryIndex+1).padStart(2,'0')+' / '+String(photoData.length).padStart(2,'0');galleryOpen.dataset.photo=String(galleryIndex);galleryThumbs.forEach((btn,idx)=>{const active=idx===galleryIndex;btn.classList.toggle('active',active);btn.setAttribute('aria-selected',String(active));btn.tabIndex=active?0:-1;});}
galleryThumbs.forEach((btn,idx)=>{btn.addEventListener('click',()=>setGallery(idx));btn.addEventListener('keydown',e=>{if(!['ArrowRight','ArrowLeft','Home','End'].includes(e.key))return;e.preventDefault();let next=e.key==='Home'?0:e.key==='End'?galleryThumbs.length-1:(idx+(e.key==='ArrowRight'?1:-1)+galleryThumbs.length)%galleryThumbs.length;setGallery(next);galleryThumbs[next].focus();});});
$('#gallery-prev').addEventListener('click',()=>setGallery(galleryIndex-1));
$('#gallery-next').addEventListener('click',()=>setGallery(galleryIndex+1));
galleryOpen.addEventListener('click',()=>openLightbox(galleryIndex,galleryOpen));
function attachSwipe(el,onPrev,onNext){if(!el)return;let sx=0,sy=0;el.addEventListener('touchstart',e=>{const t=e.changedTouches[0];sx=t.clientX;sy=t.clientY},{passive:true});el.addEventListener('touchend',e=>{const t=e.changedTouches[0];const dx=t.clientX-sx,dy=t.clientY-sy;if(Math.abs(dx)<40||Math.abs(dx)<Math.abs(dy))return;dx>0?onPrev():onNext();},{passive:true});}
attachSwipe($('.gallery-stage'),()=>setGallery(galleryIndex-1),()=>setGallery(galleryIndex+1));
attachSwipe(lightbox,()=>showPhoto(photoIndex-1),()=>showPhoto(photoIndex+1));
setGallery(0);

// Current public installer, with fallback for offline API.
(function(){
 const setup=document.getElementById('installer-link'),note=document.getElementById('release-note'),backup=document.getElementById('backup-link'),versionBadge=document.getElementById('version');
 const fallback='https://github.com/neoartmin/pulsenote-deck/releases/download/pnd-windows-v7.5.22/PulseNoteDeck_v7.5.22_Setup_preview.exe';
 if(setup)setup.href=fallback;
 if(backup)backup.href='https://github.com/neoartmin/pulsenote-deck/releases/download/pnd-windows-v7.5.22/PulseNoteDeck_v7.5.22_COMPLETE_BACKUP.zip';
 if(versionBadge)versionBadge.textContent='v7.5.22';
 if(note)note.textContent='Windows v7.5.22 프리뷰 · 실제 Windows 앱';
 fetch('/api/latest',{cache:'no-store'}).then(r=>{if(!r.ok)throw Error('release API');return r.json();}).then(data=>{
  const w=data.windows;if(!w?.installer?.url)return;
  const u=new URL(w.installer.url);
  if(u.protocol!=='https:'||u.hostname!=='github.com'||!u.pathname.startsWith('/neoartmin/pulsenote-deck/releases/'))return;
  if(setup)setup.href=u.href;
  if(versionBadge)versionBadge.textContent=w.version;
  if(backup&&w.archive?.url){const zip=new URL(w.archive.url);if(zip.protocol==='https:'&&zip.hostname==='github.com'&&zip.pathname.startsWith('/neoartmin/pulsenote-deck/releases/'))backup.href=zip.href;}
  if(note)note.textContent='최신 Windows '+w.version+(w.version==='v7.5.22'?' 프리뷰':'');
 }).catch(()=>{});
})();

// Independent page navigation. Every hash identifies one visible product category.
const pageIds=['overview','design','hardware','software','purpose','usage','assembly','parts','history','downloads','license'];
const pagePanels=pageIds.map(id=>document.getElementById(id));
const pageLinks=[...document.querySelectorAll('[data-page-link]')];
function activatePage(raw,scroll=true){
 const id=pageIds.includes(raw)?raw:'overview';
 pagePanels.forEach(p=>{const active=p.id===id;p.hidden=!active;p.setAttribute('aria-hidden',String(!active));});
 pageLinks.forEach(link=>{const selected=link.dataset.pageLink===id;link.classList.toggle('is-active',selected);if(selected){link.setAttribute('aria-current','page');}else link.removeAttribute('aria-current');});
 const activeNav=pageLinks.find(x=>x.dataset.pageLink===id);
 if(activeNav){const wrap=document.getElementById('category-viewport');if(wrap){const x=activeNav.offsetLeft-wrap.offsetLeft;const available=wrap.clientWidth;wrap.scrollTo({left:Math.max(0,x-(available-activeNav.offsetWidth)/2),behavior:'auto'});}}
 if(scroll)window.scrollTo({top:0,behavior:'auto'});
 document.body.dataset.currentPage=id;
}
function resolvePageHash(hash){
 let raw;try{raw=decodeURIComponent(hash.slice(1)||'software')}catch{return {page:'overview',anchor:null}}
 if(raw==='model-downloads')return {page:'downloads',anchor:'model-downloads'};
 return {page:pageIds.includes(raw)?raw:'overview',anchor:null};
}
function openPageFromHash(){
 const route=resolvePageHash(location.hash);activatePage(route.page);
 if(route.anchor)requestAnimationFrame(()=>document.getElementById(route.anchor)?.scrollIntoView({block:'start'}));
}
window.addEventListener('hashchange',openPageFromHash);
window.addEventListener('popstate',openPageFromHash);
document.querySelectorAll('a[href^="#"]').forEach(link=>link.addEventListener('click',e=>{
 const raw=(link.getAttribute('href')||'').slice(1);if(!pageIds.includes(raw)&&raw!=='top'&&raw!=='model-downloads')return;
 if(raw==='model-downloads'){e.preventDefault();if(location.hash!=='#'+raw)location.hash=raw;else openPageFromHash();return;}
 e.preventDefault();const id=raw==='top'?'overview':raw;
 if(location.hash!=='#'+id)location.hash=id;activatePage(id);
}));
openPageFromHash();
