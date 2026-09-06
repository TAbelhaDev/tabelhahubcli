package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) > 1 {
		switch os.Args[1] {
		case "ipc":
			os.Exit(runIPC(os.Args[2:]))
		}
	}
	printUsage()
	os.Exit(1)
}

func printUsage() {
	fmt.Fprintln(os.Stderr, "uso: tahubcli ipc <método> [key=value...] --json")
	fmt.Fprintln(os.Stderr, "métodos: suggest (lê JSON do stdin)")
}
