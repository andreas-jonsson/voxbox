//go:generate go run data/generate.go

package main

import (
	"fmt"

	"github.com/andreas-jonsson/voxbox/entry"
)

var banner = `
+------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
| Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
| Contact <mail@andreasjonsson.se>                                |
+-----------------------------------------------------------------+
`

func main() {
	//defer profile.Start(profile.ProfilePath(".")).Stop()

	fmt.Println(banner)
	entry.Entry()
}
