package testexample

import "time"

func Top() {
	Middle()
}

func Middle() {
	Bottom()
}

func Bottom() {
	time.Sleep(time.Nanosecond)
}
