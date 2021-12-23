package spider

import (
	"spider-zxxk/browser"
)

func New() *Spider {
	return &Spider{
		browser: browser.New(),
	}
}

type Spider struct {
	browser 		*browser.Browser
}
