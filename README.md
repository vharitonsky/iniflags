Combine standard go flags with your ini file.

Usage:

```go
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
	iniflags.Parse()  # used instead of flag.Parse()
}
```

Now all go flags used in this app obtain default values from ini file
read from -config path.
Flag value priority:
  - value set via command-line
  - value from ini file
  - default value
