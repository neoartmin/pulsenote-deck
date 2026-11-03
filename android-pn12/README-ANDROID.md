# PulseNote PN12 Android USB

Android USB Host (CDC-ACM) application, no firmware modification required.

- Handshake: `PND:HELLO\n`, `PND:READY:2.1`, `PND:INFO:12:C3`
- Buttons: `KEY:1` … `KEY:12`
- Requires Android USB host and USB-C data connection.
- Windows firmware and software are unchanged.
- Built APK appears in Actions artifacts named PulseNote-PN12-Android-USB-APK.
- Initial build target Android 8.0+ (API 26), target API 35.

## 텍스트 매크로 (v1.3.3)

1. 장치 연결 없이 '버튼 설정'에서 PN12 / PN15 왼손 / PN15 오른손 탭 선택
2. 키 선택 → '텍스트 입력' 선택 → 이름과 넣을 문구 입력 (여러 줄 가능)
3. '장치 및 설정' → '텍스트 자동 입력 권한 설정' → 정보 확인 및 동의 → Android 접근성 설정에서 PulseNote 텍스트 매크로 서비스 활성화
4. 다른 앱에서 일반 텍스트 입력창을 선택한 다음 지정한 키 누르기

이 접근성 서비스는 **사용자가 지정한 물리 키를 누를 때** 현재 입력창의 기존 텍스트와 커서 위치를 읽고 글자를 삽입합니다. 자동으로 다른 키나 앱을 조작하지 않으며, 화면 내용 수집·저장·전송을 하지 않습니다. 암호 필드 및 PulseNote 앱 자체의 입력창은 차단합니다. 사용하지 않을 때는 Android 설정에서 끌 수 있습니다.

한글·영문·이모지·줄바꿈을 지원하나 대상 앱이 Android의 `ACTION_SET_TEXT`를 지원해야 합니다. 일부 게임/브라우저/특수 입력기에서는 입력이 제한됩니다. 서비스가 비활성화돼 있으면 텍스트 자동 삽입을 실행하지 않습니다.

**주의:** 앱의 키별 문구는 로컬 앱 설정에 평문으로 저장됩니다. 비밀번호·인증 토큰 등 민감한 텍스트는 등록하지 마세요. 해당 접근성 기능을 Google Play에 출시할 경우 별도 정책 신고·명시적 동의 등 심사가 필요합니다. 현재 APK는 GitHub Actions debug-signed 테스트 빌드입니다.

## Android v1.3.3 프로필 및 편집 단축키

- 프로필 관리의 '+ 새 프로필 만들기'는 선택된 프로필의 버튼 이름, 기능, 매크로 내용 및 색상을 PN12(12키)/PN15 L·R(각 15키)에 맞게 모두 복제합니다. 현재 프로필과 새 프로필은 이후 각각 별도로 저장됩니다. 최대 8개까지 지원합니다.
- 단축키 기능에서 `CTRL+C`, `CTRL+V`, `CTRL+X`, `CTRL+A`를 설정하면 사용자가 활성화한 접근성 서비스의 Android 편집 동작으로 실행합니다. 복사·잘라내기는 시스템 클립보드를 변경하며 붙여넣기는 그 내용을 사용합니다.
- **Windows와 동일한 범용 단축키 주입이 아닙니다.** 선택한 일반 편집 가능 입력창에서만 사용 가능하며, `CTRL+S`, `CTRL+Z`, 임의의 키 조합 또는 시스템 전역 단축키는 지원하지 않습니다. 대상 앱이 Android 접근성 액션을 지원해야 합니다.
- 줄바꿈 텍스트는 Android `ACTION_SET_TEXT`로 삽입합니다. 앱에 따라 줄바꿈이 손실되거나 이 동작 자체가 거부될 수 있으며 Windows `Shift+Enter`와 같은 물리 키 이벤트를 주입하지 않습니다. 가능한 경우 실제 결과를 재확인해 사용자에게 점검을 안내합니다.
- 기존 Android v1.3.2의 텍스트 매크로, 기존 저장 프로필, PN12/PN15 펌웨어 및 Windows 소스는 변경하지 않습니다.
- GitHub Actions에서 독립 JVM 회귀 테스트 및 Gradle 디버그 APK 컴파일을 수행합니다. Android 실기기 채팅앱/하드웨어 실측은 별도 확인이 필요합니다.
