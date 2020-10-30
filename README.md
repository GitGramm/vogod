# :fire: vogod - V'mann optolink go daemon

`vogod` is a daemon which provides a high-level interface (currently a REST API) to a Viessmann® heating device via Optolink.

_This is unreleased, alpha quality software, not ready for deployment. Do not use, do not ingest, do not stare into beam._

## Usage
```
./vogod: V'mann optolink go daemon
    Build Date: 2020-10-30T15:38:36Z
    Build Version: v0.4.1

  -c string
        connection string, use socket://[host]:[port] for TCP or [serialDevice] for direct serial connection
  -cpuprofile file
        write cpu profile to file
  -d file
        filename of ecnDataPointType.xml like file (default "ecnDataPointType.xml")
  -e file
        filename of ecnEventType.xml like file (default "ecnEventType.xml")
  -memprofile file
        write memory profile to file
  -s string
        start http server at [bindtohost][:]port
  -v    verbose logging
```

![bildschirmfoto vom 2018-10-26 um 15 47 46](https://user-images.githubusercontent.com/1384994/47570842-6bcfa880-d937-11e8-973f-54bb8b14c9c1.png)
