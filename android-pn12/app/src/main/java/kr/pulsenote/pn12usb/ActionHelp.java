package kr.pulsenote.pn12usb;

/** Every current and legacy action has an explanation in the key editor. */
public final class ActionHelp {
    private ActionHelp(){}
    public static String summary(int kind){
        switch(kind){
            case 0:return "버튼을 눌러도 아무 작업을 하지 않습니다.";
            case 1:return "기기의 미디어 볼륨을 한 단계 올립니다.";
            case 2:return "기기의 미디어 볼륨을 한 단계 내립니다.";
            case 3:return "기기의 소리를 음소거하거나 해제합니다.";
            case 4:return "음악·동영상의 재생 상태를 전환합니다.";
            case 5:return "현재 미디어 앱의 다음 곡으로 이동합니다.";
            case 6:return "현재 미디어 앱의 이전 곡으로 이동합니다.";
            case 7:return "지정한 웹사이트를 기본 브라우저로 엽니다.";
            case 8:return "설치된 Android 앱을 실행합니다.";
            case 9:return "확장판에서 현재 입력창의 편집 동작을 실행합니다.";
            case 10:return "Android 문서 선택 화면으로 폴더를 엽니다.";
            case 11:return "화면 공유 최초 승인 후 반복 PNG 캡처가 가능합니다.";
            case 12:return "등록한 문구를 클립보드에 복사합니다.";
            case 13:return "선택한 Android 시스템 설정 화면을 엽니다.";
            case 14:return "시계 앱의 타이머 화면을 엽니다.";
            case 15:return "시계 앱의 알람 화면을 엽니다.";
            case 16:return "검색 엔진과 검색어를 선택해 브라우저로 검색합니다.";
            case 17:return "수신자·제목·본문이 채워진 메일 초안을 엽니다.";
            case 18:return "내용을 오프라인 QR 이미지로 보여줍니다.";
            case 19:return "이전 버전의 연속 동작입니다. 신규 등록은 숨겼습니다.";
            case 20:return "등록한 문구를 Android 공유 메뉴에 전달합니다.";
            default:return "이전 버전에서 지정된 기능입니다.";
        }
    }
    public static String detail(int kind){
        switch(kind){
            case 0:return "비워 둘 버튼에 지정합니다. 다른 버튼에는 영향을 주지 않습니다.";
            case 1:return "Android 미디어 음량을 증가시킵니다. 다른 앱의 개별 음량은 변경하지 않습니다.";
            case 2:return "Android 미디어 음량을 감소시킵니다.";
            case 3:return "Android 기기 음소거 상태를 전환합니다. OS 정책에 따라 일부 소리에는 적용되지 않을 수 있습니다.";
            case 4:return "재생 중인 미디어 앱에 재생·일시정지 명령을 전달합니다. 앱에서 해당 제어를 지원해야 합니다.";
            case 5:return "미디어 앱에 다음 트랙 명령을 보냅니다. 지원하지 않는 앱에서는 반응하지 않을 수 있습니다.";
            case 6:return "미디어 앱에 이전 트랙 명령을 보냅니다. 지원하지 않는 앱에서는 반응하지 않을 수 있습니다.";
            case 7:return "https:// 또는 http://로 시작하는 주소를 입력합니다. 단어 검색은 웹 검색 기능을 사용하세요.";
            case 8:return "‘설치된 앱에서 선택’을 눌러 앱을 고릅니다. 백그라운드 실행은 Android 정책에 따라 알림을 탭해야 할 수 있습니다.";
            case 9:return "확장판에서만 복사·붙여넣기·잘라내기·전체 선택을 지원합니다. 접근성 서비스 승인 필요. 일반판은 이 기능이 없습니다. Windows와 같은 물리적인 키 누름 유지는 제공되지 않습니다.";
            case 10:return "Android 폴더 선택기를 열어 직접 폴더를 고릅니다. 저장소 전체 접근 권한은 사용하지 않습니다.";
            case 11:return "화면 공유를 처음 한 번 승인하면 활성 캡처 세션 동안 키마다 PNG를 저장합니다. 알림에서 캡처 모드를 종료할 수 있으며 시스템이 공유를 중단하면 재승인해야 합니다.";
            case 12:return "여러 줄 문구를 클립보드에 저장합니다. 자동으로 붙여넣지 않으며 기존 클립보드의 내용이 바뀝니다.";
            case 13:return "설정 종류를 목록에서 선택합니다. 화면만 열며 Wi-Fi나 블루투스 상태를 임의로 변경하지 않습니다.";
            case 14:return "Android 시계 앱의 타이머 화면으로 이동합니다. 실제 시간 설정과 시작은 사용자가 직접 수행합니다.";
            case 15:return "Android 시계 앱의 알람 화면으로 이동합니다. 알람 생성·변경은 직접 수행합니다.";
            case 16:return "Google·네이버·YouTube 중 엔진을 선택하고 검색어만 입력합니다. 지정한 검색 결과를 기본 브라우저로 엽니다.";
            case 17:return "받는 사람·제목·본문을 별도로 작성합니다. 메일 앱의 작성 화면만 열며 자동 전송하지 않습니다.";
            case 18:return "UTF-8 900바이트 이하 내용을 앱 내부에서 QR로 생성합니다. 저장 위치는 사용자가 직접 선택하며 추가 카메라·저장소 권한이 필요하지 않습니다.";
            case 19:return "별도 문법을 알아야 하는 기능이라 일반 선택 목록에서 제외했습니다. 기존 설정을 보존하기 위해 데이터와 실행 경로는 유지합니다.";
            case 20:return "입력한 텍스트를 Android 공유 메뉴로 전달합니다. 실제 공유 대상 선택과 전송은 사용자가 수행합니다.";
            default:return "지원 여부가 불명확한 이전 기능입니다. 다른 기능을 선택해 주세요.";
        }
    }
}
