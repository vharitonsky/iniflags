Hybrid configuration library
============================

Combine standard go flags with ini files.

Usage:

```bash

go get -u -a github.com/vharitonsky/iniflags
```

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
	flag1 = flag.String("flag1", "default1", "Description1")
	...
	flagN = flag.Int("flagN", 123, "DescriptionN")
)

func main() {
	iniflags.Parse()  // used instead of flag.Parse()
}
```

dev.ini

```ini
    # comment1
    flag1 = "val1"  # comment2

    ...
    [section]
    flagN = 4  # comment3
```

```bash

go run main.go -config dev.ini -flagX=foobar

```

Now all unset flags will get their value from .ini file provided in -config path.
If value is not found in the .ini, flag will retain it's default value.

Flag value priority:
  - value set via command-line
  - value from ini file
  - default value

Iniflags is compatible with real .ini config files with [sections] and #comments.
Sections and comments are skipped during config file parsing.

Iniflags can import another ini files. For example,

base.ini
```ini
flag1 = value1
flag2 = value2
```

dev.ini
```ini
#import "base.ini"
# Now flag1="value1", flag2="value2"

flag2 = foobar
# Now flag1="value1", while flag2="foobar"
```


All flags defined in the app can be dumped into stdout with ini-compatible sytax
by passing -dumpflags flag to the app. The following command creates ini-file
with all the flags defined in the app:

```bash
/path/to/the/app -dumpflags > initial-config.ini
```


Iniflags also supports config reload on SIGHUP signal:

```bash
kill -s SIGHUP <app_pid>
```

