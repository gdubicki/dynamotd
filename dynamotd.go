package main

import "fmt"

func init() {
	configure()
}

func main() {
	rows := getRows()

	if isModeStatic() {
		printStatic(rows)
	} else {
		fmt.Printf("Other modes than 'static' not supported yet!")
	}
}
