package main

import (
	"github.com/fatih/color"
)

type Row struct {
	label ColorText
	value ColorText
}

type ColorText struct {
	text []ColorString
}

type ColorString struct {
	color Color
	text string
}

type Color struct {
	staticColor color.Attribute
	dynamicColor string
}

var labelColor = Color{color.FgWhite, "white"}
var valueDescriptionColor = Color{color.FgWhite, "white"}
var valueNeutralColor = Color{color.FgBlue, "lightblue"}
var valueOkColor = Color{color.FgGreen, "lightgreen"}
var valueWarningColor = Color{color.FgYellow, "lightgoldenrodyellow"}
var valueCriticalColor = Color{color.FgRed, "indianred"}

func valueDescription(text string) ColorString {
	return  ColorString{valueDescriptionColor, text}
}

func valueNeutral(text string) ColorString {
	return  ColorString{valueNeutralColor, text}
}

func valueOk(text string) ColorString {
	return ColorString{valueOkColor, text}
}

func valueWarning(text string) ColorString {
	return ColorString{valueWarningColor, text}
}

func valueCritical(text string) ColorString {
	return ColorString{valueCriticalColor, text}
}

// for labels most (every?) of the times we need to have it whole
// in a single color
func singleColorLabel(text string) ColorText {
	labelString := ColorString{labelColor, text}
	singleColorLabel := []ColorString{labelString}
	return ColorText{singleColorLabel}
}

// values we also want it in a single color
func singleColorValue(text string) ColorText {
	labelString := ColorString{valueNeutralColor, text}
	singleColorLabel := []ColorString{labelString}
	return ColorText{singleColorLabel}
}

// merge a bunch of color strings of a valueDescription to a single color text element
func toColorText(colorStrings ...ColorString) ColorText {
	return ColorText{colorStrings}
}

// generating and testing for empty lines
func emptyLine() Row {
	return Row{
		singleColorLabel(""),
		singleColorValue(""),
	}
}

func (row Row) isEmptyLine() bool {
	labelText := row.label.text
	valueText := row.value.text

	labelIsEmpty := len(labelText) == 1 && labelText[0].text == ""
	valueIsEmpty := len(valueText) == 1 && valueText[0].text == ""

	return labelIsEmpty && valueIsEmpty
}
