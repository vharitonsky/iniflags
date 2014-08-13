Hybrid configuration library
============================

Combine standard go flags with your ini file.

Usage:

main.go
```go
package main

import (
	"flag"
	...
	"github.com/vharitonsky/iniflags"
	...
)

var (
	flag1 = flag.String("flag1", "default1", "defaultFlag1")
	...
	flagN = flag.Int("flagN", 123, "defaultFlagN")
)

func main() {
	iniflags.Parse()  // used instead of flag.Parse()
}
```

dev.ini

```ini
    flag1="val1"
    flag2=4
```

```bash

go run main.go -config dev.ini

```

Now all unset flags will get their value from .ini file provided in -config path. If value is not found in the .ini, flag will retain it's default value.

Flag value priority:
  - value set via command-line
  - value from ini file
  - default value

Iniflags is compatible with real .ini config files with [sections] and #comments. They will be skipped during parsing.
