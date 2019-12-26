package main

func main() {

	// TODO: make rows configurable
	rows := []Row {
		timestamp(),
		emptyLine(),
		fqdn(),
		emptyLine(),
		load(),
	}

	// TODO: add optional (default?) dynamic mode
	printStatic(rows)
}
