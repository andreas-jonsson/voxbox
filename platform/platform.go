// +------------------=V=o=x=B=o=x=-=E=n=g=i=n=e=--------------------+
// | Copyright (C) 2016-2017 Andreas T Jonsson. All rights reserved. |
// | Contact <mail@andreasjonsson.se>                                |
// +-----------------------------------------------------------------+

package platform

import (
	"log"
	"math"
	"path"
	"sync/atomic"
)

var (
	ConfigPath string
	idCounter  uint64
)

func CfgRootJoin(p ...string) string {
	return path.Clean(path.Join(ConfigPath, path.Join(p...)))
}

func NewId64() uint64 {
	if idCounter == math.MaxUint64 {
		log.Panicln("id space exhausted")
	}
	return atomic.AddUint64(&idCounter, 1) - 1
}
