
# Known Insteon Commands

The following tables outline all the Insteon command combinations that are
known in this project.


## Standard Direct Commands

Command 1 | Command 2 | Name | Notes
----------|-----------|------|------
0x01|0x00|Assign to All-Link Group|
0x02|0x00|Delete from All-Link Group|
0x03|0x00|Product Data Request|
0x03|0x01|Fx Username Request|
0x03|0x02|Text String Request|
0x08|0x00|Exit Linking Mode|
0x09|0x00|Enter Linking Mode|
0x0a|0x00|Enter Unlinking Mode|
0x0d|0x00|Engine Version|
0x0f|0x00|Ping Request|
0x10|0x00|ID Request|
0x1f|0x00|Get Operating Flags|

## Extended Direct Commands

Command 1 | Command 2 | Name | Notes
----------|-----------|------|------
0x03|0x00|Product Data Response|
0x03|0x01|Fx Username Response|
0x03|0x02|Text String Response|
0x03|0x03|Set Text String|
0x03|0x04|Set All-Link Command Alias|
0x03|0x05|Set All-Link Command Alias Data|
0x08|0x00|Exit Linking Mode (i2cs)|Insteon version 2 with checksum devices only respond to extended linking commands
0x09|0x00|Enter Linking Mode (i2cs)|Insteon version 2 with checksum devices only respond to extended linking commands
0x0a|0x00|Enter Unlinking Mode (i2cs)|Insteon version 2 with checksum devices only respond to extended linking commands
0x20|0x00|Set Operating Flags|
0x2e|0x00|Extended Get/Set|
0x2f|0x00|Read/Write ALDB|

## All-Link Messages

Command 1 | Command 2 | Name | Notes
----------|-----------|------|------
0x06|0x00|All-link Success Report|
0x11|0x00|All-link recall|
0x12|0x00|All-link Alias 2 High|
0x13|0x00|All-link Alias 1 Low|
0x14|0x00|All-link Alias 2 Low|
0x15|0x00|All-link Alias 3 High|
0x16|0x00|All-link Alias 3 Low|
0x17|0x00|All-link Alias 4 High|
0x18|0x00|All-link Alias 4 Low|
0x21|0x00|All-link Alias 5|

## Standard Broadcast Messages

Command 1 | Command 2 | Name | Notes
----------|-----------|------|------
0x01|0x00|Set-button Pressed (responder)|
0x02|0x00|Set-button Pressed (controller)|
0x03|0x00|Test Powerline Phase|
0x04|0x00|Heartbeat|
0x27|0x00|Broadcast Status Change|

## Lighting Standard Direct Messages

Command 1 | Command 2 | Name | Notes
----------|-----------|------|------
0x11|0x00|Light On|
0x12|0x00|Light On Fast|
0x13|0x00|Light Off|
0x14|0x00|Light Off Fast|
0x15|0x00|Brighten Light|
0x16|0x00|Dim Light|
0x18|0x00|Manual Light Change Stop|
0x19|0x00|Status Request|
0x21|0x00|Light Instant Change|
0x22|0x01|Manual On|
0x23|0x01|Manual Off|
0x25|0x01|Set Button Tap|
0x25|0x02|Set Button Tap Twice|
0x27|0x00|Set Status|
0x2e|0x00|Light On At Ramp|This command is for dimmers with firmware version less than version 67
0x34|0x00|Light On At Ramp|Dimmers running firmware version 67 and higher
0x2f|0x00|Light Off At Ramp|
0x35|0x00|Light Off At Ramp|

## Dimmer Convenience Commands

Command 1 | Command 2 | Name | Notes
----------|-----------|------|------
0x17|0x01|Manual Start Brighten|
0x17|0x00|Manual Start Dim|
0x20|0x00|Enable Program Lock|
0x20|0x01|Disable Program Lock|
0x20|0x02|Enable Tx LED|
0x20|0x03|Disable Tx LED|
0x20|0x04|Enable Resume Dim|
0x20|0x05|Disable Resume Dim|
0x20|0x06|Enable Load Sense|
0x20|0x07|Disable Load Sense|
0x20|0x08|Disable Backlight|
0x20|0x09|Enable Backlight|
0x20|0x0a|Enable Key Beep|
0x20|0x0b|Disable Key Beep|

## Thermostat Standard Direct Messages

Command 1 | Command 2 | Name | Notes
----------|-----------|------|------
0x68|0x00|Decrease Temp|
0x69|0x00|Increase Temp|
0x6a|0x00|Get Zone Info|
0x6b|0x02|Get Mode|
0x6b|0x03|Get Ambient Temp|
0x6b|0x04|Set Heat|
0x6b|0x05|Set Cool|
0x6b|0x06|Set Auto|
0x6b|0x07|Turn Fan On|
0x6b|0x08|Turn Fan Off|
0x6b|0x09|Turn Thermostat Off|
0x6b|0x0a|Set Program Heat|
0x6b|0x0b|Set Program Cool|
0x6b|0x0c|Set Program Auto|
0x6b|0x0d|Get State|
0x6b|0x0e|Set State|
0x6b|0x0f|Get Temp Units|
0x6b|0x10|Set Units Fahrenheit|
0x6b|0x11|Set Units Celsius|
0x6b|0x12|Get Fan On-Speed|
0x6b|0x13|Set Fan-Speed Low|
0x6b|0x14|Set Fan-Speed Med|
0x6b|0x15|Set Fan-Speed High|
0x6b|0x16|Enable Status Change|
0x6b|0x17|Disable Status Change|
0x6c|0x00|Set Cool Set-Point|
0x6d|0x00|Set Heat Set-Point|

## Thermostat Extended Direct Messages

Command 1 | Command 2 | Name | Notes
----------|-----------|------|------
0x68|0x00|Increase Zone Temp|
0x69|0x00|Decrease Zone Temp|
0x6c|0x00|Set Zone Cool Set-Point|
0x6d|0x00|Set Zone Heat Set-Point|

