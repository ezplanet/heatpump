#### TEMPERATURES

#### STATES
Record Size = 27 - Values Size: 22
```
 +--- Vitocal Modbus Address
 |  +--- MODBUS read = 3
 |  |  +--- Data Size = 22 (2 bytes)    +--+--- Fans RPM
 |  |  |  +-- Compressor                |  |  +--+--- Compressor Hertz                                 
 |  |  |  |  +--- Control Mode          |  |  |  |  +--+--- Circulation Pump Speed % 
 |  |  |  |  |        +---Heat/Cool     |  |  |  |  |  |  +--+--- Hours     +--+--- Checksum (2 bytes inverted)   
 |  |  |  |  |        |                 |  |  |  |  |  |  |  |              |  |
 0  1  2  3  4  5  6  7  8  9 10 11 12 13 14 15 16 17 18 19 20 21 22 23 24 25 26
01 03 22 
```
Compressor Status: 1 = REQUIRED; 3 = STANDBY;

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
01 03 06              |  |
                      +--+--- Circulation Pump + Compressor           
```
Compressor status: 0 = OFF; 1 = RUNNING

Circulation Pump: 0x0000 = OFF; 0x0601 = ON; 

Compressor: 0x0000 = OFF; 0x8000 = ON


#### ERRORS
