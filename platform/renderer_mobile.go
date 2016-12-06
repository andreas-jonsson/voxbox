// +build mobile

/*
Copyright (C) 2016 Andreas T Jonsson

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU General Public License for more details.

You should have received a copy of the GNU General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

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
