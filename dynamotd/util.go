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
	Color color.Attribute
	Text  string
}

var LabelColor = color.FgWhite

var ValueDescriptionColor = color.FgWhite
var ValueNeutralColor = color.FgBlue
var ValueOkColor = color.FgGreen
var ValueWarningColor = color.FgYellow
var ValueCriticalColor = color.FgRed

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
