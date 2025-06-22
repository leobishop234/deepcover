package testexample

import "deepcover/testexample/subpkg"

func Top() {
	Bottom()
}

func Bottom() {
	subpkg.SubPkg(subpkg.Enum1)

	inter := newInterface()
	inter.Method()
}

func Alternative() {
	subpkg.SubPkg(subpkg.Enum2)
}
