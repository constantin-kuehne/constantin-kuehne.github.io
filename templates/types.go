package templates

import (
	"time"

	"github.com/a-h/templ"
)

type SiteInfo struct {
	Date     time.Time
	FileName string
	Dir      string
	Content  templ.Component
}

type Post struct {
	Title string
}
