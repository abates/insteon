# iSnoop Command

The isnoop command provides an Insteon sniffer that can be
injected between another application and the actual Insteon
PLM for debugging purposes.

The original intent for this tool was to be placed between
a PLM and an instance of Home Assistant.  My Home Assistant
instance was not always reflecting state changes of devices
in my Insteon network and I used isnoop to determine if the
network itself was dropping messages or if there was some
part of Home Assistant that was not receiving/processing
messages as expected.

## Usage

The tool is intended to be used with a tool such as socat.
isnoop will connect to the PLM serial port and then proxy
all data between the serial port and os.Stdin/os.Stdout.

Socat can execute isnoop and send the data to a newly created pts
that can then be connected to from the application being
snooped:

```shell
abates@localhost:~/socat -d -d exec:isnoop pty,raw,echo=0
2021/11/10 16:33:31 socat[1484514] N forking off child, using socket for reading and writing
2021/11/10 16:33:31 socat[1484514] N forked off child process 14515
2021/11/10 16:33:31 socat[1484514] N forked off child process 14515
2021/11/10 16:33:31 socat[1484515] N execvp'ing "isnoop"
2021/11/10 16:33:31 socat[1484514] N PTY is /dev/pts/6
2021/11/10 16:33:31 socat[1484514] N starting data transfer loop with FDs [5,5] and [7,7]
```

As an example, using the ic command via the above pts:

```shell
abates@hawkeye:~/local/devel/insteon.v2/cmd/ic$ ic -port /dev/pts/6 device 0a.0b.0c info
       Device: Dimmer (0a.0b.0c)
       Engine: I2Cs
     Category: 01.20
     Firmware: 69
Link Database:
    Flags Group Address    Data
    UR        1 04.05.06   00 1c 01
    AR        1 04.05.06   03 1c 01
    UR        1 01.02.03   fe 1c 01
    UC        1 01.02.03   03 1c 01
    UR        1 07.08.09   fe 1c 01
    UC        1 07.08.09   03 1c 01
```

You will see something similar to the following from the snoop/socat command:

```shell
ED 00.00.00 -> 0a.0b.0c 3:3 Read/Write ALDB Link Read 00.00 0
SD 0a.0b.0c -> 01.02.03 3:3 Read/Write ALDB ACK
ED 0a.0b.0c -> 01.02.03 1:1 Read/Write ALDB Link Resp 0f.ff 0 UR 1 07.08.09 0xfe 0x1c 0x01
ED 0a.0b.0c -> 01.02.03 1:1 Read/Write ALDB Link Resp 0f.f7 0 UC 1 07.08.09 0x03 0x1c 0x01
ED 0a.0b.0c -> 01.02.03 1:1 Read/Write ALDB Link Resp 0f.ef 0 UR 1 04.05.06 0x00 0x1c 0x01
ED 0a.0b.0c -> 01.02.03 1:1 Read/Write ALDB Link Resp 0f.e7 0 AR 1 04.05.06 0x03 0x1c 0x01
ED 0a.0b.0c -> 01.02.03 1:1 Read/Write ALDB Link Resp 0f.df 0 UR 1 01.02.03 0xfe 0x1c 0x01
ED 0a.0b.0c -> 01.02.03 1:1 Read/Write ALDB Link Resp 0f.d7 0 UC 1 01.02.03 0x03 0x1c 0x01
ED 0a.0b.0c -> 01.02.03 1:1 Read/Write ALDB Link Resp 0f.cf 0 AR 0 00.00.00 0x00 0x00 0x00
```
