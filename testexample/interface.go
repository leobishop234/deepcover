package testexample

import "strconv"

type Interface interface {
	Method()
}

func newInterface() Interface {
	return &Struct{}
}

type Struct struct{}

func (s *Struct) Method() {
	_, err := strconv.Atoi("1")
	if err != nil {
		panic(err)
	}
}
