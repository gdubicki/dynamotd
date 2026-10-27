package dynamotd

import (
	"fmt"

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

// NoColor means "print in the terminal's default color". It is the zero value of color.Attribute,
// so it is also used for ColorStrings without an explicit color.
const NoColor = color.Reset

var LabelColor = NoColor

var ValueDescriptionColor = NoColor
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
	return ColorText{[]ColorString{{LabelColor, text}}}
}

// values we also want it in a single color
func SingleColorValue(text string) ColorText {
	return ColorText{[]ColorString{{ValueNeutralColor, text}}}
}

// merge a bunch of color strings of a valueDescription to a single color text element
func ToColorText(colorStrings ...ColorString) ColorText {
	return ColorText{colorStrings}
}

// ErrorRow is what a plugin returns when it can't get its data, so that a single failing row
// doesn't prevent showing the rest of the MOTD.
func ErrorRow(label string, err error) Row {
	return Row{
		Label: SingleColorLabel(label),
		Value: ToColorText(ValueCritical(fmt.Sprintf("can't get: %s", err))),
	}
}

// TimedOutRow is shown for a row type that didn't generate its rows in time.
func TimedOutRow(label string) Row {
	return Row{
		Label: SingleColorLabel(label),
		Value: ToColorText(ValueWarning("timed out")),
	}
}

func (row Row) isEmptyLine() bool {
	return len(row.Label.text) == 0 && len(row.Value.text) == 0
}

// String returns the text without any colors.
func (colorText ColorText) String() string {
	result := ""
	for _, colorString := range colorText.text {
		result += colorString.Text
	}
	return result
}
