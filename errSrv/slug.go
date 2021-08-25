package errSrv

import (
	"github.com/mozillazg/go-slugify"
)

func slug(str string) string {
	return slugify.Slugify(str)
}
