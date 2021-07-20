package buildinfo

import "github.com/bluemir/0xC0DE/internal/static"

var (
	Version   string
	AppName   string
	BuildTime string = static.Static.MustString(".time")
)
