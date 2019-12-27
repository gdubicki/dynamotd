package main

func init() {
	configure()
}

func main() {
	rows := getRows()

	if isModeStatic() {
		printStatic(rows)
	} else {
		showDynamic(rows)
	}
}
