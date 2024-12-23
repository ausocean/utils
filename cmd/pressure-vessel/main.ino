/*
AUTHORS
  Saxon Nelson-Milton <saxon@ausocean.org>

LICENSE
  Copyright (C) 2020-2024 the Australian Ocean Lab (AusOcean)

  It is free software: you can redistribute it and/or modify them
  under the terms of the GNU General Public License as published by the
  Free Software Foundation, either version 3 of the License, or (at your
  option) any later version.

  It is distributed in the hope that it will be useful, but WITHOUT
  ANY WARRANTY; without even the implied warranty of MERCHANTABILITY or
  FITNESS FOR A PARTICULAR PURPOSE. See the GNU General Public License
  for more details.

  You should have received a copy of the GNU General Public License
  in gpl.txt.  If not, see http://www.gnu.org/licenses.
*/

// Display pins.
#define MAX7219DIN 3
#define MAX7219CS 4
#define MAX7219CLK 5

// Display parameters.
#define DISPLAY_BRIGHTNESS 15

// Pin numbers.
#define PRESSURE_PIN 0
#define RELAY_PIN 2
#define LED_PIN 13

// Pressures.
#define MAX_PRESSURE 200 //kPa
#define RANGE 30 // kPa
#define MIN_PRESSURE MAX_PRESSURE - RANGE
#define ABS_MAX_PRESSURE 300 // kPa

// Max pump running time.
#define MAX_PUMP_TIME 2 // Minutes

// Alarm consts.
#define MAX_PUMP_TIME_ALARM 1
#define ABS_MAX_PRESSURE_ALARM 2
#define ALARM_PULSE 1000  // milliseconds
#define ALARM_PERIOD 5000 // milliseconds

// Misc consts.
#define NO_OF_READINGS 500
#define RELAY_DELAY 1000 // milliseconds

// Forward declations
int median(byte*, unsigned int);
void bubbleSort(byte*, int);
int adjust(float);

// relay controls the relay with a delay.
void relay(int level) {
  digitalWrite(RELAY_PIN, level);
  delay(RELAY_DELAY);
}

// flashes produce n flashes of p millisecond pulse period
void flash(int n, int p) {
  while (n > 0) {
    digitalWrite(LED_PIN,HIGH);
    delay(p/2);
    digitalWrite(LED_PIN,LOW);
    delay(p/2);
    n--;
  }
}

void startPumpTimer(){
  startPumpTime = millis();
}

// getPumpTime returns the pump time in minutes.
float getPumpTime(){
  unsigned long now = millis();
  if (now < startPumpTime) {
    startPumpTime = 0;
    Serial.println("Time rolled over!");
  }
  return float(now-startPumpTime)/(1000.0*60.0);
}

int adjust(float p){
  if( p < 0 ){
    return 0;
  }
  return int(p);
}

bool pumpOn = false;
unsigned long startPumpTime = 0;
int alarmCount = 0; // Global to track alarm state entries.

void setPumpState(bool state) {
  if (state) {
    relay(HIGH);
    pumpOn = true;
    startPumpTimer();
    Serial.println("Pump turned ON.");
  } else {
    relay(LOW);
    pumpOn = false;
    Serial.println("Pump turned OFF.");
  }
}


// This will put arduino into an alarm state i.e. something
// went wrong.
void alarmed(int flashes) {
  unsigned long alarmStartTime = millis();
  alarmCount++; // Increment the alarm state counter.
  
  // Log that the system has entered the alarm state.
  Serial.print("Entering alarm state. Count: ");
  Serial.println(alarmCount);

  // Ensure pump is off for safety.
  relay(LOW);
  pumpOn = false;

  // Stay in alarm state until timeout AND pressure is below MAX_PRESSURE.
  while (true) {
    flash(flashes, ALARM_PULSE);
    delay(ALARM_PERIOD - (flashes * ALARM_PULSE));

    float currentPressure = read_pressure();

    // Check for timeout AND safe pressure.
    if ((millis() - alarmStartTime > 30000) && (currentPressure < MAX_PRESSURE)) {
      Serial.println("Pressure below MAX_PRESSURE. Exiting alarm state after timeout.");
      break;
    }

    // Log current pressure.
    Serial.print("Current pressure (kPa): ");
    Serial.println(currentPressure);
  }
}


float v_to_kPa(float v){
  float p = (((1600.0-0.0)/(4.5-0.5))*v)-(((1600.0-0.0)/(4.5-0.5))*0.5);
  return p;
}

float reading_to_v(unsigned int r){
  float v = float(r) * (5.0/1023.0);
  return v;
}

int read_pressure(){
  byte readings[NO_OF_READINGS];
  for(int i = 0; i < NO_OF_READINGS; i++ ){
    readings[i] = adjust(v_to_kPa(reading_to_v(analogRead(PRESSURE_PIN))));
  }
  return median(readings,NO_OF_READINGS);
}

int median(byte* nums, unsigned int n){
  bubbleSort(nums,n);
  int i =  n / 2;
  return nums[i];
}

void swap(byte *xp, byte *yp){
    byte temp = *xp;
    *xp = *yp;
    *yp = temp;
}

// An optimized version of Bubble Sort
void bubbleSort(byte arr[], int n){
   int i, j;
   bool swapped;
   for (i = 0; i < n-1; i++)
   {
     swapped = false;
     for (j = 0; j < n-i-1; j++)
     {
        if (arr[j] > arr[j+1])
        {
           swap(&arr[j], &arr[j+1]);
           swapped = true;
        }
     }

     // IF no two elements were swapped by inner loop, then break
     if (swapped == false)
        break;
   }
}

void setup() {
  pinMode(RELAY_PIN,OUTPUT);
  pinMode(LED_PIN,OUTPUT);

  MAX7219init();
  MAX7219brightness(DISPLAY_BRIGHTNESS);

  Serial.begin(9600);
  
  // Discard initial readings
  for (int i = 0; i < 10; i++) {
    float reading = read_pressure(); // Or whatever function gives you the pressure reading
    Serial.print("Discarding initial reading: ");
    Serial.println(reading);
  }
}

void loop() {
  // Get pressure from pressure sensor.
  float pressure = read_pressure();

  // Print pressure to display.
  MAX7219shownum(int(pressure));

  float pumpTime = 0;
  if( pumpOn ){
    pumpTime = getPumpTime();
  }

  Serial.print("Up time (minutes): ");
  Serial.println(millis()/1000/60);
  Serial.print("Pump time (minutes): ");
  Serial.println(pumpTime);
  Serial.print("Pressure (kPa): ");
  Serial.println(pressure);

  if( pumpTime > MAX_PUMP_TIME ){
    Serial.println("Pump has been running too long! Alarmed!");
    alarmed(MAX_PUMP_TIME_ALARM);
  }

  if( pressure > ABS_MAX_PRESSURE ){
    Serial.println("Pressure too high! Alarmed!");
    alarmed(ABS_MAX_PRESSURE_ALARM);
  }

  // If the pump is on and we're above max pressure, then turn it off.
  if( pumpOn && pressure > MAX_PRESSURE ){
    relay(LOW);
    pumpOn = false;
    MAX7219init();
  }

  // If pump is off but below min pressure, turn it on.
  if( !pumpOn && pressure < MIN_PRESSURE ){
    relay(HIGH);
    pumpOn = true;
    startPumpTimer();
    MAX7219init();
  }
}


void MAX7219shownum(unsigned long n){
  unsigned long k=n;
  byte blank=0;
  for(int i=1;i<9;i++){
    if(blank){
      MAX7219senddata(i,15);
    }else{
      MAX7219senddata(i,k%10);
    }
    k=k/10;
    if(k==0){blank=1;}
  }
}

void MAX7219brightness(byte b){  //0-15 is range high nybble is ignored
  MAX7219senddata(10,b);        //intensity
}

void MAX7219init(){
  pinMode(MAX7219DIN,OUTPUT);
  pinMode(MAX7219CS,OUTPUT);
  pinMode(MAX7219CLK,OUTPUT);
  digitalWrite(MAX7219CS,HIGH);   //CS off
  digitalWrite(MAX7219CLK,LOW);   //CLK low
  MAX7219senddata(15,0);        //test mode off
  MAX7219senddata(12,1);        //display on
  MAX7219senddata(9,255);       //decode all digits
  MAX7219senddata(11,7);        //scan all
  for(int i=1;i<9;i++){
    MAX7219senddata(i,0);       //blank all
  }
}

void MAX7219senddata(byte reg, byte data){
  digitalWrite(MAX7219CS,LOW);   //CS on
  for(int i=128;i>0;i=i>>1){
    if(i&reg){
      digitalWrite(MAX7219DIN,HIGH);
    }else{
      digitalWrite(MAX7219DIN,LOW);
    }
  digitalWrite(MAX7219CLK,HIGH);
  digitalWrite(MAX7219CLK,LOW);   //CLK toggle
  }
  for(int i=128;i>0;i=i>>1){
    if(i&data){
      digitalWrite(MAX7219DIN,HIGH);
    }else{
      digitalWrite(MAX7219DIN,LOW);
    }
  digitalWrite(MAX7219CLK,HIGH);
  digitalWrite(MAX7219CLK,LOW);   //CLK toggle
  }
  digitalWrite(MAX7219CS,HIGH);   //CS off
}
