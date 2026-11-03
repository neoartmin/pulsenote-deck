# PulseNote PN12 Android USB

Android USB Host (CDC-ACM) application, no firmware modification required.

- Handshake: `PND:HELLO\n`, `PND:READY:2.1`, `PND:INFO:12:C3`
- Buttons: `KEY:1` … `KEY:12`
- Requires Android USB host and USB-C data connection.
- Windows firmware and software are unchanged.
- Built APK appears in Actions artifacts named PulseNote-PN12-Android-USB-APK.
- Initial build target Android 8.0+ (API 26), target API 35.

## 텍스트 매크로 (v1.3.2)

1. 장치 연결 없이 '버튼 설정'에서 PN12 / PN15 왼손 / PN15 오른손 탭 선택
2. 키 선택 → '텍스트 입력' 선택 → 이름과 넣을 문구 입력 (여러 줄 가능)
3. '장치 및 설정' → '텍스트 자동 입력 권한 설정' → 정보 확인 및 동의 → Android 접근성 설정에서 PulseNote 텍스트 매크로 서비스 활성화
4. 다른 앱에서 일반 텍스트 입력창을 선택한 다음 지정한 키 누르기

이 접근성 서비스는 **사용자가 지정한 물리 키를 누를 때** 현재 입력창의 기존 텍스트와 커서 위치를 읽고 글자를 삽입합니다. 자동으로 다른 키나 앱을 조작하지 않으며, 화면 내용 수집·저장·전송을 하지 않습니다. 암호 필드 및 PulseNote 앱 자체의 입력창은 차단합니다. 사용하지 않을 때는 Android 설정에서 끌 수 있습니다.

한글·영문·이모지·줄바꿈을 지원하나 대상 앱이 Android의 `ACTION_SET_TEXT`를 지원해야 합니다. 일부 게임/브라우저/특수 입력기에서는 입력이 제한됩니다. 서비스가 비활성화돼 있으면 텍스트 자동 삽입을 실행하지 않습니다.

**주의:** 앱의 키별 문구는 로컬 앱 설정에 평문으로 저장됩니다. 비밀번호·인증 토큰 등 민감한 텍스트는 등록하지 마세요. 해당 접근성 기능을 Google Play에 출시할 경우 별도 정책 신고·명시적 동의 등 심사가 필요합니다. 현재 APK는 GitHub Actions debug-signed 테스트 빌드입니다.
