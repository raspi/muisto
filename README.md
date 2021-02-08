# muisto

Memory dumper for Linux. Uses `/proc/<pid>/maps` file for source.

# Usage

```
muisto - process memory address space dumper v1.0.1 (2021-02-08T23:45:18+02:00)
(c) Pekka Järvinen 2021- [ https://github.com/raspi/muisto ]

Parameters:
  -maxaddress   Max address start offset (0 = no limit)   default: "0"
  -maxsize      Max address size (0 = no limit)   default: "100MiB"
  -minaddress   Min address start offset   default: "0"
  -minsize      Min address size   default: "1KiB"
  -pid          Program ID (PID)   default: "0"
  -stop         Stop process before dumping   default: "true"
  -version      Show version information   default: "false"

Examples:
  Dump addresses which has size between 8 KiB - 100 MiB at address offsets between 512 MiB - 1 GiB:
    ./muisto -pid 4321 -minaddress 512MiB -maxaddress 1GiB -maxsize 100MiB -minsize 8KiB

See:
  `man 5 proc`, `cat /proc/<pid>/maps`
```
