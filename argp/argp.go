package argp

import (
	"os"
	"strings"
)

func Get(name string, value string) string {
	for _, arg := range os.Args {
		option := strings.SplitN(arg, "=", 2)
		if len(option) == 2 {
			if "--"+name == option[0] {
				return option[1]
			}
		}
	}
	return value
}
