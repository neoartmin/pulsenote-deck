package main

// Every available action has a concise outcome statement AND a full help
// message. Keep documentation executable UI-side, not in a README only.
type actionGuide struct {
    summary string
    detail string
}
func guideForAction(kind string) actionGuide {
    switch kind {
    case "none":
        return actionGuide{"누르더라도 아무 작업도 하지 않습니다.",
            "비워 둘 키에 지정합니다. 다른 키나 프로필에는 영향을 주지 않습니다."}
    case "shortcut":
        return actionGuide{"단축키를 입력하거나 첫 누름~다음 누름 동안 키를 고정합니다.",
            "CTRL+S, CTRL+SHIFT+S, ALT+TAB처럼 조합을 지정합니다.\n"+
                "일반: 누를 때마다 키를 눌렀다 뗍니다.\n"+
                "누름 유지 토글: 첫 누름은 누른 상태로 고정, 두 번째 누름은 해제합니다. 상태 전환마다 Windows 알림이 1회 표시됩니다.\n"+
                "USB 연결 종료, 프로필 변경, 프로그램 종료 시 강제로 해제합니다. 문자 키는 고정 시 반복 입력될 수 있습니다."}
    case "text":
        return actionGuide{"현재 선택된 입력창에 등록한 문구를 타이핑합니다.",
            "한글·영문·여러 줄 텍스트를 자동 입력합니다. 비밀번호 창이나 관리자 권한 프로그램 등은 Windows 정책상 차단될 수 있습니다."}
    case "capture":
        return actionGuide{"Windows 영역 캡처 도구를 열고 선택한 부분을 복사합니다.",
            "Win+Shift+S를 호출합니다. 사용자가 영역을 선택해야 실제 이미지가 클립보드에 복사됩니다. 자동 PNG 저장 기능이 아닙니다."}
    case "copy_text":
        return actionGuide{"등록한 문구를 클립보드에 복사합니다.",
            "한글·이모지·줄바꿈을 포함한 문구를 복사하고, 원하는 곳에서 Ctrl+V로 붙여넣습니다. 기존 클립보드 내용은 바뀝니다."}
    case "system_settings":
        return actionGuide{"선택한 Windows 시스템 설정 화면을 엽니다.",
            "설정 선택에서 Wi-Fi·블루투스·화면·소리·절전·전체 설정을 지정합니다. 설정을 켜거나 끄지는 않습니다."}
    case "open_timers":
        return actionGuide{"이전 타이머 기능은 제거되었습니다.", "이전 설정은 보존되지만 실행되지 않습니다. 다른 기능으로 변경하세요."}
    case "open_alarms":
        return actionGuide{"Windows 시계 앱을 실행합니다.",
            "시계 앱에서 알람 탭을 직접 선택하고 알람을 설정합니다. 알람은 자동 생성되지 않습니다."}
    case "auto_click":
        return actionGuide{"버튼을 누르면 반복 클릭을 시작하고 다시 누르면 중지합니다.",
            "클릭 간격: 20~10000ms. 클릭 설정에서 왼쪽/오른쪽 클릭과 무제한/10/100/1000회를 지정합니다. F8 긴급 정지와 USB 연결 해제 및 종료 시 자동 중지를 지원합니다."}
    case "audio_output":
        return actionGuide{"스피커·헤드셋 등 기본 출력 장치를 전환합니다.",
            "출력 장치 선택에서 연결된 장치를 지정하거나 다음 장치 순환을 선택합니다. 상태 알림을 지원합니다. 특정 앱이 별도 출력 장치를 지정한 경우에는 변경되지 않을 수 있습니다."}
    case "program":
        return actionGuide{"선택한 프로그램이나 바로가기를 실행합니다.",
            "‘프로그램 선택’으로 설치된 앱을 찾거나 .exe/.lnk 경로를 입력합니다. Windows에서 실행 가능한 프로그램만 지원합니다."}
    case "website":
        return actionGuide{"지정한 웹사이트를 기본 브라우저로 엽니다.",
            "https:// 또는 http://로 시작하는 완전한 주소를 입력합니다. 검색이 목적이면 ‘웹 검색’을 사용하세요."}
    case "folder":
        return actionGuide{"지정한 Windows 폴더를 탐색기로 엽니다.",
            "C:\\Users\\...\\Documents 같은 실제 폴더 경로를 입력합니다. 폴더가 이동하거나 삭제되면 실행되지 않습니다."}
    case "volume_up":
        return actionGuide{"Windows 시스템 볼륨을 한 단계 올립니다.","Windows의 멀티미디어 볼륨 증가 키를 누릅니다. 특정 앱 전용 볼륨이 아닌 기본 장치의 볼륨입니다."}
    case "volume_down":
        return actionGuide{"Windows 시스템 볼륨을 한 단계 내립니다.","Windows의 멀티미디어 볼륨 감소 키를 누릅니다."}
    case "volume_mute":
        return actionGuide{"Windows 음소거 상태를 전환합니다.","한 번 누르면 음소거, 다시 누르면 음소거 해제됩니다. 별도 매크로 토글 설정이 필요하지 않습니다."}
    case "media_play":
        return actionGuide{"음악이나 영상의 재생·일시정지를 전환합니다.","Windows 미디어 재생/일시정지 키를 보냅니다. 실제 동작은 현재 미디어 앱의 지원 여부에 따라 다릅니다."}
    case "media_next":
        return actionGuide{"미디어 앱의 다음 트랙으로 이동합니다.","현재 재생 중인 음악·영상 앱이 다음 트랙 키를 지원해야 합니다."}
    case "media_prev":
        return actionGuide{"미디어 앱의 이전 트랙으로 이동합니다.","현재 재생 중인 음악·영상 앱이 이전 트랙 키를 지원해야 합니다."}
    case "web_search":
        return actionGuide{"선택한 검색 엔진에서 검색어를 검색합니다.",
            "‘검색 엔진’ 버튼에서 Google·네이버·YouTube를 고르고 검색어만 입력합니다. 기본 브라우저에서 검색 결과가 열립니다."}
    case "compose_email":
        return actionGuide{"수신자·제목·본문이 채워진 이메일 초안을 엽니다.",
            "‘이메일 설정’으로 받는 사람, 제목, 본문을 각각 입력합니다. 기본 메일 프로그램이 필요하며, 메시지를 자동 발송하지 않습니다."}
    case "qr_code":
        return actionGuide{"등록한 내용을 오프라인 QR코드 PNG로 생성합니다.",
            "내용을 입력하면 512px QR 이미지를 ‘사진/PulseNote Deck’ 폴더에 저장하고 기본 사진 뷰어를 엽니다. 외부 웹사이트에 텍스트를 전송하지 않습니다."}
    case "sequence":
        return actionGuide{"기존 연속 동작 설정입니다. 신규 등록은 숨겼습니다.",
            "별도 문법을 알아야 해 일반 기능 목록에서 제외했습니다. 이전 프로필에 저장된 연속 실행 데이터는 삭제하지 않으며 기존 설정은 실행할 수 있습니다."}
    case "clipboard_history":
        return actionGuide{"Windows 클립보드 기록 창을 엽니다.",
            "Win+V를 눌러 이전 복사 기록을 표시합니다. 클립보드 기록이 비활성화된 PC에서는 Windows 안내에 따라 직접 켜야 합니다."}
    case "window_control":
        return actionGuide{"현재 창·바탕화면·가상 데스크톱을 조작합니다.",
            "‘창 동작 선택’에서 최대화·최소화·좌우 정렬·바탕화면 표시·가상 데스크톱 좌우 이동 중 하나를 선택합니다."}
    default:
        return actionGuide{"이전 버전에서 지정한 기능입니다.",
            "현재 앱에서 지원하지 않는 저장된 기능입니다. 키의 기능을 다시 선택해 주세요."}
    }
}
