package dynamotd

import (
	"github.com/fatih/color"
)

type Row struct {
	Label ColorText
	Value ColorText
}

type ColorText struct {
	text []ColorString
}

type ColorString struct {
	Color Color
	Text  string
}

type Color struct {
	staticColor  color.Attribute
	dynamicColor string
}

var LabelColor = Color{color.FgWhite, "white"}
var ValueDescriptionColor = Color{color.FgWhite, "white"}
var ValueNeutralColor = Color{color.FgBlue, "lightblue"}
var ValueOkColor = Color{color.FgGreen, "lightgreen"}
var ValueWarningColor = Color{color.FgYellow, "lightgoldenrodyellow"}
var ValueCriticalColor = Color{color.FgRed, "indianred"}

func ValueDescription(text string) ColorString {
	return ColorString{ValueDescriptionColor, text}
}

func ValueNeutral(text string) ColorString {
	return ColorString{ValueNeutralColor, text}
}

func ValueOk(text string) ColorString {
	return ColorString{ValueOkColor, text}
}

func ValueWarning(text string) ColorString {
	return ColorString{ValueWarningColor, text}
}

func ValueCritical(text string) ColorString {
	return ColorString{ValueCriticalColor, text}
}

// for labels most (every?) of the times we need to have it whole
// in a single color
func SingleColorLabel(text string) ColorText {
	labelString := ColorString{LabelColor, text}
	singleColorLabel := []ColorString{labelString}
	return ColorText{singleColorLabel}
}

// values we also want it in a single color
func SingleColorValue(text string) ColorText {
	labelString := ColorString{ValueNeutralColor, text}
	singleColorLabel := []ColorString{labelString}
	return ColorText{singleColorLabel}
}

// merge a bunch of color strings of a valueDescription to a single color text element
func ToColorText(colorStrings ...ColorString) ColorText {
	return ColorText{colorStrings}
}

func (row Row) isEmptyLine() bool {
	return len(row.Label.text) == 0 && len(row.Value.text) == 0
}
