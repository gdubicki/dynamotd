package main

import (
	"fmt"
	"github.com/fatih/color"
)

func printStatic(rows []Row) {
	for _, row := range rows {
		printStaticRow(row)
	}
}

func printStaticRow(row Row) {

	if row.isEmptyLine() {
		fmt.Print("\n")
	} else {
		labelToPrint := ""
		valueToPrint := ""

		for _, colorString := range row.label.text {
			labelToPrint += getStringToPrint(colorString)
		}

		for _, colorString := range row.value.text {
			valueToPrint += getStringToPrint(colorString)
		}

		fmt.Printf("%s : %s\n", labelToPrint, valueToPrint)
	}
}

func getStringToPrint(colorString ColorString) string {
	// generate function that prints in given color
	// (https://github.com/fatih/color#insert-into-noncolor-strings-sprintfunc)
	// TODO: pregenerate (store in struct?) the color printing functions
	colorPrint := color.New(colorString.color.staticColor).SprintFunc()

	return fmt.Sprintf("%s", colorPrint(colorString.text))
}
