package dynamotd

import (
	"fmt"
	"github.com/fatih/color"
)

func Print(rows []Row) {
	for _, row := range rows {
		printRow(row)
	}
}

func printRow(row Row) {
	if row.isEmptyLine() {
		fmt.Print("\n")
	} else {
		labelToPrint := ""
		valueToPrint := ""

		for _, colorString := range row.Label.text {
			labelToPrint += getStringToPrint(colorString)
		}

		for _, colorString := range row.Value.text {
			valueToPrint += getStringToPrint(colorString)
		}

		// TODO: make the width of label column dynamic
		fmt.Printf("%-23s : %s\n", labelToPrint, valueToPrint)
	}
}

func getStringToPrint(colorString ColorString) string {
	// generate function that prints in given color
	// (https://github.com/fatih/color#insert-into-noncolor-strings-sprintfunc)
	// TODO: pregenerate (store in struct?) the color printing functions
	colorPrint := color.New(colorString.Color).SprintFunc()

	return fmt.Sprintf("%s", colorPrint(colorString.Text))
}
