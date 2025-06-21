package subpkg

import (
	"strconv"
)

func SubPkg() {
	_, err := strconv.Atoi("1")
	if err != nil {
		panic(err)
	}
}
