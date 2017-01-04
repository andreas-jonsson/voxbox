// +build mobile

// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package platform

import "github.com/goxjs/gl"

type mobileRenderer struct {
}

func NewRenderer(configs ...Config) (*mobileRenderer, error) {
	r := mobileRenderer{}

	for _, cfg := range configs {
		if err = cfg(&rnd); err != nil {
			return nil, err
		}
	}

	return &r, nil
}

func (p *mobileRenderer) ToggleFullscreen() {
}

func (p *mobileRenderer) Clear() {
	gl.Clear(gl.COLOR_BUFFER_BIT | gl.DEPTH_BUFFER_BIT)
}

func (p *mobileRenderer) Present() {
}

func (p *mobileRenderer) Shutdown() {
}

func (p *mobileRenderer) SetWindowTitle(title string) {
}
