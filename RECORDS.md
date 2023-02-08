#### TEMPERATURES

#### STATES
Record Size = 27 - Values Size: 22
```
 +--- Vitocal Modbus Address
 |  +--- MODBUS read = 3
 |  |  +--- Data Size = 22 (2 bytes)    +--+--- Compressor speed
 |  |  |  +-- Compressor                |  |  +--+--- Fans RPM                                 
 |  |  |  |  +--- Control Mode          |  |  |  |  +--+--- Circulation Pump Speed % 
 |  |  |  |  |        +---Heat/Cool     |  |  |  |  |  |  +--+--- Hours     +--+--- Checksum (2 bytes inverted)   
 |  |  |  |  |        |                 |  |  |  |  |  |  |  |              |  |
 0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26
01 03 22 
```
Compressor Status: 0x10 = REQUIRED; 0x02 = STANDBY; 0x30 = DEFROST STARTING; 0x50 DEFROST ACTIVE
X X X X X X X X
  | | |     | |      
  | | |     | +- Always ON
  | | |     +- 0 = Compressor required; 1 = Compressor OFF
  | | +- 1 = Compressor off
  | +- 1 = Defrost starting
  +- 1 = Defrost Active

Mode: 0x00 = OFF; 0x01 = COOL; 0x02 = HEAT

#### MACHINE
Record Size = 11 - Values Size: 6
```
 +--- Vitocal Modbus Address
 |  +--- MODBUS read = 3
 |  |  +--- Data Size = 6 (2 bytes words)
 |  |  |  +-- Comp. status  +--+--- Checksum (2 bytes inverted)
 |  |  |  |                 |  | 
 0  1  2  3  4  5  6  7  8  9 10
01 03 06     |        |  |
             |        +--+--- Circulation Pump + Compressor
             +--- Circulation Pump + Compressor status           
```
Compressor status: 0x00 = OFF; 0x01 = RUNNING; 0x10 = ALWAYS; 0x80 OIL HEATER ON
Byte 3 bitmap:
```
X X X X X X X X
|       |     |      
|       |     +- Compressor ON 
|       + Always ON
|
+- Oil Heater ON
```
Circulation Pump + Compressor status: 0x40 = Circulation Pump ON; 0x04 Compressor running; ??? = 0x08
Byte 4 bitmap:
```
X X X X X X X X
  |     | |
  |     | +- Compressor running
  |     +- ON when running sometimes. What is it?  
  +- Circulation Pump ON
```
Circulation Pump: 0x0000 = OFF; 0x0601 = ON; 

Compressor: 0x0000 = OFF; 0x8000 = ON


#### ERRORS
