INA228 Power Sensor & ESP32-C3 Setup Guide
==========================================

This document outlines the hardware wiring, software configuration, and common troubleshooting steps for using the Adafruit INA228 power sensor with an ESP32-C3 development board in the Arduino IDE.

Hardware Requirements
---------------------
* Adafruit INA228 Breakout Board
* ESP32-C3 Development Board
* Data-capable USB cable
* DC Power Supply
* Test Load (e.g., 270-ohm resistor)


1. I2C Data Wiring (Sensor to Microcontroller)
----------------------------------------------
The INA228 communicates with the ESP32-C3 via the I2C protocol. Connect the logic pins as follows:

* VIN / VCC -> 3V3 (Power for the sensor chip)
* GND       -> GND (System ground)
* SDA       -> GPIO 8 (I2C Data Line)
* SCL       -> GPIO 9 (I2C Clock Line)


2. High-Side Sensing Wiring (Power Source to Load)
--------------------------------------------------
To safely and accurately measure power consumption, the INA228 must be wired in series between the power supply and your load, sharing a common ground.

1. Connect the Power Supply Negative (-) to a common ground point (like a twisted wire joint).
2. Connect the ESP32-C3 GND to that exact same common ground point.
3. Connect the Power Supply Positive (+) to the IN+ terminal on the INA228.
4. Add a small jumper wire connecting the IN+ terminal directly to the VBUS pin. 
5. Connect one leg of your load (e.g., 270-ohm resistor) to the IN- terminal.
6. Connect the other leg of your load to the common ground point.

WARNING: Never connect the power supply directly across IN+ and IN- without a load in place. The INA228's internal shunt resistor has near-zero resistance, and doing so will cause a short circuit.


3. Software Setup & Testing
---------------------------
1. Open the Arduino IDE.
2. Go to Sketch > Include Library > Manage Libraries, search for "Adafruit INA228", and install it along with all prompted dependencies.
3. Go to Tools > Board and select "ESP32 C3 Dev Module".
4. Go to Tools > Port and select your board's COM port.
5. Navigate to File > Examples > Adafruit INA228 Library and open the "ina228_test" sketch.
6. Click Upload.
7. Open the Serial Monitor and set the baud rate to 115200. You should see the voltage, current, and power readings begin to populate.


4. Troubleshooting
------------------
* Bus Voltage reads 0.00V: The VBUS pin is floating. Verify you have a physical jumper wire connecting the power source (usually IN+) to the VBUS pin.
* Serial Monitor shows gibberish/random characters: Your Serial Monitor's baud rate dropdown does not match the Serial.begin(115200); speed declared in the code.
* No COM port available: Ensure you are using a data-sync USB cable, not a charge-only cable. If the cable is good, you may need to download and install the CP210x or CH340 USB-to-UART drivers for your operating system.
* Current reading fluctuates slightly: Tiny fluctuations (e.g., 0.1 mA) are completely normal. This is caused by standard power supply ripple and thermal drift as your load resistor warms up.
