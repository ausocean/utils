#include <Wire.h>
#include <Adafruit_GFX.h>
#include <Adafruit_INA228.h>
#include <Adafruit_SSD1306.h>

Adafruit_INA228 ina228 = Adafruit_INA228();

const int INA228_SDA_PIN = 21;
const int INA228_SCL_PIN = 22;
const int OLED_SDA_PIN = 16;
const int OLED_SCL_PIN = 17;
const int BOOT_BUTTON_PIN = 0;

const int SCREEN_WIDTH = 128;
const int SCREEN_HEIGHT = 64;
const int OLED_RESET = -1;
const uint8_t OLED_ADDRESS = 0x3C;
const unsigned long SAMPLE_INTERVAL_MS = 1000;
const unsigned long BUTTON_DEBOUNCE_MS = 250;
const int PAGE_COUNT = 4;
const int GRAPH_POINTS = 112;

TwoWire OledWire = TwoWire(1);
Adafruit_SSD1306 display(SCREEN_WIDTH, SCREEN_HEIGHT, &OledWire, OLED_RESET);
bool displayReady = false;

int currentPage = 0;
bool lastButtonState = HIGH;
unsigned long lastButtonPressMs = 0;
unsigned long lastSampleMs = 0;

float lastVoltage_V = 0.0;
float lastCurrent_mA = 0.0;
float lastPower_W = 0.0;
float lastWattHours = 0.0;

float peakPower_W = 0.0;
float minPower_W = 0.0;
float averagePower_W = 0.0;
float powerSum_W = 0.0;
unsigned long sampleCount = 0;

unsigned long sessionStartMs = 0;

float powerGraph[GRAPH_POINTS];
int graphHead = 0;
int graphCount = 0;

void drawHeader(const char *title) {
  display.setTextSize(1);
  display.setTextColor(SSD1306_WHITE);
  display.setCursor(0, 0);
  display.print(title);
  display.setCursor(108, 0);
  display.print(currentPage + 1);
  display.print("/");
  display.print(PAGE_COUNT);
  display.drawLine(0, 10, SCREEN_WIDTH, 10, SSD1306_WHITE);
}

void drawMainView() {
  display.clearDisplay();
  drawHeader("Power Meter");

  display.setCursor(0, 18);
  display.print("Voltage: ");
  display.print(lastVoltage_V, 3);
  display.println(" V");

  display.setCursor(0, 32);
  display.print("Current: ");
  display.print(lastCurrent_mA, 2);
  display.println(" mA");

  display.setCursor(0, 46);
  display.print("Energy:  ");
  display.print(lastWattHours, 4);
  display.println(" Wh");

  display.display();
}

void drawStatsView() {
  display.clearDisplay();
  drawHeader("Stats");

  display.setCursor(0, 18);
  display.print("Peak P: ");
  display.print(peakPower_W, 3);
  display.println(" W");

  display.setCursor(0, 32);
  display.print("Avg P:  ");
  display.print(averagePower_W, 3);
  display.println(" W");

  display.setCursor(0, 46);
  display.print("Min P:  ");
  display.print(minPower_W, 3);
  display.println(" W");

  display.display();
}

void drawSessionView() {
  display.clearDisplay();
  drawHeader("Session");

  unsigned long elapsedSec = (millis() - sessionStartMs) / 1000;
  unsigned long h = elapsedSec / 3600;
  unsigned long m = (elapsedSec % 3600) / 60;
  unsigned long s = elapsedSec % 60;

  display.setCursor(0, 18);
  display.print("Total: ");
  display.print(lastWattHours, 4);
  display.println(" Wh");

  display.setCursor(0, 32);
  display.print("Time:  ");
  if (h < 10) display.print("0");
  display.print(h);
  display.print(":");
  if (m < 10) display.print("0");
  display.print(m);
  display.print(":");
  if (s < 10) display.print("0");
  display.println(s);

  display.setCursor(0, 46);
  display.print("Rate:  ");
  float elapsedHours = elapsedSec / 3600.0;
  float rateWh_hr = (elapsedHours > 0) ? (lastWattHours / elapsedHours) : 0.0;
  display.print(rateWh_hr, 3);
  display.println(" Wh/hr");

  display.display();
}

void drawGraphView() {
  display.clearDisplay();
  drawHeader("Power Graph");

  const int graphX = 0;
  const int graphY = 16;
  const int graphW = SCREEN_WIDTH;
  const int graphH = 34;

  float maxPower = 0.001;
  for (int i = 0; i < graphCount; i++) {
    int index = (graphHead - graphCount + i + GRAPH_POINTS) % GRAPH_POINTS;
    if (powerGraph[index] > maxPower) {
      maxPower = powerGraph[index];
    }
  }

  display.drawRect(graphX, graphY, graphW, graphH, SSD1306_WHITE);

  int previousX = -1;
  int previousY = -1;
  for (int i = 0; i < graphCount; i++) {
    int index = (graphHead - graphCount + i + GRAPH_POINTS) % GRAPH_POINTS;
    int x = graphX + i;
    int y = graphY + graphH - 1 - round((powerGraph[index] / maxPower) * (graphH - 2));
    y = constrain(y, graphY + 1, graphY + graphH - 2);

    if (previousX >= 0) {
      display.drawLine(previousX, previousY, x, y, SSD1306_WHITE);
    } else {
      display.drawPixel(x, y, SSD1306_WHITE);
    }
    previousX = x;
    previousY = y;
  }

  display.setCursor(0, 54);
  display.print("Max ");
  display.print(maxPower, 2);
  display.print("W  Now ");
  display.print(lastPower_W, 2);
  display.print("W");

  display.display();
}

void updateDisplay() {
  if (!displayReady) {
    return;
  }

  switch (currentPage) {
  case 0:
    drawMainView();
    break;
  case 1:
    drawStatsView();
    break;
  case 2:
    drawGraphView();
    break;
  case 3:
    drawSessionView();
    break;
  }
}

void handleBootButton() {
  bool buttonState = digitalRead(BOOT_BUTTON_PIN);
  unsigned long now = millis();

  if (lastButtonState == HIGH && buttonState == LOW &&
      now - lastButtonPressMs > BUTTON_DEBOUNCE_MS) {
    currentPage = (currentPage + 1) % PAGE_COUNT;
    lastButtonPressMs = now;
    updateDisplay();
  }

  lastButtonState = buttonState;
}

void addPowerGraphSample(float power_W) {
  powerGraph[graphHead] = power_W < 0.0 ? 0.0 : power_W;
  graphHead = (graphHead + 1) % GRAPH_POINTS;
  if (graphCount < GRAPH_POINTS) {
    graphCount++;
  }
}

void updateStats(float power_W) {
  if (sampleCount == 0) {
    peakPower_W = power_W;
    minPower_W = power_W;
  } else {
    if (power_W > peakPower_W) {
      peakPower_W = power_W;
    }
    if (power_W < minPower_W) {
      minPower_W = power_W;
    }
  }

  powerSum_W += power_W;
  sampleCount++;
  averagePower_W = powerSum_W / sampleCount;
}



void setup() {
  Serial.begin(115200);
  // Wait until serial port is opened
  while (!Serial) {
    delay(10);
  }

  sessionStartMs = millis();

  Wire.begin(INA228_SDA_PIN, INA228_SCL_PIN);
  OledWire.begin(OLED_SDA_PIN, OLED_SCL_PIN);
  pinMode(BOOT_BUTTON_PIN, INPUT_PULLUP);

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
  handleBootButton();

  unsigned long now = millis();
  if (now - lastSampleMs < SAMPLE_INTERVAL_MS) {
    delay(10);
    return;
  }
  lastSampleMs = now;

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
  float power_W = power_mW / 1000.0;

  lastVoltage_V = busVoltage_V;
  lastCurrent_mA = current_mA;
  lastPower_W = power_W;
  lastWattHours = wattHours;

  updateStats(power_W);
  addPowerGraphSample(power_W);

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

  updateDisplay();

  Serial.println();
}
