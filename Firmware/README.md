# PulseNote Deck — 펌웨어 기준 원본

이 디렉터리의 세 스케치만 **기준 펌웨어(Single Source of Truth)** 입니다.
Windows용/Android용으로 서로 다른 펌웨어를 개발하지 않습니다.

| 실제 하드웨어 | 기준 경로 | USB 응답 | 매트릭스 |
|---|---|---|---|
| PN12 | `PN12/PulseNoteDeck_C3_12Key.ino` | `PND:INFO:12:C3` | 행 0,1,6,7 / 열 3,4,21 |
| PN15 오른손 | `PN15/PulseNoteDeck_C3_15Key.ino` | `PND:INFO:15:R:C3` | 행 0,1,6 / 열 3,4,21,7,5 |
| PN15 왼손 | `PN15_LEFT/PulseNoteDeck_C3_15Key_LEFT.ino` | `PND:INFO:15:L:C3` | 행 0,1,6 / 열 3,4,21,7,5 |

PN15 좌/우 펌웨어는 **장치 식별 문자열과 첫 주석만 다릅니다**.
키 번호는 행 우선 01~15로 동일하며, 좌/우 대각선 UI 반전은 PC/Android 앱에서 처리합니다.
기존 PN15 식별자 `PND:INFO:15:C3`는 앱에서 오른손으로 취급합니다.

## 사용법

- ESP32-C3 Super Mini / ESP32C3 Dev Module
- Arduino IDE의 **USB CDC On Boot 활성화**, 시리얼 115200bps
- 각 하드웨어에 위 표의 해당 `.ino`만 업로드
- Windows 또는 Android 앱 설치 시 펌웨어가 ESP32에 자동 업로드되는 것은 **아님**
- 다이오드 없는 매트릭스이므로 **단일 키만 입력 가능** (같은 장치 내 다중 키 입력 차단)
- Windows/Android 앱은 다른 기종의 USB CDC 장치 세 개를 각각 독립 연결하는 구조

## 배포 복사본 관리

`Windows/펌웨어/`와 Windows 설치기 `Windows/소스/installer/firmware_*.ino`
파일은 이 디렉터리의 **배포용 바이트 복사본**입니다.

```sh
python tools/firmware_sync.py --check
python tools/firmware_sync.py --write
python tools/firmware_sync.py --check
```

`--check`는 원본과 여섯 복사본의 내용이 한 바이트라도 다르면 실패합니다.
Windows와 Android GitHub Actions는 모두 `--check`를 통과한 뒤 빌드합니다.
원본을 수정하면 `--write`로 복사본을 갱신해 **같은 커밋에 포함**해야 합니다.
Windows 설치기 컴파일은 이 복사본을 Go `embed`로 포함합니다.
Android 전체 파일집은 `Firmware/` 원본 자체를 포함합니다.

실제 ESP32-C3 업로드 및 세 장치 동시 연결 시험은 별도 검증이 필요합니다.
