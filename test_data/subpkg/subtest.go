package subpkg

import (
	"time"
)

const (
	Enum1 = iota
	Enum2
)

func SubPkg(e int) {
	if e == Enum1 {
		time.Sleep(1 * time.Nanosecond)
	}

	if e == Enum2 {
		time.Sleep(1 * time.Nanosecond)
	}
}
