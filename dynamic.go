package main

import (
	"fmt"
	"github.com/rivo/tview"
)

func showDynamic(rows []Row) {
	textView := tview.NewTextView().
		SetDynamicColors(true).
		SetRegions(false)

	textView.SetBorder(false)

	for _, row := range rows {
		showDynamicRow(row, textView)
	}

	if err := tview.NewApplication().SetRoot(textView, true).Run(); err != nil {
		panic(err)
	}
}

func showDynamicRow(row Row, textView *tview.TextView) {

	if row.isEmptyLine() {
		fmt.Fprintf(textView, "\n")
	} else {
		labelToShow := ""
		valueToShow := ""

		for _, colorString := range row.label.text {
			labelToShow += getStringToShow(colorString)
		}

		for _, colorString := range row.value.text {
			valueToShow += getStringToShow(colorString)
		}

		fmt.Fprintf(textView, "%-20s : %s\n", labelToShow, valueToShow)
	}
}

func getStringToShow(colorString ColorString) string {
	// tview uses tags like "[white]" to set color for appended text
	return "[" + colorString.color.dynamicColor + "]" + colorString.text + "[-]"
}
