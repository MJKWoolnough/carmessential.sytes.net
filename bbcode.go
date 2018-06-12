package main

import (
	bbc "vimagination.zapto.org/bbcode"
	"vimagination.zapto.org/bbcode/bbcodehtml"
)

var bbcode = bbc.New(bbc.HTMLText,
	bbcodehtml.Align,
	bbcodehtml.LeftAlign,
	bbcodehtml.CentreAlign,
	bbcodehtml.CenterAlign,
	bbcodehtml.RightAlign,
	bbcodehtml.FullAlign,
	bbcodehtml.Color,
	bbcodehtml.Colour,
	bbcodehtml.Font,
	bbcodehtml.Bold,
	bbcodehtml.Italic,
	bbcodehtml.Strikethough,
	bbcodehtml.Underline,
	bbcodehtml.Size,
	bbcodehtml.Heading1,
	bbcodehtml.Heading2,
	bbcodehtml.Heading3,
	bbcodehtml.Heading4,
	bbcodehtml.Heading5,
	bbcodehtml.Heading6,
	bbcodehtml.Heading7,
	bbcodehtml.Code,
	bbcodehtml.Image,
	bbcodehtml.List,
	bbcodehtml.Table,
	bbcodehtml.URL,
)
