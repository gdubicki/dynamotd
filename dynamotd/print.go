package dynamotd

import (
	"fmt"
	"strings"
	"unicode/utf8"

	"github.com/fatih/color"
)

func Print(rows []Row) {
	labelWidth := 0
	for _, row := range rows {
		labelWidth = max(labelWidth, utf8.RuneCountInString(row.Label.String()))
	}

	for _, row := range rows {
		printRow(row, labelWidth)
	}
}

func printRow(row Row, labelWidth int) {
	if row.isEmptyLine() {
		fmt.Print("\n")
		return
	}

	labelToPrint := ""
	valueToPrint := ""

	for _, colorString := range row.Label.text {
		labelToPrint += getStringToPrint(colorString)
	}

	for _, colorString := range row.Value.text {
		valueToPrint += getStringToPrint(colorString)
	}

	// pad based on the visible length, as the label may contain color escape codes
	padding := strings.Repeat(" ", labelWidth-utf8.RuneCountInString(row.Label.String()))
	fmt.Printf("%s%s : %s\n", labelToPrint, padding, valueToPrint)
}

func getStringToPrint(colorString ColorString) string {
	if colorString.Color == NoColor {
		return colorString.Text
	}
	return color.New(colorString.Color).Sprint(colorString.Text)
}
