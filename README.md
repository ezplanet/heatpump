# Viessman Vitocal 100A, Rinnai Shimanto, Thermocold MEX Vsx - Heatpump Telemetry

## Purpose
The purpose of this service is to make Viessmann Vitocal 100A, Rinnai Shimanto, Thermocold MEX Vsx series heatpumps telemetry available to IoT services and monitors 

## Description
This service reads the MODBUS communications between a Viessmann Remote Touch Controller (master) and a Vitocal 100A heatpump (slave) and posts heatpump data in json format to an MQTT topic.
A Remote Touch Controller (RTC) is essential because this service only reads MODBUS data, it does not query the heatpump directly. The Remote Touch Controller is the MODBUS master that initiates the communications with a Vitocal 100A heatpump and queries its telemetry.
It could be possible to reproduce the queries sent by the RTC, however this service was designed to be read only in order to avoid the risk of sending unwanted configuration changes to the heatpump. 

## Note
Viessmann, Rinnai, Thermocold MODBUS data packets are undocumented (The Manufacturers do not provide documentation), they have been decoded by observing the heatpump behaviour and patterns and thus in some cases they might be inaccurate or incorrect.

## Decoding
Four query/response records have been decoded, each are identified by their length 
### MODBUS Read Registers Records
A Vitocal 100A response record looks as follows:
```
+--------------------------------+
| AA FF SZ VVVV1 VVVV2 VVVx CKSM |
+--------------------------------+
AA = Modbus Address, 1 Byte (default = 1)
FF = Modbus Function, 1 Byte (read registers = 3)
SZ = Data registers size in bytes (1 register = 2 bytes), 1 Byte
VVV1 = Register 1
VVV2 = Register 2
VVVx = Register x
CKSM = Checksum (inverted bytes)
```
Four response types have been identified: STATES, MACHINE, TEMPERATURES, ERRORS.
The record type is indentified by its size which is fixed for each type
```
STATES       = 27 bytes
MACHINE      = 11 bytes
TEMPERATURES = 100 bytes
ERRORS       = 15 bytes
```
The data payload size for each type is its size less 5 bytes (1 address, 1 function, 1 payload size, 2 checksum), thus the validity of a response record is verified by checking that the value of the third byte (payload size) is equal the total record size less five.
#### STATES
Data payload is (27-5)/2 = 11 * 2 byte registers
```
BYTE   CONTENT
 0     Target address
 1     Modbus Function (3 = read registers) 
 2     Data registers size (registers are 2 bytes)
 3     Status: 3 = OFF
 4     Control Mode: 0 = OFF; 1 = ON; 2 = REMOTE/AUTO
 5
 6
 7     Mode heat/cool (heat - 0x40, cool = 0x80)

```

### JSON telemetry data

```
{
    "timestamp":"2022-11-14T11:45:19.454544965+01:00",
    "control_mode": 2,
    "status":0,
    "mode":1,
    "compressor_required":false,
    "compressor_status":0,
    "compressor_hz":0,
    "pump_status":0,
    "pump_speed":0,
    "fan_speed":0,
    "temperatures":{
        "water_in":"18.7",
        "water_out":"17.9",
        "external":"15.6",
        "compressor_in":"17.0"
        "compressor_out":"50.5"
    },
    "pressure_high":1189,
    "pressure_low":1201,
    "hours":17,
    "errors":{
        "error_1":0,
        "error_2":0,
        "error_3":0,
        "error_4":0,
        "error_5":0
    }
```



