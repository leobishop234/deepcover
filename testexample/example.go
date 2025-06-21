package testexample

import "deepcover/testexample/subpkg"

func Top() {
	Bottom()
}

func Bottom() {
	subpkg.SubPkg()

	inter := newInterface()
	inter.Method()
}
