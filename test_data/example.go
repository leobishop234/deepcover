package test_data

import "github.com/leobishop234/deepcover/test_data/subpkg"

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
