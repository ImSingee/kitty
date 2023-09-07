package tl

import "github.com/ImSingee/go-ex/pp"

func symBlue(s string) string {
	return pp.BlueString(s).GetForStdout()
}

func symRed(s string) string {
	return pp.RedString(s).GetForStdout()
}

func symGreen(s string) string {
	return pp.GreenString(s).GetForStdout()
}

var gray = pp.GetColor(38, 5, 240)

func symGray(s string) string {
	return pp.ColorString(gray, s).GetForStdout()
}
