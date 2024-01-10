package init_command

import (
	"flag"
	"fmt"
	"log"
)

func InitCommand(args []string) {
	initCmd := flag.NewFlagSet("init_command", flag.ExitOnError)
	bare := initCmd.Bool("bare", false, "Create a bare repository")
	gitDir := initCmd.String("git-directory", "", "Set the git directory path")
	mode := initCmd.Int("mode", 0666, "Set the mode")

	// Parse the init_command command arguments
	if err := initCmd.Parse(args); err != nil {
		log.Fatalln("Error parsing init_command command:", err)
	}

	// Access the additional arguments after the flags
	additionalArgs := initCmd.Args()

	// Use the flags and additional arguments
	fmt.Println("Init Command:")
	fmt.Printf("Bare: %v, Git Directory: %s, Mode: %o\n", *bare, *gitDir, *mode)
	fmt.Println("Additional Args:", additionalArgs)
}
