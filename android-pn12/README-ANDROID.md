# PulseNote PN12 Android USB

Android USB Host (CDC-ACM) application, no firmware modification required.

- Handshake: `PND:HELLO\n`, `PND:READY:2.1`, `PND:INFO:12:C3`
- Buttons: `KEY:1` … `KEY:12`
- Requires Android USB host and USB-C data connection.
- Windows firmware and software are unchanged.
- Built APK appears in Actions artifacts named PulseNote-PN12-Android-USB-APK.
- Initial build target Android 8.0+ (API 26), target API 35.
