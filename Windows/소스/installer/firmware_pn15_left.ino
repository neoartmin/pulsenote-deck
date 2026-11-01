// PN15 GRIP LEFT-HAND version; only upload to left-hand device.
// PulseNote Deck PN15 GRIP / ESP32-C3 Super Mini
// 3 rows (0,1,6) x 5 columns (3,4,21,7,5); GPIO8 total.
// KEY 1..15 row-major; USB CDC On Boot must be enabled.
// No diodes: reject simultaneous presses, release-lockout to avoid ghost keys.
#include <Arduino.h>
#include <string.h>
static const uint8_t ROWS[]={0,1,6};
static const uint8_t COLS[]={3,4,21,7,5};
static constexpr uint32_t DEBOUNCE_MS=25;
static int8_t candidate=-1;
static bool waitingForRelease=false, releaseTiming=false, usbSessionAnnounced=false;
static uint32_t candidateSince=0, releaseSince=0;
static char serialLine[40];
static uint8_t serialLength=0;
static void identify(){
  Serial.println("PND:READY:2.1");
  Serial.println("PND:INFO:15:L:C3");
}
static void announceWhenUSBReady(){
  bool ready=(bool)Serial;
  if(!ready){usbSessionAnnounced=false;return;}
  if(!usbSessionAnnounced){usbSessionAnnounced=true;identify();}
}
static void pollCommands(){
  while(Serial.available()){
    char c=(char)Serial.read();
    if(c=='\r')continue;
    if(c=='\n'){
      serialLine[serialLength]='\0';
      if(strcmp(serialLine,"PND:HELLO")==0)identify();
      serialLength=0;
    }else if(static_cast<size_t>(serialLength)+1U<sizeof(serialLine))serialLine[serialLength++]=c;
    else serialLength=0;
  }
}
static int scanOnlyKey(){
  int found=-1;
  uint8_t count=0;
  for(uint8_t col=0;col<5;col++){
    digitalWrite(COLS[col],LOW);
    pinMode(COLS[col],OUTPUT);
    delayMicroseconds(35);
    for(uint8_t row=0;row<3;row++){
      if(digitalRead(ROWS[row])==LOW){found=row*5+col;count++;}
    }
    pinMode(COLS[col],INPUT);
  }
  return count>1?-2:(count==1?found:-1);
}
void setup(){
  Serial.begin(115200);
  for(uint8_t i=0;i<3;i++)pinMode(ROWS[i],INPUT_PULLUP);
  for(uint8_t i=0;i<5;i++)pinMode(COLS[i],INPUT);
  candidateSince=millis();
}
void loop(){
  announceWhenUSBReady();
  pollCommands();
  const int key=scanOnlyKey();
  const uint32_t now=millis();
  if(waitingForRelease){
    if(key==-1){
      if(!releaseTiming){releaseTiming=true;releaseSince=now;}
      if((uint32_t)(now-releaseSince)>=DEBOUNCE_MS){
        waitingForRelease=false;releaseTiming=false;candidate=-1;candidateSince=now;
      }
    }else releaseTiming=false;
  }else if(key==-2){
    waitingForRelease=true;releaseTiming=false;
  }else{
    if(key!=candidate){candidate=key;candidateSince=now;}
    if(key>=0&&(uint32_t)(now-candidateSince)>=DEBOUNCE_MS){
      Serial.print("KEY:");Serial.println(key+1);
      waitingForRelease=true;releaseTiming=false;
    }
  }
  delay(2);
}
