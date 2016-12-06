// +----------------=V=o=x=B=o=x=-=E=n=g=i=n=e=-----------------+
// | Copyright (C) 2016 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                           |
// +------------------------------------------------------------+

package platform

type Renderer interface {
	Clear()
	Present()
	Shutdown()
	ToggleFullscreen()
	SetWindowTitle(title string)
}
