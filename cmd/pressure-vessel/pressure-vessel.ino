/*
AUTHORS
  Saxon Nelson-Milton <saxon@ausocean.org>
  Alan Noble <alan@ausocean.org>
  Trek Hopton <trek@ausocean.org>

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
  in gpl.txt. If not, see http://www.gnu.org/licenses.
*/

/*
SAFETY NOTES
  Everything in this sketch is SOFTWARE protection. It does not replace:
    - a mechanical relief valve set above ABS_MAX_PRESSURE and below the
      rating of the weakest component under test,
    - a hardware pressure switch in series with the pump supply,
    - a flyback diode / snubber across the pump and relay coil.

  Alarms are LATCHED while the system is powered: the pump stays off and the
  alarm code flashes until the operator clears it. The simplest way to clear
  an alarm is to POWER CYCLE the system. The sketch always boots into normal
  operation (it never boots straight into an alarm state because of an earlier
  one). Alternatively send 'c' over serial (115200 baud) or, if CLEAR_PIN is
  enabled below, ground that pin. Serial/button clears are refused while the
  sensor is faulty or pressure is >= MAX_PRESSURE.

  The last alarm code is recorded in EEPROM purely for diagnostics and is
  printed over serial at boot. It is NEVER acted on. It is stored as a code
  plus its bitwise complement so stale data from other sketches is ignored.

  Because a power cycle clears an alarm (and the pump timers), a persistent
  fault will simply re-trip after restart. Hardware protections (relief
  valve, series pressure switch) are what protect against repeated cycling.

  Alarm flash codes (number of flashes per cycle):
    1 = pump ran too long continuously
    2 = absolute overpressure
    3 = sensor fault (reading outside valid 0.5-4.5 V band)
    4 = pressure not rising while pump is on (leak / pump or hose fault)
    5 = pump duty limit exceeded (cumulative on-time)
*/

#include <avr/wdt.h>
#include <EEPROM.h>

// ============================================================================
// Display Configuration (select exactly ONE; both use pins 3, 4 and 5)
// ============================================================================
#define DISP_74HC595
// #define MAX7219

#if defined(MAX7219) && defined(DISP_74HC595)
#error "Select only one display: MAX7219 or DISP_74HC595"
#endif

#ifdef MAX7219
#define MAX7219DIN 3
#define MAX7219CS  4
#define MAX7219CLK 5
#endif

#ifdef DISP_74HC595
#include <74HC595Plus.h>
HC595Plus display(3, 5, 4);
#endif

#define DISPLAY_BRIGHTNESS 15

// ============================================================================
// Hardware Pin Assignments
// ============================================================================
#define PRESSURE_PIN 0
#define RELAY_PIN    2
#define LED_PIN      13

// Relay polarity. Change both if your relay board is active-low.
#define RELAY_ON  HIGH
#define RELAY_OFF LOW

// Optional: ground this pin to clear a latched alarm (uses internal pull-up).
// Uncomment and pick a free pin if you want a physical clear button.
// #define CLEAR_PIN 7

// Set to 0 to disable the watchdog (e.g. if your bootloader doesn't support it).
// Very old Arduino bootloaders can get stuck in a reset loop with the watchdog.
#define USE_WATCHDOG 1

// ============================================================================
// Pressure Setpoints (20m water depth testing)
// 1m H2O ~ 10 kPa nominal (196.1 kPa exact fresh water, 201.0 kPa sea water)
// ============================================================================
#define MAX_PRESSURE     200                    // kPa target test pressure (~20m depth)
#define RANGE            30                     // kPa allowable lower pressure window
#define MIN_PRESSURE     (MAX_PRESSURE - RANGE) // kPa (170 kPa) - pump kicks in below this
#define ABS_MAX_PRESSURE 300                    // kPa SOFTWARE overpressure cutoff
                                                // (NOT a hardware cutoff - see notes above)

// ============================================================================
// Sensor Validity
// Transducer outputs 0.5-4.5 V. Anything well outside that band is treated as
// a sensor/wiring fault (open circuit, short to ground or supply).
// ADC counts = V * 1023 / 5. Check these against your transducer datasheet.
// ============================================================================
#define ADC_FAULT_LOW  82   // < 0.4 V
#define ADC_FAULT_HIGH 942  // > 4.6 V

// Startup: let the transducer settle, then allow a few retries before
// declaring a sensor fault at boot.
#define STARTUP_SETTLE_MS     2000
#define STARTUP_SENSOR_TRIES  10
#define STARTUP_RETRY_MS      500

// ============================================================================
// Timing & Limits
// ============================================================================
#define MAX_PUMP_TIME          5      // Minutes allowable for one continuous pump run
#define MAX_PUMP_TIME_ALARM    1
#define ABS_MAX_PRESSURE_ALARM 2
#define SENSOR_FAULT_ALARM     3
#define NO_RISE_ALARM          4
#define DUTY_LIMIT_ALARM       5
#define MAX_ALARM_CODE         5

// Pressure-rise check: while the pump runs, pressure must rise by at least
// MIN_RISE_KPA every RISE_WINDOW_S seconds. TUNE to your pump and volume.
#define RISE_WINDOW_S 30
#define MIN_RISE_KPA  5

// Duty limit: maximum cumulative pump-on time per (tumbling) window.
// TUNE to your test article's expected leak rate.
#define DUTY_WINDOW_MIN 60
#define MAX_PUMP_ON_MIN 20

#define ALARM_PULSE  1000 // Milliseconds per flash
#define ALARM_PERIOD 8000 // Milliseconds per alarm cycle (>= 5 flashes + a gap)

#define NO_OF_READINGS 50
#define RELAY_DELAY    1000 // Milliseconds to settle after switching relay

// EEPROM layout for the last-alarm diagnostic record: code at 0, complement at 1.
#define EE_ALARM_ADDR 0
#define EE_CHECK_ADDR 1

// ============================================================================
// Forward Declarations
// ============================================================================
int median(int*, unsigned int);
void bubbleSort(int*, int);
void swap(int*, int*);
int adjust(float);
float v_to_kPa(float);
float reading_to_v(unsigned int);
int readRawMedian();
bool readPressure(int&);
void forcePumpOff();
void setPumpState(bool);
void startPumpTimer();
float getPumpTime();
void alarmed(byte);
void flash(int, int);
void relay(int);
void serviceDelay(unsigned long);
bool clearRequested();
void resetRuntimeState();
byte readLastAlarm();
void recordAlarm(byte);
void displayPressure(float);
void displayError(byte);
void displayDashes();

#ifdef MAX7219
void MAX7219brightness(byte);
void MAX7219init();
void MAX7219senddata(byte, byte);
#endif

// ============================================================================
// Global State
// ============================================================================
bool pumpOn = false;
byte alarmCode = 0;               // non-zero while in an alarm (used by display refresh)
int alarmCount = 0;
unsigned long startPumpTime = 0;
unsigned long lastPrintTime = 0;

int lastRaw = 0;                  // raw ADC median from the most recent readPressure()
int lastPressure = 0;             // most recent valid reading (kPa)
int riseBaseline = 0;             // pressure at start of current rise window
unsigned long riseWindowStart = 0;

unsigned long lastLoopMs = 0;     // for duty accounting
unsigned long dutyWindowStart = 0;
unsigned long pumpOnMsInWindow = 0;

// ============================================================================
// Hardware Control Helpers
// ============================================================================

// Delay that keeps the watchdog fed and, during an alarm, keeps the error
// code on the display (needed if the display library is multiplexed).
void serviceDelay(unsigned long ms) {
  unsigned long t = millis();
  while (millis() - t < ms) {
#if USE_WATCHDOG
    wdt_reset();
#endif
    if (alarmCode) {
      displayError(alarmCode);
    }
    delay(2);
  }
}

// Immediately de-energise the pump with no delay. Safe to call repeatedly.
void forcePumpOff() {
  digitalWrite(RELAY_PIN, RELAY_OFF);
  pumpOn = false;
}

void relay(int level) {
  digitalWrite(RELAY_PIN, level);
  serviceDelay(RELAY_DELAY);
}

void flash(int n, int p) {
  while (n > 0) {
    digitalWrite(LED_PIN, HIGH);
    serviceDelay(p / 2);
    digitalWrite(LED_PIN, LOW);
    serviceDelay(p / 2);
    n--;
  }
}

void startPumpTimer() {
  startPumpTime = millis();
}

float getPumpTime() {
  unsigned long now = millis();
  // Unsigned subtraction correctly handles 49.7 day millis rollover
  return float(now - startPumpTime) / (1000.0f * 60.0f);
}

int adjust(float p) {
  if (p < 0.0f) {
    return 0;
  }
  return int(p);
}

void setPumpState(bool state) {
  if (state) {
    relay(RELAY_ON);
    pumpOn = true;
    startPumpTimer();
    riseBaseline = lastPressure;
    riseWindowStart = millis();
    Serial.println(F("Pump turned ON."));
  } else {
    relay(RELAY_OFF);
    pumpOn = false;
    Serial.println(F("Pump turned OFF."));
  }
}

void resetRuntimeState() {
  pumpOn = false;
  startPumpTime = 0;
  pumpOnMsInWindow = 0;
  dutyWindowStart = millis();
  lastLoopMs = millis();
  riseWindowStart = millis();
  riseBaseline = lastPressure;
}

// Returns true if the operator has asked for the alarm to be cleared.
bool clearRequested() {
  bool requested = false;
  while (Serial.available() > 0) {
    char c = Serial.read();
    if (c == 'c' || c == 'C') {
      requested = true;
    }
  }
#ifdef CLEAR_PIN
  if (digitalRead(CLEAR_PIN) == LOW) {
    requested = true;
  }
#endif
  return requested;
}

// ============================================================================
// Last-alarm diagnostic record
// Informational only: printed at boot, never used to decide behaviour.
// The code is stored with its bitwise complement so stale bytes from other
// sketches are not reported as a real alarm.
// ============================================================================
byte readLastAlarm() {
  byte code = EEPROM.read(EE_ALARM_ADDR);
  byte check = EEPROM.read(EE_CHECK_ADDR);
  if (code >= 1 && code <= MAX_ALARM_CODE && check == (byte)~code) {
    return code;
  }
  return 0;
}

void recordAlarm(byte code) {
  // update() only writes when the value differs, limiting EEPROM wear.
  EEPROM.update(EE_ALARM_ADDR, code);
  EEPROM.update(EE_CHECK_ADDR, (byte)~code);
}

// ============================================================================
// Alarm Handling
// Alarms are latched until power is cycled. This function only returns if the
// operator clears the alarm via serial/button AND the system is in a safe
// state. Callers should `return` from loop() afterwards so the next iteration
// starts with a fresh reading rather than a stale one.
// ============================================================================
void alarmed(byte code) {
  alarmCount++;
  alarmCode = code;

  forcePumpOff();
  recordAlarm(code);

  Serial.print(F("ALARM "));
  Serial.print(code);
  Serial.print(F(" latched (count "));
  Serial.print(alarmCount);
  Serial.println(F("). Power cycle (or send 'c') once the fault is fixed."));

  while (true) {
    forcePumpOff(); // keep re-asserting pump off
    flash(code, ALARM_PULSE);

    long pauseTime = (long)ALARM_PERIOD - ((long)code * ALARM_PULSE);
    if (pauseTime > 0) {
      serviceDelay(pauseTime);
    }

    int pressure = 0;
    bool sensorOk = readPressure(pressure);

    Serial.print(F("Alarm "));
    Serial.print(code);
    Serial.print(F(" - raw ADC: "));
    Serial.print(lastRaw);
    if (sensorOk) {
      Serial.print(F("  pressure (kPa): "));
      Serial.println(pressure);
    } else {
      Serial.println(F("  SENSOR FAULT"));
    }

    if (clearRequested()) {
      if (sensorOk && pressure < MAX_PRESSURE) {
        Serial.println(F("Alarm cleared by operator."));
        break;
      }
      Serial.println(F("Clear refused: sensor faulty or pressure still >= MAX_PRESSURE."));
    }
  }

  alarmCode = 0;
  resetRuntimeState();
}

// ============================================================================
// Sensor Processing & Math
// ============================================================================
float reading_to_v(unsigned int r) {
  return float(r) * (5.0f / 1023.0f);
}

float v_to_kPa(float v) {
  // Transducer: 0.5V to 4.5V linearly mapped across 0 to 1600 kPa
  // (assumes a GAUGE sensor and an ADC reference equal to the 5 V sensor supply)
  return (v - 0.5f) * (1600.0f / 4.0f);
}

// Median of raw ADC counts. Taking the median before conversion is equivalent
// to converting first (the conversion is monotonic) but lets us validate the
// raw value BEFORE negative pressures get clamped to zero.
int readRawMedian() {
  int readings[NO_OF_READINGS];
  for (int i = 0; i < NO_OF_READINGS; i++) {
    readings[i] = analogRead(PRESSURE_PIN);
  }
  return median(readings, NO_OF_READINGS);
}

// Returns false on sensor fault. On success, kpa holds the pressure.
// The raw median is always left in lastRaw for diagnostics.
bool readPressure(int &kpa) {
  int raw = readRawMedian();
  lastRaw = raw;
  if (raw < ADC_FAULT_LOW || raw > ADC_FAULT_HIGH) {
    kpa = 0;
    return false;
  }
  kpa = adjust(v_to_kPa(reading_to_v(raw)));
  return true;
}

int median(int* nums, unsigned int n) {
  bubbleSort(nums, n);
  return nums[n / 2];
}

void swap(int *xp, int *yp) {
  int temp = *xp;
  *xp = *yp;
  *yp = temp;
}

void bubbleSort(int arr[], int n) {
  int i, j;
  bool swapped;
  for (i = 0; i < n - 1; i++) {
    swapped = false;
    for (j = 0; j < n - i - 1; j++) {
      if (arr[j] > arr[j + 1]) {
        swap(&arr[j], &arr[j + 1]);
        swapped = true;
      }
    }
    if (!swapped) {
      break;
    }
  }
}

// ============================================================================
// Display Drivers
// ============================================================================
void displayPressure(float n) {
  #ifdef MAX7219
  unsigned long k = (unsigned long)n;
  byte blank = 0;
  for (int i = 1; i < 9; i++) {
    if (blank) {
      MAX7219senddata(i, 15);
    } else {
      MAX7219senddata(i, k % 10);
    }
    k = k / 10;
    if (k == 0) {
      blank = 1;
    }
  }
  #endif

  #ifdef DISP_74HC595
  // Zero decimal places ensures 3-digit values (e.g. 200) do not bleed into "kpa"
  display.displayNumber(n, 4, 0);
  display.displayTxt("kpa");
  #endif
}

// Shows "Err<code>" so an alarm is visible on the display, not just the LED.
// Verify your 74HC595Plus library renders "Err" correctly on your hardware.
void displayError(byte code) {
  #ifdef MAX7219
  // Decode mode B: 11 = 'E', 15 = blank. Digit 1 is the rightmost digit.
  MAX7219senddata(1, code % 10);
  MAX7219senddata(2, 11);
  for (int i = 3; i < 9; i++) {
    MAX7219senddata(i, 15);
  }
  #endif

  #ifdef DISP_74HC595
  char buf[6];
  snprintf(buf, sizeof(buf), "Err%d", code % 10);
  display.displayTxt(buf);
  #endif
}

void displayDashes() {
  #ifdef MAX7219
  MAX7219init();
  #endif
  #ifdef DISP_74HC595
  display.displayTxt("--------");
  #endif
}

#ifdef MAX7219
void MAX7219brightness(byte b) {
  MAX7219senddata(10, b);
}

void MAX7219init() {
  pinMode(MAX7219DIN, OUTPUT);
  pinMode(MAX7219CS, OUTPUT);
  pinMode(MAX7219CLK, OUTPUT);
  digitalWrite(MAX7219CS, HIGH);
  digitalWrite(MAX7219CLK, LOW);
  MAX7219senddata(15, 0);
  MAX7219senddata(12, 1);
  MAX7219senddata(9, 255);
  MAX7219senddata(11, 7);
  for (int i = 1; i < 9; i++) {
    MAX7219senddata(i, 0);
  }
}

void MAX7219senddata(byte reg, byte data) {
  digitalWrite(MAX7219CS, LOW);
  for (int i = 128; i > 0; i = i >> 1) {
    if (i & reg) {
      digitalWrite(MAX7219DIN, HIGH);
    } else {
      digitalWrite(MAX7219DIN, LOW);
    }
    digitalWrite(MAX7219CLK, HIGH);
    digitalWrite(MAX7219CLK, LOW);
  }
  for (int i = 128; i > 0; i = i >> 1) {
    if (i & data) {
      digitalWrite(MAX7219DIN, HIGH);
    } else {
      digitalWrite(MAX7219DIN, LOW);
    }
    digitalWrite(MAX7219CLK, HIGH);
    digitalWrite(MAX7219CLK, LOW);
  }
  digitalWrite(MAX7219CS, HIGH);
}
#endif

// ============================================================================
// Arduino Setup & Main Loop
// ============================================================================
void setup() {
#if USE_WATCHDOG
  wdt_disable(); // avoid a watchdog reset loop carrying over from a previous run
#endif

  // Drive the relay to its OFF level before making the pin an output so the
  // pump never glitches on at startup.
  digitalWrite(RELAY_PIN, RELAY_OFF);
  pinMode(RELAY_PIN, OUTPUT);
  pinMode(LED_PIN, OUTPUT);

#ifdef CLEAR_PIN
  pinMode(CLEAR_PIN, INPUT_PULLUP);
#endif

  #ifdef MAX7219
  MAX7219init();
  MAX7219brightness(DISPLAY_BRIGHTNESS);
  #endif

  #ifdef DISP_74HC595
  display.begin();
  #endif

  Serial.begin(115200);
  delay(1000);

#if USE_WATCHDOG
  wdt_enable(WDTO_4S);
#endif

  Serial.println(F("Pressure controller starting."));

  // Informational only: report the last alarm recorded before this boot.
  // A reboot always starts fresh; it never resumes an earlier alarm state.
  byte lastAlarm = readLastAlarm();
  if (lastAlarm != 0) {
    Serial.print(F("Note: last recorded alarm before this boot was "));
    Serial.println(lastAlarm);
  }

  // Let the transducer and supply settle before validating anything.
  serviceDelay(STARTUP_SETTLE_MS);

  // Discard initial readings during transducer stabilization
  for (int i = 0; i < 10; i++) {
#if USE_WATCHDOG
    wdt_reset();
#endif
    int p;
    bool ok = readPressure(p);
    Serial.print(F("Discarding initial reading - raw ADC: "));
    Serial.print(lastRaw);
    if (ok) {
      Serial.print(F("  kPa: "));
      Serial.println(p);
    } else {
      Serial.println(F("  (out of range)"));
    }
  }

  // Require a valid sensor reading before running, with a few retries to
  // tolerate a slow-starting transducer.
  int p = 0;
  bool sensorOk = false;
  for (int i = 0; i < STARTUP_SENSOR_TRIES; i++) {
    if (readPressure(p)) {
      sensorOk = true;
      break;
    }
    Serial.print(F("Startup sensor check failed, raw ADC: "));
    Serial.println(lastRaw);
    serviceDelay(STARTUP_RETRY_MS);
  }

  if (!sensorOk) {
    Serial.println(F("Sensor fault at startup! Alarmed!"));
    alarmed(SENSOR_FAULT_ALARM);
  } else {
    lastPressure = p;
  }

  resetRuntimeState();
}

void loop() {
#if USE_WATCHDOG
  wdt_reset();
#endif

  unsigned long now = millis();

  // --- Read and validate the sensor ---------------------------------------
  int pressure;
  if (!readPressure(pressure)) {
    Serial.print(F("Sensor fault! raw ADC: "));
    Serial.println(lastRaw);
    alarmed(SENSOR_FAULT_ALARM);
    return; // re-read from scratch rather than act on stale data
  }
  lastPressure = pressure;

  // --- Absolute overpressure (software) -----------------------------------
  if (pressure > ABS_MAX_PRESSURE) {
    Serial.println(F("Pressure too high! Alarmed!"));
    alarmed(ABS_MAX_PRESSURE_ALARM);
    return;
  }

  displayPressure(pressure);

  // --- Cumulative pump duty accounting ------------------------------------
  // Unsigned subtraction handles millis() rollover.
  if (pumpOn) {
    pumpOnMsInWindow += (now - lastLoopMs);
  }
  lastLoopMs = now;

  if (now - dutyWindowStart >= (unsigned long)DUTY_WINDOW_MIN * 60000UL) {
    dutyWindowStart = now;
    pumpOnMsInWindow = 0;
  }

  // --- Periodic serial logging every 1000 ms ------------------------------
  if (now - lastPrintTime >= 1000) {
    lastPrintTime = now;

    Serial.print(F("Up time (minutes): "));
    Serial.println(now / 1000 / 60);
    Serial.print(F("Pump time (minutes): "));
    Serial.println(pumpOn ? getPumpTime() : 0.0f);
    Serial.print(F("Pump on this window (minutes): "));
    Serial.println(float(pumpOnMsInWindow) / 60000.0f);
    Serial.print(F("Raw ADC: "));
    Serial.println(lastRaw);
    Serial.print(F("Pressure (kPa): "));
    Serial.println(pressure);
  }

  // --- Upper threshold reached: stop pumping ------------------------------
  if (pumpOn && pressure > MAX_PRESSURE) {
    setPumpState(false);
    displayDashes();
    return; // next iteration re-reads after the relay delay
  }

  // --- Pump health checks (only while pumping) ----------------------------
  if (pumpOn) {
    // Continuous run limit
    if (getPumpTime() > MAX_PUMP_TIME) {
      Serial.println(F("Pump has been running too long! Alarmed!"));
      alarmed(MAX_PUMP_TIME_ALARM);
      return;
    }

    // Pressure must be rising, otherwise there is a leak or pump/hose fault
    if (now - riseWindowStart >= (unsigned long)RISE_WINDOW_S * 1000UL) {
      if (pressure - riseBaseline < MIN_RISE_KPA) {
        Serial.println(F("Pressure not rising while pumping! Alarmed!"));
        alarmed(NO_RISE_ALARM);
        return;
      }
      riseBaseline = pressure;
      riseWindowStart = now;
    }

    // Cumulative on-time limit catches slow leaks that never trip the
    // per-run limit because each run is short.
    if (pumpOnMsInWindow > (unsigned long)MAX_PUMP_ON_MIN * 60000UL) {
      Serial.println(F("Pump duty limit exceeded! Alarmed!"));
      alarmed(DUTY_LIMIT_ALARM);
      return;
    }
  }

  // --- Lower threshold reached: start pumping -----------------------------
  if (!pumpOn && pressure < MIN_PRESSURE) {
    setPumpState(true);
    displayDashes();
  }
}
