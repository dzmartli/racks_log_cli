# racks_log_cli

Simple CLI log viewer for deprecated Racks version

![](https://img.shields.io/badge/go-1.18-%230ea5e9) ![](https://img.shields.io/badge/mongo--driver-1.11.1-%230ea5e9)

```
Usage:
  -action string
    	entry action all|add|update|delete (default "all")
  -last uint
    	number of entries, has less priority than -range flag (default 100)
  -range string
    	range of dates in YYYY-MM-DD_YYYY-MM-DD format (default "2023-01-28_2023-01-29")
```