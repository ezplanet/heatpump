#### TEMPERATURES

#### STATES

#### MACHINE
Record Size = 11 - Values Size: 6
```
 +--- Vitocal Modbus Address
 |  +--- MODBUS read = 3
 |  |  +--- Data Size = 6 (2 bytes words)
 |  |  |  +-- Compressor status +--+--- Checksum (2 bytes inverted)
 |  |  |  |                     |  |
 0  1  2  3  4  5  6  7  8  9 10 11
01    06

```
Compressor status: 0 = OFF; 1 = ON;  
ERRORS
