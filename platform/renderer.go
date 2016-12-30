// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

package platform

import (
	"log"
	"strings"

	"github.com/goxjs/gl"
)

type Renderer interface {
	Clear()
	Present()
	Shutdown()
	ToggleFullscreen()
	SetWindowTitle(title string)
}

func LogGLInfo() {
	log.Println("OpenGL Info")
	log.Println(gl.GetString(gl.VERSION))
	log.Println(gl.GetString(gl.VENDOR))
	log.Println(gl.GetString(gl.RENDERER))

	log.Println()
	for _, ext := range strings.Split(gl.GetString(gl.EXTENSIONS), " ") {
		log.Println(ext)
	}
}
