// +build ignore

// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

package main

import (
	"log"
	"net/http"
	"path"

	"github.com/shurcooL/vfsgen"
)

type fsWrapper struct {
	http.FileSystem
}

func (fs fsWrapper) Open(name string) (http.File, error) {
	return fs.FileSystem.Open(path.Join("data", "src", name))
}

func main() {
	fs := fsWrapper{http.Dir("")}
	err := vfsgen.Generate(&fs, vfsgen.Options{
		Filename:     path.Join("data", "data.go"),
		PackageName:  "data",
		BuildTags:    "!dev",
		VariableName: "FS",
	})
	if err != nil {
		log.Fatalln(err)
	}
}
