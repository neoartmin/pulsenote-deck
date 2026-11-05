'use strict';
/* PulseNote Deck v7.5.22 browser configuration editor.
   This edits the EXACT config_unified.json v7 model. The Win32 program
   continues to execute actions; no device-control claims are made here. */
(function(){
var mount=document.getElementById('pnd-editor-mount');
if(!mount)return;
var KEY='pnd-real-config-v7-20261105';
var COLORS=[['파랑','#94B4D5'],['빨강','#D998A0'],['회색','#D6DEE8'],['보라','#B9A6D5'],['주황','#E4B48F'],['초록','#A0C9B4'],['노랑','#DED09E']];
var TYPES=['shortcut','text','capture','copy_text','system_settings','auto_click','audio_output','open_alarms','program','website','folder','volume_up','volume_down','volume_mute','media_play','media_next','media_prev','none','web_search','qr_code','clipboard_history','window_control'];
var NAMES={none:'기능 없음',shortcut:'단축키',text:'텍스트 입력',capture:'화면 캡처',copy_text:'문구 복사',system_settings:'설정 바로가기',auto_click:'오토클릭',audio_output:'출력 장치 전환',open_alarms:'알람 열기',program:'프로그램',website:'웹사이트',folder:'폴더',volume_up:'볼륨 +',volume_down:'볼륨 -',volume_mute:'음소거',media_play:'재생 / 정지',media_next:'다음 곡',media_prev:'이전 곡',web_search:'웹 검색',qr_code:'QR코드 생성',clipboard_history:'클립보드 기록',window_control:'창 제어'};
var PARAMS={shortcut:'키보드 단축키 (예: CTRL+S)',text:'입력할 텍스트',copy_text:'복사할 문구',program:'실행 파일 · 바로가기 경로',website:'웹사이트 주소',folder:'폴더 경로',web_search:'검색어',qr_code:'QR로 만들 내용 (900바이트 이내)',system_settings:'설정 종류',window_control:'창 동작',audio_output:'출력 장치 식별자',auto_click:'오토클릭 상세 설정'};
var NEEDS={shortcut:1,text:1,copy_text:1,program:1,website:1,folder:1,web_search:1,qr_code:1,system_settings:1,window_control:1,audio_output:1};
var state={config:null,tab:'keys',selected:0,actionPage:0,folder:null,status:'실제 Windows 설정 형식 v7 · 브라우저에서 편집',timer:null};
function esc(v){return String(v==null?'':v).replace(/[&<>"']/g,function(x){return {'&':'&amp;','<':'&lt;','>':'&gt;','"':'&quot;',"'":'&#39;'}[x]});}
function defKey(n){return {label:'버튼 '+String(n).padStart(2,'0'),type:'none',value:'',color:'회색'};}
function defModel(count){return {active:0,profiles:[{name:'기본 프로필',keys:Array.from({length:count},function(_,i){return defKey(i+1)})}]};}
function defaults(){return {version:7,selected:15,selectedSide:'R',preferredCOM:'',models:{'12':defModel(12),'15':defModel(15),'15L':defModel(15)}};}
function modelId(){return state.config.selected===12?'12':state.config.selectedSide==='L'?'15L':'15';}
function getModel(){return state.config.models[modelId()];}
function getProfile(){var m=getModel();return m.profiles[m.active];}
function getKey(){return getProfile().keys[state.selected];}
function normalize(input){
 if(!input||typeof input!=='object'||input.version!==7||!input.models||typeof input.models!=='object')throw Error('PulseNoteDeck 설정 버전 7의 models 객체가 필요합니다.');
 var c=JSON.parse(JSON.stringify(input));
 c.selected=c.selected===12?12:15;
 c.selectedSide=c.selectedSide==='L'?'L':'R';
 c.preferredCOM=typeof c.preferredCOM==='string'?c.preferredCOM.trim().toUpperCase():'';
 ['12','15','15L'].forEach(function(id){
   var count=id==='12'?12:15,m=c.models[id];
   if(!m||!Array.isArray(m.profiles)||!m.profiles.length)m=defModel(count);
   m.profiles=m.profiles.slice(0,8).map(function(p,pi){
     if(!p||typeof p!=='object')p={};
     p.name=typeof p.name==='string'&&p.name.trim()?p.name:'프로필 '+(pi+1);
     var keys=Array.isArray(p.keys)?p.keys.slice(0,count):[];
     while(keys.length<count)keys.push(defKey(keys.length+1));
     p.keys=keys.map(function(k,i){
       if(!k||typeof k!=='object')k=defKey(i+1);
       k.label=typeof k.label==='string'&&k.label?k.label.slice(0,45):defKey(i+1).label;
       k.type=TYPES.includes(k.type)||['sequence','compose_email','open_timers'].includes(k.type)?k.type:'none';
       k.value=typeof k.value==='string'?k.value.slice(0,4096):'';
       if(!COLORS.some(function(x){return x[0]===k.color}))k.color='회색';
       if(k.type!=='shortcut')delete k.toggle;
       else if(k.toggle!==true)delete k.toggle;
       return k;
     });
     return p;
   });
   m.active=Number.isInteger(m.active)&&m.active>=0&&m.active<m.profiles.length?m.active:0;
   c.models[id]=m;
 });
 c.version=7;
 return c;
}
function remember(){
 try{localStorage.setItem(KEY,JSON.stringify(state.config));}catch(_){}
}
try{var saved=localStorage.getItem(KEY);state.config=saved?normalize(JSON.parse(saved)):defaults();}catch(_){state.config=defaults();}
function logo(){return '<svg aria-hidden="true" viewBox="0 0 1024 1024" xmlns="http://www.w3.org/2000/svg"><rect x="36" y="36" width="952" height="952" rx="210" fill="#111114" stroke="#44454b" stroke-width="18"/><rect x="226" y="226" width="258" height="258" rx="54" fill="#d9d9de"/><rect x="540" y="226" width="258" height="258" rx="54" fill="#d62839"/><rect x="226" y="540" width="258" height="258" rx="54" fill="#d9d9de"/><rect x="540" y="540" width="258" height="258" rx="54" fill="#d9d9de"/></svg>';}
function textMacroArtwork(cls){
 // Direct vector translation of gui_windows.go's drawTextMacroArtwork,
 // preserving Win32 RGB(35,99,235), RGB(100,157,255), white T and 3 lines.
 return '<svg class="pnw-native-text-glyph '+esc(cls||'')+'" aria-hidden="true" focusable="false" viewBox="0 0 100 100" xmlns="http://www.w3.org/2000/svg">'
   +'<rect x="1.5" y="1.5" width="97" height="97" rx="20" fill="#2363eb" stroke="#649dff" stroke-width="3"/>'
   +'<text x="11" y="70" font-family="Segoe UI,Malgun Gothic,Arial,sans-serif" font-size="57" font-weight="700" fill="#ffffff">T</text>'
   +'<rect x="61" y="31" width="25" height="7" rx="3.5" fill="#ffffff"/>'
   +'<rect x="61" y="48" width="25" height="7" rx="3.5" fill="#ffffff"/>'
   +'<rect x="61" y="65" width="25" height="7" rx="3.5" fill="#ffffff"/>'
   +'</svg>';
}
function icon(t,cls){
 if(t==='text')return textMacroArtwork(cls);
 var allowed=['auto_click','audio_output','capture','clipboard_history','copy_text','folder','media_next','media_play','media_prev','none','open_alarms','program','qr_code','shortcut','system_settings','volume_down','volume_mute','volume_up','web_search','website','window_control'];
 return allowed.includes(t)?'<img class="'+(cls||'')+'" loading="lazy" src="/api/native-icon?name='+encodeURIComponent(t)+'" alt=""/>':'<span class="pnw-icon-fallback">'+esc(t==='auto_click'?'AC':t==='audio_output'?'AU':t==='text'?'T':t==='compose_email'?'@':'✦')+'</span>';
}
function button(action,label,cls,extra){return '<button type="button" class="'+(cls||'pnw-btn')+'" data-pnw="'+esc(action)+'" '+(extra||'')+'>'+label+'</button>';}
function activeTitle(){return state.config.selected===12?'PN12 · 12키':state.config.selectedSide==='L'?'PN15 · 왼손':'PN15 · 오른손';}
function keyDisplay(k,i){
 return '<button type="button" data-pnw="key" data-index="'+i+'" data-color="'+esc(k.color)+'" class="pnw-key '+(state.selected===i?'selected':'')+'"><span class="pnw-key-number">'+String(i+1).padStart(2,'0')+'</span><span class="pnw-key-icon">'+icon(k.type)+'</span><span class="pnw-key-label" data-keylabel="'+i+'">'+esc(k.label.slice(0,9))+'</span><small>'+esc(NAMES[k.type]||k.type)+'</small></button>';
}
function btnProfile(p,i,active){return button('profile',esc(p.name), 'pnw-profile-item '+(i===active?'selected':''),'data-index="'+i+'"');}
function tabs(){return '<div class="pnw-tabs">'+[['12','PN12'],['15L','15 왼손'],['15','15 오른손']].map(function(t){return button('model',esc(t[1]),'pnw-btn '+(modelId()===t[0]?'selected':''),'data-model="'+t[0]+'"')}).join('')+'</div>';}
function drawKeys(){
 var m=getModel(),p=getProfile(),key=getKey(),names=TYPES;
 var start=state.actionPage*9,count=Math.ceil(names.length/9);
 var panel='<div class="pnw-key-panel"><div class="pnw-key-heading"><div><h4>나의 '+state.config.selected+'키 컨트롤러</h4><p>'+(state.config.selected===12?'3열 × 4행 · 키별 자동 저장':'5열 × 3행 · '+(state.config.selectedSide==='L'?'왼쪽':'오른쪽')+' 하강 대각선 배열')+'</p></div>'+tabs()+'</div><div class="pnw-key-stage pnw-layout-'+modelId()+'">'+p.keys.map(keyDisplay).join('')+'</div><p class="pnw-profile-caption">프로필 · '+esc(p.name)+'</p></div>';
 var actions=names.slice(start,start+9).map(function(t,j){return button('type',icon(t,'pnw-pick-icon')+'<span>'+esc(NAMES[t])+'</span>','pnw-action '+(key.type===t?'selected':''),'data-type="'+esc(t)+'"')}).join('');
 var valueControl='';
 if(key.type==='auto_click'){
   var bits=String(key.value||'100|0|left').split('|'),ms=Number(bits[0]);if(!Number.isInteger(ms)||ms<20||ms>10000)ms=100;
   var repetitions=Number(bits[1]);if(!Number.isInteger(repetitions)||repetitions<0||repetitions>100000)repetitions=0;
   var hand=bits[2]==='right'?'right':'left';
   valueControl='<div class="pnw-autoclick"><label>클릭 간격 (20~10000ms)<input type="number" data-field="auto-ms" min="20" max="10000" value="'+ms+'"></label><label>반복 횟수 (0=무제한)<input type="number" data-field="auto-count" min="0" max="100000" value="'+repetitions+'"></label><label>클릭 버튼<select data-field="auto-button"><option value="left" '+(hand==='left'?'selected':'')+'>왼쪽</option><option value="right" '+(hand==='right'?'selected':'')+'>오른쪽</option></select></label></div>';
 }else if(NEEDS[key.type]){
   valueControl=(key.type==='text'||key.type==='copy_text')?'<textarea data-field="value" rows="3" maxlength="4096">'+esc(key.value)+'</textarea>':'<input data-field="value" maxlength="4096" value="'+esc(key.value)+'" placeholder="'+esc(PARAMS[key.type]||'설정 값')+'">';
 }else{valueControl='<p class="pnw-no-param">이 기능은 추가 입력이 필요하지 않습니다.</p>';}
 var extra=key.type==='shortcut'?'<label class="pnw-check"><input type="checkbox" data-field="toggle" '+(key.toggle?'checked':'')+'> 단축키 누름 유지 토글</label>':'';
 var color=COLORS.map(function(c){return button('color',key.color===c[0]?'✓':'','pnw-color '+(key.color===c[0]?'selected':''),'data-color="'+esc(c[0])+'" title="'+esc(c[0])+'"')}).join('');
 var right='<div class="pnw-edit-panel"><h4>키 '+String(state.selected+1).padStart(2,'0')+' 설정</h4><div class="pnw-current">'+icon(key.type)+'<div><strong>'+esc(key.label)+'</strong><small>'+esc(NAMES[key.type]||key.type)+'</small></div></div><h5>기능</h5><div class="pnw-action-grid">'+actions+'</div><div class="pnw-action-nav">'+button('action-page','◀ 이전','pnw-btn','data-index="'+(state.actionPage-1)+'" '+(state.actionPage===0?'disabled':''))+'<span>기능 '+(state.actionPage+1)+' / '+count+'</span>'+button('action-page','다음 ▶','pnw-btn','data-index="'+(state.actionPage+1)+'" '+(state.actionPage>=count-1?'disabled':''))+'</div><label class="pnw-field"><strong>버튼 이름</strong><input data-field="label" maxlength="45" value="'+esc(key.label)+'"></label><label class="pnw-field"><strong>'+esc(PARAMS[key.type]||'기능 설정')+'</strong>'+valueControl+'</label>'+extra+'<h5>키캡 색상</h5><div class="pnw-colors">'+color+'</div></div>';
 return '<div class="pnw-key-columns">'+panel+right+'</div>';
}
function drawProfiles(){
 var m=getModel();
 return '<div class="pnw-wide-card"><h4>저장된 프로필 · '+esc(activeTitle())+'</h4><p>실제 Windows 프로그램과 동일하게 모델별로 최대 8개 프로필을 저장합니다.</p><div class="pnw-profile-list">'+m.profiles.map(function(p,i){return btnProfile(p,i,m.active)}).join('')+'</div><label class="pnw-field"><strong>선택 프로필 이름</strong><input data-field="profile-name" maxlength="45" value="'+esc(m.profiles[m.active].name)+'"></label><div class="pnw-profile-actions">'+button('profile-add','+ 새 프로필','pnw-btn')+button('profile-delete','선택 프로필 삭제','pnw-btn')+'</div></div>';
}
function drawSystem(){
 return '<div class="pnw-wide-card"><h4>PN12 / PN15 장치 구성</h4><p class="pnw-device-label">'+esc(activeTitle())+' · ESP32-C3</p><p>실제 설정 파일에서 PN12·PN15 오른쪽·PN15 왼쪽 프로필을 별도로 보관합니다.</p><p><strong>장치 실시간 연결:</strong> Windows 프로그램에서 관리합니다. 브라우저에서 COM 포트를 연결했다고 표시하지 않습니다.</p><label class="pnw-field"><strong>선호 COM 포트 (선택 사항)</strong><input data-field="preferred-com" maxlength="32" placeholder="예: COM4" value="'+esc(state.config.preferredCOM)+'"></label><p class="pnw-system-tip">저장 위치: %APPDATA%\\PulseNoteDeck\\config_unified.json<br>설정을 파일에 적용하기 전 Windows 프로그램을 종료하세요.</p></div>';
}
function render(){
 var iconLogo=logo(),inner='<div class="pnw-desktop"><aside class="pnw-sidebar"><div class="pnw-brand">'+iconLogo+'<span><strong>PulseNote Deck</strong><small>CONTROL CENTER</small></span></div><p class="pnw-side-label">WORKSPACE</p>'+button('tab','⌨　 버튼 설정','pnw-nav '+(state.tab==='keys'?'selected':''),'data-tab="keys"')+button('tab','▤　 프로필 관리','pnw-nav '+(state.tab==='profiles'?'selected':''),'data-tab="profiles"')+'<p class="pnw-side-label">SYSTEM</p>'+button('tab','⚙　 장치 및 설정','pnw-nav '+(state.tab==='system'?'selected':''),'data-tab="system"')+'<div class="pnw-side-bottom"><div class="pnw-disconnected">●　오프라인 설정 편집</div><small>v7.5.22 · WINDOWS CONFIG v7</small></div></aside><div class="pnw-main"><div class="pnw-heading"><div><h3>'+esc(state.tab==='keys'?'버튼 설정':state.tab==='profiles'?'프로필 관리':'장치 및 설정')+'</h3><p>'+esc(state.tab==='keys'?'실제 Windows 설정 형식으로 버튼·색상·기능을 편집합니다.':state.tab==='profiles'?'모델별 프로필을 편집하고 저장합니다.':'연결·Windows 시작 옵션은 설치된 Windows 앱에서 관리합니다.')+'</p></div><span class="pnw-model-badge">'+esc(activeTitle())+'</span></div>'+(state.tab==='keys'?drawKeys():state.tab==='profiles'?drawProfiles():drawSystem())+'<div class="pnw-footer">●　 설정 파일 v7 호환 · USB 동작 및 Windows 기능 실행은 설치 프로그램에서 수행</div></div></div>';
 var toolbar='<div class="pnw-editor-toolbar"><span><strong>PulseNote Deck v7.5.22</strong> · 실제 설정 파일 편집</span><div class="pnw-editor-tools">'+button('import','설정 파일 불러오기','pnw-tool')+button('export','설정 JSON 다운로드','pnw-tool')+button('save-folder','실제 설정 폴더에 저장','pnw-tool pnw-tool-primary')+'</div></div>';
 mount.innerHTML='<section class="pnd-native-editor">'+toolbar+inner+'<div class="pnw-status" role="status" aria-live="polite">'+esc(state.status)+'</div><div class="pnw-reality-note"><b>실제 파일 연동:</b> Windows 앱과 동일한 <code>config_unified.json</code> 형식을 읽고 수정합니다. 파일로 저장한 변경은 Windows 앱 재실행 후 사용할 수 있습니다. 웹사이트에서 Windows 실행 파일을 직접 구동하거나 USB 단축키를 실행하지 않습니다. 브라우저 설정은 서버로 전송하지 않습니다.</div><input type="file" accept=".json,application/json" class="pnw-file-input" hidden></section>';
}
function status(msg){state.status=msg;var label=mount.querySelector('.pnw-status');if(label)label.textContent=msg;}
function mark(){remember();status('편집 내용 저장됨 · Windows 설정 파일로 적용하려면 파일 다운로드 또는 폴더 저장을 이용하세요.');}
function exportFile(){
 var text=JSON.stringify(normalize(state.config),null,2)+'\n';
 var blob=new Blob([text],{type:'application/json;charset=utf-8'}),url=URL.createObjectURL(blob);
 var a=document.createElement('a');a.href=url;a.download='config_unified.json';a.click();
 setTimeout(function(){URL.revokeObjectURL(url)},1000);status('실제 Windows 설정 형식의 config_unified.json 다운로드 완료');
}
async function writeFolder(){
 if(typeof window.showDirectoryPicker!=='function'){status('이 브라우저는 폴더 저장을 지원하지 않습니다. Chrome·Edge에서 이용하거나 JSON을 다운로드하세요.');return;}
 if(!confirm('Windows PulseNote Deck 프로그램을 종료했습니까?\n실제 config_unified.json 및 백업(.bak)을 수정합니다.'))return;
 try{
   var dir=await window.showDirectoryPicker({mode:'readwrite'}),fh=await dir.getFileHandle('config_unified.json',{create:true});
   var old=await (await fh.getFile()).text(),now=JSON.stringify(normalize(state.config),null,2)+'\n';
   if(old.trim()){
     var parsed=JSON.parse(old);
     if(!parsed||parsed.version!==7||!parsed.models)throw Error('선택한 폴더의 config_unified.json 형식이 다릅니다. 안전을 위해 덮어쓰지 않았습니다.');
     if(old!==now){
       var back=await dir.getFileHandle('config_unified.json.bak',{create:true}),bw=await back.createWritable();
       await bw.write(old);await bw.close();
     }
   }
   var writer=await fh.createWritable();await writer.write(now);await writer.close();
   status('실제 설정 파일 및 .bak 백업 저장 완료. Windows PulseNote Deck을 재실행해 적용하세요.');
 }catch(e){if(e.name==='AbortError'){status('폴더 선택 취소');return;}status('저장 실패: '+(e.message||e));}
}
function updateAuto(){
 var ms=mount.querySelector('[data-field="auto-ms"]'),ct=mount.querySelector('[data-field="auto-count"]'),bt=mount.querySelector('[data-field="auto-button"]');
 if(!ms||!ct||!bt)return;
 var a=Number(ms.value),b=Number(ct.value);
 if(!Number.isInteger(a)||a<20||a>10000||!Number.isInteger(b)||b<0||b>100000){status('오토클릭 간격 20~10000ms, 횟수 0~100000 범위를 지켜야 합니다.');return;}
 getKey().value=a+'|'+b+'|'+(bt.value==='right'?'right':'left');mark();
}
mount.addEventListener('error',function(ev){if(ev.target&&ev.target.tagName==='IMG')ev.target.hidden=true;},true);
mount.addEventListener('click',function(event){
 var b=event.target.closest('[data-pnw]');if(!b)return;
 var op=b.dataset.pnw;
 if(op==='model'){
   var id=b.dataset.model;state.config.selected=id==='12'?12:15;state.config.selectedSide=id==='15L'?'L':'R';state.selected=0;state.actionPage=0;mark();render();
 }else if(op==='tab'){state.tab=b.dataset.tab;render();
 }else if(op==='key'){state.selected=Number(b.dataset.index);render();
 }else if(op==='type'){
   var k=getKey(),t=b.dataset.type,old=k.type;
   var auto=String(k.label).startsWith('버튼 ')||k.label.replace(/\s+/g,'')===(NAMES[old]||old).replace(/\s+/g,'');
   if(old!==t){k.type=t;k.value=t==='auto_click'?'100|0|left':t==='audio_output'?'cycle':'';delete k.toggle;if(auto)k.label=NAMES[t]||t;}
   mark();render();
 }else if(op==='action-page'){var n=Number(b.dataset.index);if(n>=0&&n<Math.ceil(TYPES.length/9)){state.actionPage=n;render();}
 }else if(op==='color'){getKey().color=b.dataset.color;mark();render();
 }else if(op==='profile'){getModel().active=Number(b.dataset.index);state.selected=0;mark();render();
 }else if(op==='profile-add'){
   var m=getModel();if(m.profiles.length>=8){status('프로필은 모델당 최대 8개입니다.');return;}
   var p=JSON.parse(JSON.stringify(getProfile()));p.name='프로필 '+(m.profiles.length+1);m.profiles.push(p);m.active=m.profiles.length-1;state.selected=0;mark();render();
 }else if(op==='profile-delete'){
   var m=getModel();if(m.profiles.length<=1){status('기본 프로필 1개는 남겨야 합니다.');return;}
   if(confirm('선택한 프로필과 해당 키 설정을 삭제할까요?')){m.profiles.splice(m.active,1);m.active=0;state.selected=0;mark();render();}
 }else if(op==='import'){mount.querySelector('input[type=file]').click();
 }else if(op==='export'){exportFile();
 }else if(op==='save-folder'){writeFolder();}
});
mount.addEventListener('input',function(e){
 var field=e.target.dataset.field;if(!field)return;
 if(field==='label'){getKey().label=e.target.value.slice(0,45);var el=mount.querySelector('[data-keylabel="'+state.selected+'"]');if(el)el.textContent=getKey().label.slice(0,9);mark();}
 else if(field==='value'){getKey().value=e.target.value.slice(0,4096);mark();}
 else if(field==='profile-name'){getProfile().name=e.target.value||'기본 프로필';mark();}
 else if(field==='preferred-com'){state.config.preferredCOM=e.target.value.toUpperCase().trim();mark();}
 else if(field.startsWith('auto-'))updateAuto();
});
mount.addEventListener('change',function(e){
 if(e.target.dataset.field==='toggle'){if(e.target.checked)getKey().toggle=true;else delete getKey().toggle;mark();}
 if(e.target.dataset.field==='auto-button')updateAuto();
 if(e.target.matches('input[type=file]')){
   var file=e.target.files&&e.target.files[0];if(!file)return;
   file.text().then(function(raw){
     var cfg=normalize(JSON.parse(raw));state.config=cfg;state.selected=0;state.actionPage=0;remember();status('실제 Windows 설정 파일을 불러왔습니다: '+file.name);render();
   }).catch(function(err){status('설정 파일 읽기 실패: '+err.message)});
   e.target.value='';
 }
});
render();
window.PulseNoteWebConfig={getConfig:function(){return JSON.parse(JSON.stringify(state.config))},version:'7.5.22',configVersion:7};
})();