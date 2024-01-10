package main

import (
	"flag"
	"fmt"
	"jit/cmd/init_command"
	"jit/pkg/util"
	"log"
	"os"
)

var help bool
var version bool

func init() {
	flag.BoolVar(&help, "help", false, "jit -h | jit --help")
	flag.BoolVar(&help, "h", false, "jit -h | jit --help")

	flag.BoolVar(&version, "version", false, "jit -v | jit --version")
	flag.BoolVar(&version, "v", false, "jit -v | jit --version")
}

func handleCommand(command string, args []string) {

	switch command {
	case util.Init:
		init_command.InitCommand(args)
		break
	default:
		log.Fatalf("Invalid command %s: use jit -h for help\n", command)
	}
}

func main() {
	flag.Parse()

	if help {
		util.DisplayHelpDocs("index")
		os.Exit(0)
	}

	if version {
		fmt.Println("Jit Version: 1.0.0")
		os.Exit(0)
	}

	// Additional command handling
	if len(flag.Args()) > 0 {
		command := flag.Arg(0)
		commandArgs := flag.Args()[1:]
		handleCommand(command, commandArgs)
	} else {
		log.Fatalln("No command provided: use jit -h for help")
	}
}
