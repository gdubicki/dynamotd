package main

import (
	"github.com/fatih/color"
	"github.com/gdamore/tcell"
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
	dynamicColor tcell.Color
}

var labelColor = Color{color.FgWhite, tcell.ColorWhite}
var valueColor = Color{color.FgBlue, tcell.ColorWhiteSmoke}
var keyValueOkColor = Color{color.FgGreen, tcell.ColorGreen}
var keyValueWarningColor = Color{color.FgYellow, tcell.ColorYellow}
var keyValueCriticalColor = Color{color.FgRed, tcell.ColorRed}

func value(text string) ColorString {
	return  ColorString{valueColor, text}
}

func keyValueOk(text string) ColorString {
	return ColorString{keyValueOkColor, text}
}

func keyValueWarning(text string) ColorString {
	return ColorString{keyValueWarningColor, text}
}

func keyValueCritical(text string) ColorString {
	return ColorString{keyValueCriticalColor, text}
}

// for labels most (every?) of the times we need to have it whole
// in a single color
func singleColorLabelText(text string) ColorText {
	labelString := ColorString{labelColor, text}
	singleColorLabel := []ColorString{labelString}
	return ColorText{singleColorLabel}
}

// values we also want it in a single color
func singleColorValueText(text string) ColorText {
	labelString := ColorString{valueColor, text}
	singleColorLabel := []ColorString{labelString}
	return ColorText{singleColorLabel}
}

// merge a bunch of color strings of a value to a single color text element
func toColorText(colorStrings ...ColorString) ColorText {
	return ColorText{colorStrings}
}

// generating and testing for empty lines
func emptyLine() Row {
	return Row{
		singleColorLabelText(""),
		singleColorValueText(""),
	}
}

func (row Row) isEmptyLine() bool {
	labelText := row.label.text
	valueText := row.value.text

	labelIsEmpty := len(labelText) == 1 && labelText[0].text == ""
	valueIsEmpty := len(valueText) == 1 && valueText[0].text == ""

	return labelIsEmpty && valueIsEmpty
}
