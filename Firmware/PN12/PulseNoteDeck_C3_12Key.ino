// PulseNote Deck v7.4 - ESP32-C3 Super Mini, 12 physical keys (3 columns x 4 rows)
// 12 keys on 7 GPIOs without diodes: SINGLE KEY ONLY (multi-key ghosting excluded).
// ROW0..3: GPIO {0,1,6,7}; COL0..2: GPIO {3,4,21}
// Each key joins one ROW to one COLUMN. NO switch wired directly to GND!
// KEY1 = row0/col0, KEY2 = row0/col1, KEY3 = row0/col2 ... KEY12 = row3/col2.
// Arduino IDE: ESP32C3 Dev Module, USB CDC On Boot: Enabled, Serial 115200.
// Existing 8-key direct-wired ESP32-C3 firmware is a different sketch.
#include <Arduino.h>
#include <string.h>

static const uint8_t ROWS[4] = {0,1,6,7};
static const uint8_t COLS[3] = {3,4,21};
static constexpr uint8_t KEY_COUNT = 12;
static constexpr uint32_t DEBOUNCE_MS = 25;

// State machine: one event per full press/release cycle. If more than one
// switch is observed, lock all inputs until the whole matrix is released.
static int8_t candidate = -1;
static bool waitingForRelease = false;
static uint32_t candidateSince = 0;
static uint32_t releaseSince = 0;
static bool releaseTiming = false;
static char line[40];
static uint8_t lineLength = 0;

static void identify() {
  Serial.println("PND:READY:2.1");
  Serial.println("PND:INFO:12:C3");
}
// Respond both when a host first opens USB CDC and when it sends HELLO.
// This avoids a race between device boot and the Windows app's first probe.
static bool usbSessionAnnounced = false;
static void announceWhenUSBReady() {
  const bool ready = (bool)Serial;
  if (!ready) { usbSessionAnnounced = false; return; }
  if (!usbSessionAnnounced) { usbSessionAnnounced = true; identify(); }
}
static void pollCommands() {
  while (Serial.available() > 0) {
    const char c = (char)Serial.read();
    if (c == '\r') continue;
    if (c == '\n') {
      line[lineLength] = '\0';
      if (strcmp(line, "PND:HELLO") == 0) identify();
      lineLength = 0;
    } else if (static_cast<size_t>(lineLength) + 1 < sizeof(line)) {
      line[lineLength++] = c;
    } else {
      lineLength = 0;
    }
  }
}

static int scanOnlyKey() {
  int candidate = -1;
  uint8_t count = 0;
  for (uint8_t col = 0; col < 3; ++col) {
    // All other columns remain floating/high-impedance. Never drive them HIGH
    // against another active column through simultaneous key presses.
    digitalWrite(COLS[col], LOW);
    pinMode(COLS[col], OUTPUT);
    delayMicroseconds(35);
    for (uint8_t row = 0; row < 4; ++row) {
      if (digitalRead(ROWS[row]) == LOW) {
        candidate = row * 3 + col;
        ++count;
      }
    }
    pinMode(COLS[col], INPUT);
  }
  // More than one held key is deliberately ignored (no diodes, no NKRO).
  return count > 1 ? -2 : (count == 1 ? candidate : -1);
}
void setup() {
  Serial.begin(115200);
  for (uint8_t row = 0; row < 4; ++row) pinMode(ROWS[row], INPUT_PULLUP);
  for (uint8_t col = 0; col < 3; ++col) pinMode(COLS[col], INPUT);
  candidateSince = millis();
}
void loop() {
  announceWhenUSBReady();
  pollCommands();
  const int key = scanOnlyKey(); // -1: idle; -2: ambiguous; 0..11: one key
  const uint32_t now = millis();

  if (waitingForRelease) {
    if (key == -1) {
      if (!releaseTiming) { releaseTiming = true; releaseSince = now; }
      if ((uint32_t)(now - releaseSince) >= DEBOUNCE_MS) {
        waitingForRelease = false;
        releaseTiming = false;
        candidate = -1;
        candidateSince = now;
      }
    } else {
      releaseTiming = false;
    }
  } else if (key == -2) {
    // A multi-key/ghost event must not turn into a different key action as
    // a finger lifts. Re-arm only after a complete stable release.
    waitingForRelease = true;
    releaseTiming = false;
  } else {
    if (key != candidate) { candidate = key; candidateSince = now; }
    if (key >= 0 && (uint32_t)(now - candidateSince) >= DEBOUNCE_MS) {
      Serial.print("KEY:"); Serial.println(key + 1);
      waitingForRelease = true;
      releaseTiming = false;
    }
  }
  delay(2);
}
