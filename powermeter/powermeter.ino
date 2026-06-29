#include <Wire.h>
#include <Adafruit_GFX.h>
#include <Adafruit_INA228.h>
#include <Adafruit_SSD1306.h>

Adafruit_INA228 ina228 = Adafruit_INA228();

const int INA228_SDA_PIN = 21;
const int INA228_SCL_PIN = 22;
const int OLED_SDA_PIN = 16;
const int OLED_SCL_PIN = 17;

const int SCREEN_WIDTH = 128;
const int SCREEN_HEIGHT = 64;
const int OLED_RESET = -1;
const uint8_t OLED_ADDRESS = 0x3C;

TwoWire OledWire = TwoWire(1);
Adafruit_SSD1306 display(SCREEN_WIDTH, SCREEN_HEIGHT, &OledWire, OLED_RESET);
bool displayReady = false;

void showOledReadings(float voltage, float current_mA, float wattHours) {
  if (!displayReady) {
    return;
  }

  display.clearDisplay();
  display.setTextColor(SSD1306_WHITE);

  display.setTextSize(1);
  display.setCursor(0, 0);
  display.println("Power Meter");
  display.drawLine(0, 10, SCREEN_WIDTH, 10, SSD1306_WHITE);

  display.setCursor(0, 18);
  display.print("Voltage: ");
  display.print(voltage, 3);
  display.println(" V");

  display.setCursor(0, 32);
  display.print("Current: ");
  display.print(current_mA, 2);
  display.println(" mA");

  display.setCursor(0, 46);
  display.print("Energy:  ");
  display.print(wattHours, 4);
  display.println(" Wh");

  display.display();
}

void setup() {
  Serial.begin(115200);
  // Wait until serial port is opened
  while (!Serial) {
    delay(10);
  }

  Wire.begin(INA228_SDA_PIN, INA228_SCL_PIN);
  OledWire.begin(OLED_SDA_PIN, OLED_SCL_PIN);

  Serial.println("Adafruit INA228 Test");

  displayReady = display.begin(SSD1306_SWITCHCAPVCC, OLED_ADDRESS);
  if (displayReady) {
    display.clearDisplay();
    display.setTextColor(SSD1306_WHITE);
    display.setTextSize(1);
    display.setCursor(0, 0);
    display.println("Power Meter");
    display.println("Starting...");
    display.display();
  } else {
    Serial.println("Couldn't find SSD1306 OLED");
  }

  if (!ina228.begin()) {
    Serial.println("Couldn't find INA228 chip");
    if (displayReady) {
      display.clearDisplay();
      display.setCursor(0, 0);
      display.println("INA228 not found");
      display.println("Check I2C wiring");
      display.display();
    }
    while (1)
      ;
  }
  Serial.println("Found INA228 chip");
  // set shunt resistance and max current
  ina228.setShunt(0.015, 10.0);

  ina228.setAveragingCount(INA228_COUNT_16);
  uint16_t counts[] = {1, 4, 16, 64, 128, 256, 512, 1024};
  Serial.print("Averaging counts: ");
  Serial.println(counts[ina228.getAveragingCount()]);

  // set the time over which to measure the current and bus voltage
  ina228.setVoltageConversionTime(INA228_TIME_150_us);
  Serial.print("Voltage conversion time: ");
  switch (ina228.getVoltageConversionTime()) {
  case INA228_TIME_50_us:
    Serial.print("50");
    break;
  case INA228_TIME_84_us:
    Serial.print("84");
    break;
  case INA228_TIME_150_us:
    Serial.print("150");
    break;
  case INA228_TIME_280_us:
    Serial.print("280");
    break;
  case INA228_TIME_540_us:
    Serial.print("540");
    break;
  case INA228_TIME_1052_us:
    Serial.print("1052");
    break;
  case INA228_TIME_2074_us:
    Serial.print("2074");
    break;
  case INA228_TIME_4120_us:
    Serial.print("4120");
    break;
  }
  Serial.println(" uS");

  ina228.setCurrentConversionTime(INA228_TIME_280_us);
  Serial.print("Current conversion time: ");
  switch (ina228.getCurrentConversionTime()) {
  case INA228_TIME_50_us:
    Serial.print("50");
    break;
  case INA228_TIME_84_us:
    Serial.print("84");
    break;
  case INA228_TIME_150_us:
    Serial.print("150");
    break;
  case INA228_TIME_280_us:
    Serial.print("280");
    break;
  case INA228_TIME_540_us:
    Serial.print("540");
    break;
  case INA228_TIME_1052_us:
    Serial.print("1052");
    break;
  case INA228_TIME_2074_us:
    Serial.print("2074");
    break;
  case INA228_TIME_4120_us:
    Serial.print("4120");
    break;
  }
  Serial.println(" uS");

  // default polarity for the alert is low on ready, but
  // it can be inverted!
  // ina228.setAlertPolarity(1);
}

void loop() {

  // by default the sensor does continuous reading, but
  // we can set to triggered mode. to do that, we have to set
  // the mode to trigger a new reading, then wait for a conversion
  // either by checking the ALERT pin or reading the ready register
  // ina228.setMode(INA228_MODE_TRIGGERED);
  // while (!ina228.conversionReady())
  //  delay(1);

  float current_mA = ina228.getCurrent_mA();
  float busVoltage_V = ina228.getBusVoltage_V();
  float shuntVoltage_mV = ina228.getShuntVoltage_mV();
  float power_mW = ina228.getPower_mW();
  float energy_J = ina228.readEnergy();
  float charge_C = ina228.readCharge();
  float dieTemp_C = ina228.readDieTemp();
  float wattHours = energy_J / 3600.0;

  Serial.print("Current: ");
  Serial.print(current_mA);
  Serial.println(" mA");

  Serial.print("Bus Voltage: ");
  Serial.print(busVoltage_V);
  Serial.println(" V");

  Serial.print("Shunt Voltage: ");
  Serial.print(shuntVoltage_mV);
  Serial.println(" mV");

  Serial.print("Power: ");
  Serial.print(power_mW);
  Serial.println(" mW");

  Serial.print("Energy: ");
  Serial.print(energy_J);
  Serial.println(" J");

  Serial.print("Energy: ");
  Serial.print(wattHours, 6);
  Serial.println(" Wh");
  
  Serial.print("Charge: ");
  Serial.print(charge_C);
  Serial.println(" C");

  Serial.print("Temperature: ");
  Serial.print(dieTemp_C);
  Serial.println(" *C");

  showOledReadings(busVoltage_V, current_mA, wattHours);

  Serial.println();
  delay(1000);
}
