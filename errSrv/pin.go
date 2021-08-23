package errSrv

import (
	"github.com/mozillazg/go-slugify"
)

func trans(str string) string {
	return slugify.Slugify(str)
}
