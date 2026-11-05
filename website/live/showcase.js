'use strict';
/* Actual Windows action list from gui_windows.go (v7.5.23 branch);
   functional gallery, switch hardware, verified milestone timeline, dynamic versions. */
(function(){
 const $=(q,root=document)=>root.querySelector(q);
 const esc=x=>String(x??'').replace(/[&<>"']/g,v=>({'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[v]));
 const repo='https://github.com/neoartmin/pulsenote-deck';
 const actions=[
  ['shortcut','단축키 실행','키보드 조합을 물리 버튼 하나에 할당합니다.','예: Ctrl+S 저장, Alt+Tab 창 전환','키보드'],
  ['text','텍스트 입력','저장해 둔 한글·영문·여러 줄 문구를 입력합니다.','예: 자주 사용하는 인사말 또는 문장','키보드'],
  ['capture','화면 캡처','Windows 캡처 기능을 신속히 호출합니다.','예: 작업 중인 설계 화면 저장','생산성'],
  ['copy_text','문구 복사','정해진 텍스트를 클립보드에 복사합니다.','예: 반복 사용하는 양식·안내문','키보드'],
  ['system_settings','설정 바로가기','Windows 설정의 자주 쓰는 항목을 호출합니다.','예: 운영체제 설정 화면 열기','시스템'],
  ['auto_click','오토클릭','지정된 간격·횟수로 마우스 클릭을 반복합니다.','예: 반복적인 테스트·단순 작업','자동화'],
  ['audio_output','스피커 전환','Windows 오디오 출력 장치를 전환합니다.','예: 헤드셋과 데스크톱 스피커 변경','미디어'],
  ['open_alarms','알람 실행','Windows 시계·알람 화면을 실행합니다.','예: 휴식·작업 시간 관리','생산성'],
  ['program','프로그램 실행','선택한 실행 파일이나 응용프로그램을 실행합니다.','예: CAD·편집 프로그램 열기','실행'],
  ['website','웹사이트 열기','저장된 웹 주소를 기본 브라우저로 엽니다.','예: 프로젝트 관리 사이트 열기','실행'],
  ['folder','폴더 열기','선택한 작업 폴더를 탐색기에서 엽니다.','예: 설계 파일·인쇄 프로젝트 폴더','실행'],
  ['volume_up','볼륨 높이기','Windows 시스템 볼륨을 올립니다.','예: 회의 후 소리 조정','미디어'],
  ['volume_down','볼륨 낮추기','Windows 시스템 볼륨을 내립니다.','예: 알림 볼륨 빠르게 줄이기','미디어'],
  ['volume_mute','음소거','시스템 소리의 음소거 상태를 바꿉니다.','예: 통화·집중 작업 중 즉시 음소거','미디어'],
  ['media_play','재생·일시정지','운영체제 미디어 재생 명령을 보냅니다.','예: 음악·동영상 멈춤','미디어'],
  ['media_next','다음 트랙','재생 중인 미디어의 다음 항목을 실행합니다.','예: 다음 음악으로 이동','미디어'],
  ['media_prev','이전 트랙','미디어의 이전 항목을 실행합니다.','예: 이전 곡으로 되돌아가기','미디어'],
  ['none','사용 안 함','버튼의 매크로 기능을 비활성화합니다.','예: 사용하지 않는 키의 오작동 예방','시스템'],
  ['web_search','웹 검색','지정한 키워드로 웹 검색을 실행합니다.','예: 기술 자료 빠른 검색','생산성'],
  ['qr_code','QR 코드 만들기','텍스트·URL을 QR 코드 이미지로 생성합니다.','예: 링크 공유용 PNG 만들기','생산성'],
  ['clipboard_history','클립보드 기록','Windows 클립보드 기록을 호출합니다.','예: 최근 복사 내용 찾아보기','생산성'],
  ['window_control','창 제어','현재 창의 위치·상태를 단축 명령으로 조정합니다.','예: 창 최대화·정렬·가상 데스크톱 이동','시스템']
 ];
 function textIcon(){return '<svg aria-hidden="true" focusable="false" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg"><rect x="2" y="2" width="96" height="96" rx="20" fill="#2363eb" stroke="#649dff" stroke-width="3"/><text x="11" y="70" font-family="Segoe UI,Arial,sans-serif" font-size="57" font-weight="700" fill="#fff">T</text><rect x="61" y="31" width="25" height="7" rx="3.5" fill="#fff"/><rect x="61" y="48" width="25" height="7" rx="3.5" fill="#fff"/><rect x="61" y="65" width="25" height="7" rx="3.5" fill="#fff"/></svg>';}
 function icon(name){return name==='text'?textIcon():'<img loading="lazy" decoding="async" src="/api/native-icon?name='+encodeURIComponent(name)+'" alt="" />';}
 const groupNames=['전체','키보드','실행','자동화','생산성','미디어','시스템'];
 const groups=groupNames.map(g=>'<button type="button" data-fg-category="'+esc(g)+'" class="pnd-gallery-filter'+(g==='전체'?' is-current':'')+'" aria-pressed="'+(g==='전체'?'true':'false')+'">'+esc(g)+'</button>').join('');
 function featureCard(item){
  return '<article class="pnd-function-card" data-fg-group="'+esc(item[4])+'"><div class="pnd-function-icon">'+icon(item[0])+'</div><div class="pnd-function-copy"><span class="pnd-function-kind">'+esc(item[4])+'</span><h4>'+esc(item[1])+'</h4><p>'+esc(item[2])+'</p><small>'+esc(item[3])+'</small></div></article>';
 }
 const mount=$('#pnd-feature-gallery');
 if(mount){
  mount.innerHTML='<section class="pnd-feature-block" aria-label="PulseNote Deck 주요 기능 아이콘 갤러리"><div class="pnd-showcase-heading"><span>REAL WINDOWS ACTIONS / ICON LIBRARY</span><h3>아이콘으로 살펴보는 <em>22가지 기능</em></h3><p>최신 Windows 제어 프로그램의 기능 목록을 기준으로 정리했습니다. 아이콘은 실제 프로그램의 이미지 자산과 텍스트 입력 벡터 도형을 사용합니다.</p></div><div class="pnd-gallery-filters" aria-label="기능 유형별 필터">'+groups+'</div><div class="pnd-feature-grid">'+actions.map(featureCard).join('')+'</div><p class="pnd-section-footnote">실행 예시는 사용자가 직접 구성하는 작업 예시입니다. 웹 편집기는 설정 파일을 저장하며, 실제 Windows 명령은 설치된 프로그램에서 동작합니다.</p></section>';
  mount.addEventListener('click',e=>{
   const btn=e.target.closest('[data-fg-category]');if(!btn)return;
   const label=btn.dataset.fgCategory;
   mount.querySelectorAll('[data-fg-category]').forEach(b=>{const active=b.dataset.fgCategory===label;b.classList.toggle('is-current',active);b.setAttribute('aria-pressed',String(active));});
   mount.querySelectorAll('.pnd-function-card').forEach(card=>{card.hidden=label!=='전체'&&card.dataset.fgGroup!==label;});
  });
 }
 const photos=[
  {id:'cream-yellow',type:'OFFICE / QUIET TACTILE',name:'오테뮤 크림 옐로우 프로',en:'Outemu Cream Yellow Pro',spec:'저소음 택타일 · 작동 약 45gf · 바닥 약 60gf',why:'입력 피드백은 유지하면서 타건 소음을 낮춰 주변 사람을 배려해야 하는 사무실·독서실 환경에 적합합니다. 무의식적 중복 입력을 줄이는 구분감도 장점입니다.',source:'https://www.clicknbrewcafe.com/products/outemu-cream-yellow-pro-silent-switch'},
  {id:'sea-salt',type:'GAMING / LINEAR',name:'하이무 바다소금',en:'Haimu Sea Salt Linear',spec:'리니어 · 작동 47±5gf · 이동 3.6±0.3mm',why:'걸림이 없는 리니어 입력과 공장 윤활, 2단 장스프링으로 빠르고 반복적인 조작이 필요한 게임에 어울립니다. 다만 게이밍 반응 속도는 펌웨어·디바운스·시스템 지연에도 영향을 받습니다.',source:'https://www.lemokey.com/products/haimu-sea-salt-linear-switch'},
  {id:'peach',type:'SOUND LINEUP / SILENT LINEAR',name:'오테뮤 피치',en:'Outemu Silent Peach',spec:'저소음 리니어 · V3 대표치 작동 약 40gf',why:'부드럽고 걸림 없는 키감과 감쇠된 타건음을 비교하기 위한 구성입니다. 소리 자체를 키우는 축이 아니라 조용하고 부드러운 타건감을 체험하는 축입니다.',source:'https://www.ozokeyboards.com/products/outemu-peach-v3-silent-linear-switches'},
  {id:'lime',type:'SOUND LINEUP / SILENT TACTILE',name:'오테뮤 라임',en:'Outemu Silent Lime',spec:'저소음 택타일 · V3 대표치 작동 약 35gf',why:'가벼운 구분감과 저소음 구조를 갖춰 피치의 리니어 키감과 비교하기 좋습니다. 소리 크기보다 입력감·잔향의 차이를 보여주는 라인업입니다.',source:'https://www.ozokeyboards.com/products/outemu-lime-v3-silent-tactile-switches'}
 ];
 const hardware=$('#pnd-switch-showcase');
 if(hardware){
  hardware.innerHTML='<section class="pnd-hardware-material" aria-label="스위치와 PETG 하드웨어 설계"><div class="pnd-showcase-heading"><span>SWITCH SELECTION / 4 OPTIONS</span><h3>사용 목적에 따라 고른 <em>4가지 스위치</em></h3><p>사무용·게이밍·저소음 타건감 비교라는 구체적인 목적을 기준으로 구성했습니다. 사진은 해당 제품의 판매처 제공 이미지로, 피치·라임은 V3 대표 사진이며 실제 보유품의 세대는 미확인입니다.</p></div><div class="pnd-switch-grid">'+photos.map(p=>'<article class="pnd-switch-card"><div class="pnd-switch-photo"><img src="/api/switch-photo?name='+p.id+'" alt="'+esc(p.en)+' 기계식 키보드 스위치 실제 제품 사진" loading="lazy" decoding="async"/><span class="pnd-photo-fallback" hidden>제품 사진을 불러오지 못했습니다.</span></div><div class="pnd-switch-body"><span class="pnd-switch-type">'+esc(p.type)+'</span><h4>'+esc(p.name)+'</h4><small>'+esc(p.spec)+'</small><p>'+esc(p.why)+'</p><a href="'+p.source+'" rel="noopener noreferrer" target="_blank">사진 및 사양 출처 · 제품 확인 ↗</a></div></article>').join('')+'</div><p class="pnd-section-footnote">타건음은 키캡·보강판·하우징·장착 방식과 사용 환경에 따라 달라집니다. 세대별 키압·스트로크도 달라질 수 있어 대표 수치로 표기했습니다.</p><div class="pnd-petg-layout"><div class="pnd-petg-image"><img src="assets/design/housing-angle.png" alt="PulseNote Deck 경사형 PETG 하우징 CAD 설계 이미지" loading="lazy"/><div class="pnd-petg-image-caption">실제 하우징 설계 자료 · 3D 프린팅용 모델</div></div><div class="pnd-petg-copy"><span class="pnd-switch-type">MATERIAL ENGINEERING / PETG</span><h3>PETG 하우징,<br/><em>소량 제작에 적합한 균형</em></h3><p>PLA보다 일반적으로 높은 내열성과 인성을 기대할 수 있으며, 적층 접착력·낮은 수축·상대적으로 쉬운 출력 덕분에 키 스위치 장착부와 분리형 하우징 등 반복 조립되는 구조에 적합합니다.</p><p>PETG는 ABS·ASA처럼 고온 챔버를 필수로 요구하지 않는 경우가 많아 소량 제조 시 장비 부담을 낮출 수 있습니다. 반면 실 끌림(stringing), 작은 정밀 형상과 브리징, 나사 체결부의 장기 크리프는 별도 관리가 필요합니다.</p><a href="https://help.prusa3d.com/article/petg_2059" target="_blank" rel="noopener noreferrer">PETG 물성·출력 특성 근거 · Prusa 자료 ↗</a></div></div><div class="pnd-wire-heading"><span>ASSEMBLY &amp; WIRING / LOW-COST OPTIMIZATION</span><h3>저비용을 유지하면서 <em>절연·배선 품질 확보</em></h3></div><div class="pnd-wiring-grid"><article><div class="pnd-wiring-symbol">01</div><h4>단심 구리선으로 고정 배선</h4><p>단심 구리선은 형태를 잡아 일정한 경로에 배치하기 쉽고 저렴합니다. 짧고 정돈된 경로와 접점 수 관리는 접촉 불량 및 간섭 위험을 낮추는 데 유리합니다.</p></article><article><div class="pnd-wiring-symbol">02</div><h4>교차점 절연·납땜부 보호</h4><p>노출 도체가 다른 열·행 배선, ESP32-C3 핀 또는 USB 부품과 닿지 않도록 절연 피복, 수축튜브·절연 테이프·기계적 고정으로 분리합니다. 조립 전 단락·연속성 검사를 수행해야 합니다.</p></article><article><div class="pnd-wiring-symbol">03</div><h4>신호 안정성은 회로 전체에서 결정</h4><p>단심 구리선 자체가 EMI를 제거하는 것은 아닙니다. 공통 접지, 올바른 풀업·배선 길이, 스위치 디바운스와 USB/전원 경로 관리가 실제 입력 안정성에 더 직접적으로 작용합니다.</p></article><article><div class="pnd-wiring-symbol">04</div><h4>소량 제작 중심의 비용 효율화</h4><p>범용 ESP32-C3, 기계식 스위치, PETG 출력 하우징 및 단심 배선을 조합해 전용 기판·금형의 초기 투자 부담을 줄입니다. 단, 수작업 납땜 시간과 유지보수성은 별도의 비용입니다.</p></article></div></section>';
  hardware.addEventListener('error',e=>{if(e.target.tagName==='IMG'&&e.target.closest('.pnd-switch-photo')){e.target.hidden=true;const f=e.target.parentNode.querySelector('.pnd-photo-fallback');if(f)f.hidden=false;}},true);
 }
 const assembly=$('#pnd-assembly-safety');
 if(assembly)assembly.innerHTML='<div class="pnd-assembly-warning"><strong>조립 안전 체크</strong><p>단심 구리선의 절연 피복 손상, 배선 교차점·납땜부 접촉, ESP32-C3 전원과 GND 단락을 조립 전 검사하세요. 전원이 차단된 상태에서 연속성·단락을 확인하고, 고정 배선은 반복 굽힘으로 인한 단선이 생기지 않도록 여유와 지지점을 확보해야 합니다.</p></div>';
 const milestones=[
  {tag:'pnd-windows-v7.5.22',date:'2026.11.05',version:'Windows v7.5.22',title:'모니터 DPI 대응·UI 선명도 개선',desc:'PerMonitorV2 DPI 처리와 프로그램 화면 렌더링 개선을 반영한 Windows 프리뷰를 공개했습니다.',icon:'screen'},
  {tag:'pnd-windows-v7.5.17',date:'2026.11.04',version:'Windows v7.5.17',title:'단축키 유지 토글·사용 안내',desc:'키 누름 유지 토글, 상태 알림 및 기능별 도움말을 정리하고 Windows 단독 배포를 진행했습니다.',icon:'shortcut'},
  {tag:'pnd-windows-v7.5.16',date:'2026.11.04',version:'Windows v7.5.16',title:'생산성 기능 확장',desc:'웹 검색, 이메일 작성, 로컬 QR, 키 시퀀스, 클립보드 기록 및 창 제어를 추가했습니다.',icon:'qr_code'},
  {tag:'pn12-pn15-android-v1.4.1',date:'2026.11.04',version:'Android v1.4.1',title:'Android 기능·권한 구조 개선',desc:'안드로이드 키 시퀀스와 웹 검색·공유 등을 확장하고 기본 APK의 접근성 권한 요구를 줄였습니다.',icon:'device'},
  {tag:'pnd-windows-v7.5.12',date:'2026.11.03',version:'Windows v7.5.12',title:'여러 줄 텍스트 매크로 개선',desc:'멀티라인 텍스트 편집과 채팅 입력 중 줄바꿈 처리를 개선했습니다.',icon:'text'}
 ];
 const tm=$('#pnd-version-timeline');
 if(tm){
  tm.innerHTML='<section class="pnd-history-version"><div class="pnd-showcase-heading"><span>SOFTWARE MILESTONES / VERIFIED RELEASES</span><h3>굵직한 버전 업데이트, <em>개발 타임라인</em></h3><p>GitHub 릴리스 본문과 실제 공개 날짜(한국시간)를 기준으로 정리한 주요 업데이트입니다. 작은 수정까지 나열하지 않고 사용 경험이 달라진 지점을 표시했습니다.</p></div><div id="pnd-current-milestones" class="pnd-current-milestones"><div id="pnd-current-release" class="pnd-milestone-current">최신 배포 버전 확인 중…</div><div id="pnd-current-development" class="pnd-milestone-development" hidden></div></div><div class="pnd-milestone-list">'+milestones.map((m,i)=>'<article class="pnd-milestone"><span class="pnd-milestone-no">'+String(milestones.length-i).padStart(2,'0')+'</span><div class="pnd-milestone-detail"><time>'+m.date+'</time><span class="pnd-milestone-tag">'+esc(m.version)+'</span><h4>'+esc(m.title)+'</h4><p>'+esc(m.desc)+'</p><a href="'+repo+'/releases/tag/'+m.tag+'" target="_blank" rel="noopener noreferrer">원본 릴리스 확인 ↗</a></div></article>').join('')+'</div></section>';
 }
 const versionRx=/^v?(\d+)\.(\d+)\.(\d+)$/;
 function compare(a,b){const x=(a||'').match(versionRx),y=(b||'').match(versionRx);if(!x||!y)return 0;for(let i=1;i<=3;i++){const d=Number(x[i])-Number(y[i]);if(d)return d;}return 0;}
 function updateVersion(current,source){
  let published=current&&versionRx.test(current.version)?current.version:null,sourceVersion=source&&versionRx.test(source.version)?source.version:null;
  const version=sourceVersion||published;
  if(!version)return;
  const isAhead=sourceVersion&&published&&compare(sourceVersion,published)>0;
  document.querySelectorAll('[data-pnd-current-version]').forEach(el=>el.textContent=version);
  document.querySelectorAll('[data-pnd-published-version]').forEach(el=>el.textContent=published||'미확인');
  const heading=$('#software .sectionhead p');
  if(heading)heading.textContent=heading.textContent.replace(/v\d+\.\d+\.\d+/g,version);
  const featureLabel=$('#software .subsection-head span');
  if(featureLabel)featureLabel.textContent=featureLabel.textContent.replace(/v\d+\.\d+\.\d+/g,version);
  const featureTitle=$('#software .subsection-head h3');
  if(featureTitle)featureTitle.textContent=featureTitle.textContent.replace(/v\d+\.\d+\.\d+/g,version);
  const toolbar=$('.pnw-editor-toolbar strong');
  if(toolbar)toolbar.textContent='PulseNote Deck '+version;
  const sidebar=$('.pnw-side-bottom small');
  if(sidebar)sidebar.textContent=version+' · WINDOWS CONFIG v7';
  if(window.PulseNoteWebConfig)window.PulseNoteWebConfig.version=version.replace(/^v/,'');
  const publishingStatus=isAhead?('개발 '+sourceVersion+' / 공개 배포 '+published):'최신 공개 배포 '+(published||version);
  document.title='PulseNote Deck | '+publishingStatus+' · PN12 / PN15';
  const title=$('meta[property="og:title"]');
  if(title)title.content='PulseNote Deck | '+publishingStatus;
  const dev=$('#pnd-current-development');
  if(dev){
   dev.hidden=!isAhead;
   if(isAhead)dev.innerHTML='<strong>개발 소스 '+esc(sourceVersion)+'</strong><p>공개 설치 파일은 '+esc(published)+'입니다. 개발 버전은 릴리스 검증 후 자동 반영됩니다.</p><a href="'+repo+'/pulls" target="_blank" rel="noopener noreferrer">개발 변경 내역 확인 ↗</a>';
  }
  const currentEl=$('#pnd-current-release');
  if(currentEl&&published){
   currentEl.innerHTML='<strong>최신 공개 Windows '+esc(published)+'</strong><p>GitHub Release의 설치 파일을 기준으로 자동 인식합니다.</p><a href="'+esc(current.releaseUrl||repo+'/releases')+'" target="_blank" rel="noopener noreferrer">최신 릴리스 열기 ↗</a>';
  }
 }
 function validateRelease(u){try{const url=new URL(u);return url.protocol==='https:'&&url.hostname==='github.com'&&url.pathname.startsWith('/neoartmin/pulsenote-deck/releases/')?url.href:null;}catch(_){return null;}}
 const publicReq=fetch('/api/latest',{cache:'no-store'}).then(r=>{if(!r.ok)throw Error('GitHub release');return r.json();});
 const sourceReq=fetch('/api/source-version',{cache:'no-store'}).then(r=>{if(!r.ok)throw Error('GitHub source');return r.json();});
 Promise.allSettled([publicReq,sourceReq]).then(results=>{
  const release=results[0].status==='fulfilled'?results[0].value?.windows:null;
  const source=results[1].status==='fulfilled'?results[1].value:null;
  const button=$('#installer-link'),archive=$('#backup-link'),badge=$('#version'),note=$('#release-note');
  if(release&&release.version&&validateRelease(release.installer?.url)){
   if(button)button.href=validateRelease(release.installer.url);
   if(archive&&validateRelease(release.archive?.url))archive.href=validateRelease(release.archive.url);
   if(badge)badge.textContent=release.version;
   if(note)note.textContent='GitHub 확인 완료 · Windows '+release.version+(release.version==='v7.5.22'?' 프리뷰':'');
  }
  updateVersion(release,source);
 }).catch(()=>{});
})();
