# ADB GUI 🚀

**Wails**, **React**, **Ant Design**으로 구축된 강력하고 현대적인 독립형 Android 관리 도구입니다.

> ✨ **참고**: 이 애플리케이션은 순수한 **vibecoding**의 결과물입니다.

![License](https://img.shields.io/badge/license-MIT-blue.svg)
![Platform](https://img.shields.io/badge/platform-macOS%20%7C%20Windows%20%7C%20Linux-lightgrey.svg)

## ✨ 주요 기능

### 📱 장치 관리
- 연결된 장치의 실시간 모니터링.
- 장치 ID, 모델 및 연결 상태 확인.
- 앱, 쉘, Logcat 및 미러링에 대한 원클릭 액세스.

### 📦 앱 관리
- 설치된 모든 패키지(시스템 및 사용자 앱) 목록 표시.
- 이름 또는 유형별 앱 필터링 및 검색.
- **작업**: 강제 중지, 데이터 삭제, 활성화/비활성화 및 제거.
- **빠른 Logcat**: 앱 목록에서 특정 앱의 로그로 직접 이동.

### 🖥️ 화면 미러링 (Scrcpy)
- **내장 Scrcpy**: 외부 설치가 필요하지 않습니다.
- 세밀한 제어:
  - 비디오 비트레이트 및 최대 FPS.
  - 해상도(최대 크기).
  - 화면 켜짐 유지 및 화면 끄기 옵션.
  - 항상 위 창.
  - 오디오 스트리밍 토글.

### 📜 고급 Logcat
- 자동 스크롤 기능이 있는 실시간 로그 스트리밍.
- **앱별 필터링**: 특정 패키지 이름으로 로그 필터링.
- **자동 모니터링**: 앱이 열리기 전에 로깅을 시작합니다. 도구가 자동으로 PID를 감지하고 앱이 실행되면 필터링을 시작합니다.
- 키워드 검색/필터링.

### 💻 ADB 쉘
- ADB 명령 실행을 위한 통합 터미널.
- 출력 기록이 포함된 빠른 명령 실행.

---

## 🛠️ 내장 바이너리

이 애플리케이션은 완전히 독립적입니다. 다음을 포함합니다:
- **ADB** (Android Debug Bridge)
- **Scrcpy** 실행 파일
- **Scrcpy-server**

시작 시 이러한 파일은 임시 디렉토리에 추출되어 자동으로 사용됩니다. 시스템 PATH를 구성할 필요가 없습니다.

---

## ⚠️ Xiaomi/Poco/Redmi 사용자 주의 사항

Scrcpy에서 **터치 제어**를 활성화하려면 다음을 수행해야 합니다:
1. **개발자 옵션**으로 이동합니다.
2. **USB 디버깅**을 활성화합니다.
3. **USB 디버깅(보안 설정)**을 활성화합니다.
   *(참고: 대부분의 Xiaomi 장치에서는 SIM 카드와 Mi 계정 로그인이 필요합니다)*.

---

## 🚀 시작하기

### 필수 조건
- **Go** (v1.21)
- **Node.js** (v18 LTS)
- **Wails CLI** (v2.9.2)
  ```bash
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.9.2
  ```

### 개발
```bash
wails dev
```

### 빌드
```bash
wails build
```
컴파일된 애플리케이션은 `build/bin`에서 확인할 수 있습니다.

### 릴리스
이 프로젝트는 GitHub Actions를 사용하여 다중 플랫폼 빌드를 자동화합니다. 새 릴리스를 만들려면:
1. 커밋에 태그 달기: `git tag v1.0.0`
2. 태그 푸시: `git push origin v1.0.0`
GitHub Action은 macOS, Windows 및 Linux용으로 자동 빌드하고 아티팩트를 릴리스 페이지에 업로드합니다.

---

## 🔧 문제 해결

### macOS: "앱이 손상되었기 때문에 열 수 없습니다"
GitHub에서 앱을 다운로드할 때 *"adbGUI.app이(가) 손상되었기 때문에 열 수 없습니다"* 오류가 표시된다면, 이는 macOS Gatekeeper 격리 기능 때문입니다.

이 문제를 해결하려면 터미널에서 다음 명령어를 실행하세요:
```bash
sudo xattr -cr /path/to/adbGUI.app
```
*(`/path/to/adbGUI.app`을 실제 다운로드한 애플리케이션 경로로 변경하세요)*

> **직접 빌드하기:** Gatekeeper를 우회하고 싶지 않다면, 로컬에서 [소스 코드로 앱을 빌드](#-시작하기)할 수 있습니다. 몇 분이면 충분합니다!

### Windows: "Windows의 PC 보호"
파란색 SmartScreen 창이 앱 실행을 차단하는 경우:
1. **추가 정보 (More info)** 를 클릭하세요.
2. **실행 (Run anyway)** 을 클릭하세요.

---

## 📄 라이선스
이 프로젝트는 MIT 라이선스에 따라 라이선스가 부여됩니다.
