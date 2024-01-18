package cmd

import (
	"flag"
	"fmt"
	"jit/pkg/util"
	"log"
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
		Initialize(args)
		break
	default:
		log.Fatalf("Invalid command %s: use jit -h for help\n", command)
	}
}

func Jit() {
	flag.Parse()

	if help {
		util.DisplayHelpDocs("index")
		return
	}

	if version {
		fmt.Printf("Jit Version %s", util.JitVersion)
		return
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
