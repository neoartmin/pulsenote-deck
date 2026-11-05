'use strict';
/* Visual explanation coverage 2026-11: original independent vector diagrams
   for every major card, step and rationale; real production photos are retained. */
(function(){
 const base='/api/illustration?topic=';
 function image(topic,alt,cls){
  const img=document.createElement('img');
  img.src=base+encodeURIComponent(topic);
  img.alt=alt+' — PulseNote 설명용 개념 일러스트 (실제품 사진 아님)';
  img.className=cls||'pnd-concept-image';
  img.loading='lazy';img.decoding='async';
  return img;
 }
 function visual(node,topic,alt,where='prepend'){
  if(!node||node.querySelector(':scope > .pnd-concept-visual'))return;
  const figure=document.createElement('figure');figure.className='pnd-concept-visual';figure.dataset.scene=topic;
  figure.append(image(topic,alt));
  const c=document.createElement('figcaption');c.textContent='CONCEPT VISUAL · '+alt;figure.append(c);
  if(where==='append')node.append(figure);
  else if(where==='aftertitle'){
   const h=node.querySelector('h3,h4,h5,b,strong');if(h&&h.parentElement===node)h.after(figure);else node.prepend(figure);
  }else node.prepend(figure);
 }
 function series(sel,topics,alt,where='prepend'){
  const els=[...document.querySelectorAll(sel)];
  els.forEach((el,i)=>{const heading=el.querySelector('h3,h4,b,strong')?.textContent?.trim().replace(/\s+/g,' ').slice(0,58)||alt||'제품 정보';
   visual(el,topics[i%topics.length]+'-'+String(i+1).padStart(2,'0'),heading,where);
  });
 }
 function narrative(selector,topics){
  document.querySelectorAll(selector).forEach((p,i)=>{
   if(p.closest('.pnd-illustrated-paragraph'))return;
   const row=document.createElement('div');row.className='pnd-illustrated-paragraph';
   const fig=document.createElement('figure');fig.append(image(topics[i%topics.length]+'-story-'+(i+1),'브랜드 스토리 단계 '+(i+1),'pnd-paragraph-art'));
   const label=document.createElement('figcaption');label.textContent='PULSENOTE MAKING STORY · '+String(i+1).padStart(2,'0');fig.append(label);
   p.before(row);row.append(fig,p);
  });
 }
 // Every top-level section gets its own independent visual introduction.
 const intro={
  overview:['deck-introduction','완성형 데스크톱 컨트롤러 구조'],
  design:['cad-shape','경사형 15키 하우징 구조'],
  hardware:['chip-keyboard','스위치에서 입력 신호까지'],
  software:['screen-control','PC 전용 소프트웨어 키 설정'],
  purpose:['printer-personal','제작 이유와 직접 만드는 과정'],
  usage:['workflow-setup','사용 시작의 전체 흐름'],
  assembly:['assemble-parts','부품 조립 순서'],
  parts:['chip-components','내부 구성 요소'],
  history:['history-evolution','제품 설계와 개발 기록'],
  downloads:['download-program','제어 프로그램과 펌웨어 파일'],
  license:['license-rights','프로젝트 라이선스와 이용 범위']
 };
 for(const [id,[topic,alt]] of Object.entries(intro)){
  const page=document.getElementById(id),heading=page?.querySelector(':scope > .sectionhead');
  if(!page)return;
  const origin=heading||(id==='overview'?page.querySelector(':scope > .hero'):null);
  if(origin){const f=document.createElement('div');f.className='pnd-intro-art';f.append(image(topic,alt));
    const meta=document.createElement('span');meta.textContent='PULSENOTE · EXPLANATORY ART';f.append(meta);origin.after(f);}
 }
 // Product overview: the three main benefits should all contain visuals.
 series('#overview .launch-card',['cad','chip','screen'],'제품 구성','prepend');
 series('#overview .hero-metrics > div',['keyboard','chip','usb'],'핵심 사양');
 visual(document.querySelector('#overview .overview-final'),'deck-finished','맞춤형 데스크톱 입력 장치');
 // Hardware design: iconography for each dimensional and ergonomic claim.
 series('#design .design-spec > div',['matrix','usb','printer'],'하우징 상세 설계');
 series('#design .design-metrics > div',['measure','measure','cad','switch'],'설계 치수');
 series('#design .advantage-grid article',['keyboard','switch','assemble','printer'],'설계 가치');
 narrative('#design .product-lead > p',['measure','switch']);
 narrative('#design .tech-story > div:last-child > p',['assemble','cad']);
 // Hardware: correct the visibly mismatched old photos with exact explanatory diagrams.
 const parts=['spool','shield','wire'];
 const hardwareCards=[...document.querySelectorAll('#hardware .pnd-build-grid article')];
 hardwareCards.forEach((article,i)=>{
  const old=article.querySelector(':scope > img');if(old)old.remove();
  visual(article,parts[i]+'-hardware-design',article.querySelector('h4')?.textContent||'하드웨어 설계');
 });
 // Product references already have genuine sourced photos; add detail illustration
 // to show how the switch works and EPDM is applied instead of showing fake photos.
 series('#hardware .pnd-switch-card',['switch','switch','sound','switch'],'기계식 스위치 구조','append');
 series('#hardware .pnd-selected-card',['spool','epdm'],'소재와 책상 접촉');
 narrative('#hardware .tech-story > div:last-child > p',['matrix','usb']);
 // Software: function icon is still the real native one; extra unique visual explains usage.
 const featureScenes=['keyboard','file','screen','note','screen','mouse','sound','note','program','browser','folder','sound','sound','sound','sound','sound','sound','shield','browser','file','note','screen'];
 document.querySelectorAll('#software .pnd-feature-card').forEach((card,i)=>{
  const title=card.querySelector('h4')?.textContent||'기능';
  visual(card,featureScenes[i]+'-action-'+String(i+1).padStart(2,'0'),title+'의 사용 장면','aftertitle');
 });
 narrative('#software .product-lead > p',['screen','profiles']);
 narrative('#software .tech-story > div:last-child > p',['chip','wire']);
 // Brand: do not rewrite the founder's original wording. Pair EACH paragraph with a
 // newly created independent concept image, plus a six-card making journey.
 narrative('#purpose .pnd-manifesto-prose > p',['soundwave','printer','robot','sensor','workflow','book']);
 const stage=[
  ['wallet-value','내게 맞는 도구를 합리적으로','기성 장비의 정해진 사양에 맞추기보다, 필요한 기능을 직접 고를 수 있는 도구를 만들고 싶었습니다.'],
  ['cad-personal-design','원하는 형태를 직접 설계','키의 개수와 각도, 포트 위치까지 CAD에서 결정하고 실제 사용감에 맞춰 수정합니다.'],
  ['printer-iteration','3D 프린팅으로 아이디어를 실물로','큰 초기 설비 부담 없이 출력하고 확인하며 하우징을 개선하는 방식을 선택했습니다.'],
  ['choices-switch-feel','내 취향의 키감과 소리','사무용 저소음부터 부드러운 게임용 리니어까지, 스위치 특성에 맞게 조합합니다.'],
  ['screen-personal-functions','자주 쓰는 기능을 원하는 키로','프로그램 실행, 단축키, 문구 입력, 오디오 제어를 작업 환경에 맞춰 배치합니다.'],
  ['note-making-log','완성보다 중요한 제작 기록','설계·실험·출력·수정·코딩까지 과정을 남기며 다음 프로젝트의 밑거름으로 만듭니다.']
 ];
 const purpose=document.getElementById('purpose'),manifesto=purpose?.querySelector('.pnd-brand-manifesto');
 if(purpose&&manifesto){
  const section=document.createElement('section');section.className='pnd-purpose-journey';
  section.innerHTML='<div class="pnd-journey-title"><span class="kicker">THE REASON / SIX MAKING PRINCIPLES</span><h3>왜 직접 만드는가? <em>그림으로 보는 여섯 가지 답.</em></h3><p>PulseNote Deck은 개인의 작업 방식과 취향을 선택할 수 있는 컨트롤러를 합리적으로 제작하려는 시도에서 시작했습니다.</p></div>';
  const grid=document.createElement('div');grid.className='pnd-journey-grid';
  stage.forEach(([topic,title,desc],i)=>{
   const card=document.createElement('article');card.className='pnd-journey-card';
   card.append(image(topic,title));
   const body=document.createElement('div');const num=document.createElement('span');num.textContent='0'+(i+1)+' / WHY PULSENOTE';const h=document.createElement('h4');h.textContent=title;
   const p=document.createElement('p');p.textContent=desc;body.append(num,h,p);card.append(body);grid.append(card);
  });section.append(grid);manifesto.after(section);
 }
 series('#purpose .advantage-grid article',['keyboard','cad','chip'],'맞춤 제작 설계 가치');
 narrative('#purpose .purpose-hero > p',['wallet','deck']);
 narrative('#purpose .tech-story > div:last-child > p',['printer','note']);
 // Usage: 4 setup stages + 3 complete use cases, and firmware model explanation.
 series('#usage .detailed-steps article',['download','usb','profiles','deck'],'사용법 단계');
 series('#usage .usecase-grid article',['cad','office','sound'],'작업 활용 예시');
 narrative('#usage .tech-story > div:last-child > p',['chip','usb']);
 // Assembly: every step illustrated, including protective insulation and EPDM.
 series('#assembly .assemble-list li',['case','switch','matrix','firmware','epdm','test'],'조립 단계');
 narrative('#assembly .tech-story > div:last-child > p',['repair','solder']);
 // Component list: give every row its own correct component image.
 const partsList=['chip','switch','keyboard','case','wire','spool','epdm','usb','firmware','screen'];
 const partsRows=[...document.querySelectorAll('#parts .component-table .component-row:not(.component-head)')];
 partsRows.forEach((row,i)=>{
  const title=row.querySelector('b')?.textContent||'구성 부품';visual(row,partsList[i]+'-parts-'+(i+1),title);
 });
 narrative('#parts .tech-story > div:last-child > p',['assemble','files']);
 // History: six stages and each meaningful software milestone get their own images.
 series('#history .timeline > div',['light','cad','chip','screen','note','deck'],'제작 과정');
 series('#history .pnd-release-timeline article',['screen','keyboard','file','sound','text','history'].map(x=>x==='text'?'note':x),'버전별 주요 업데이트');
 // Download: 3 app/firmware + 4 model files and two helpful notes.
 series('#downloads .download-card',['download','firmware','firmware','cad','case','keyboard','keyboard'],'제공 파일');
 series('#downloads .download-help > div',['firmware','download'],'다운로드 도움말');
 // License: illustrative usage scopes while preserving original legal caveats.
 series('#license .terms-grid > div',['shield','license'],'라이선스 범위');
 narrative('#license .license-card > div:first-child > p',['book','shield']);
 const count=document.querySelectorAll('.pnd-concept-visual img,.pnd-journey-card img,.pnd-intro-art img,.pnd-illustrated-paragraph img').length;
 document.documentElement.dataset.pndVisualCount=String(count);
})();
