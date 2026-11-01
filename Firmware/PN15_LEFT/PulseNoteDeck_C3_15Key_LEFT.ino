// PN15 GRIP LEFT hand firmware; pin wiring remains identical; visual positions are mirrored.
// PulseNote Deck PN15 GRIP -- ESP32-C3 Super Mini, 3 rows x 5 columns.
// Rows GPIO {0,1,6}; columns GPIO {3,4,21,7,5} (8 pins total).
// KEY 1..15 are row-major. USB CDC On Boot: Enabled; baud: 115200.
// No diodes: only ONE physical switch at a time, with full-release lockout.
#include <Arduino.h>
#include <string.h>

static const uint8_t ROWS[] = {0, 1, 6};
static const uint8_t COLS[] = {3, 4, 21, 7, 5};
static constexpr uint8_t ROW_COUNT = sizeof(ROWS)/sizeof(ROWS[0]);
static constexpr uint8_t COL_COUNT = sizeof(COLS)/sizeof(COLS[0]);
static constexpr uint32_t DEBOUNCE_MS = 25;

static int8_t candidate = -1;
static bool waitingForRelease = false;
static uint32_t candidateSince = 0, releaseSince = 0;
static bool releaseTiming = false;
static char serialLine[40];
static uint8_t serialLength = 0;
static bool usbSessionAnnounced = false;

static void identify() {
  Serial.println("PND:READY:2.1");
  Serial.println("PND:INFO:15:L:C3");
}
static void announceWhenUSBReady() {
  bool ready = (bool)Serial;
  if (!ready) { usbSessionAnnounced = false; return; }
  if (!usbSessionAnnounced) {
    usbSessionAnnounced = true;
    identify();
  }
}
static void pollCommands() {
  while (Serial.available()) {
    char c = (char)Serial.read();
    if (c == '\r') continue;
    if (c == '\n') {
      serialLine[serialLength] = '\0';
      if (strcmp(serialLine, "PND:HELLO") == 0) identify();
      serialLength = 0;
    } else if (serialLength + 1 < sizeof(serialLine)) {
      serialLine[serialLength++] = c;
    } else {
      serialLength = 0;
    }
  }
}
static int scanOnlyKey() {
  int found = -1;
  uint8_t count = 0;
  for (uint8_t col = 0; col < COL_COUNT; ++col) {
    // Other columns stay high-impedance, never driven HIGH.
    digitalWrite(COLS[col], LOW);
    pinMode(COLS[col], OUTPUT);
    delayMicroseconds(35);
    for (uint8_t row = 0; row < ROW_COUNT; ++row) {
      if (digitalRead(ROWS[row]) == LOW) {
        found = row * COL_COUNT + col;
        ++count;
      }
    }
    pinMode(COLS[col], INPUT);
  }
  return count > 1 ? -2 : (count == 1 ? found : -1);
}
void setup() {
  Serial.begin(115200);
  for (uint8_t r = 0; r < ROW_COUNT; ++r) pinMode(ROWS[r], INPUT_PULLUP);
  for (uint8_t c = 0; c < COL_COUNT; ++c) pinMode(COLS[c], INPUT);
  candidateSince = millis();
}
void loop() {
  announceWhenUSBReady();
  pollCommands();
  int key = scanOnlyKey();  // -1 idle; -2 multiple; 0..14 one switch
  uint32_t now = millis();
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
    // Avoid ghost transitions as multiple keys are released.
    waitingForRelease = true;
    releaseTiming = false;
  } else {
    if (key != candidate) { candidate = key; candidateSince = now; }
    if (key >= 0 && (uint32_t)(now - candidateSince) >= DEBOUNCE_MS) {
      Serial.print("KEY:");
      Serial.println(key + 1);
      waitingForRelease = true;
      releaseTiming = false;
    }
  }
  delay(2);
}
