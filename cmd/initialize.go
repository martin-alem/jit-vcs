// File: Initialize.go
// Package: cmd

// Program Description:
// This file handles the parsing of the command flags and arguments
// It also makes sure string flags are validated before calling the InitializeJitRepository method

// Author: Martin Alemajoh
// Jit-VCS - v1.0.0
// Created on: January 16, 2024

package cmd

import (
	"flag"
	"jit/internal"
	"log"
)

var initCmd *flag.FlagSet
var quiet bool
var bare bool
var template string
var separateJitDir string
var objectFormat string
var branch string
var permission string

func init() {
	initCmd = flag.NewFlagSet("initialize", flag.ExitOnError)
	initCmd.BoolVar(&quiet, "quiet", false, "Only print error and warning messages; all other output will be suppressed.")
	initCmd.BoolVar(&quiet, "q", false, "Only print error and warning messages; all other output will be suppressed.")
	initCmd.BoolVar(&bare, "bare", false, "Create a bare repository. If JIT_DIR environment is not set, it is set to the current working directory")
	initCmd.StringVar(&template, "template", "", "Specify the directory from which templates will be used")
	initCmd.StringVar(&separateJitDir, "separate-jit-dir", "", "Instead of initializing the repository as a directory to either $JIT_DIR or ./.jit/, create a text file there containing the path to the actual repository")
	initCmd.StringVar(&objectFormat, "object-format", "sha1", "Specify the given object format (hash algorithm) for the repository. The valid values are sha1 and sha256. sha1 is the default.")
	initCmd.StringVar(&branch, "b", "main", "Use the specified name for the initial branch in the newly created repository. Default branch is main")
	initCmd.StringVar(&branch, "initial-branch", "main", "Use the specified name for the initial branch in the newly created repository. Default branch is main")
	initCmd.StringVar(&permission, "perm", "0755", "Specifies the directory's permission. Default is 0755")
}

func Initialize(args []string) {
	// Parse the initialize command arguments
	if err := initCmd.Parse(args); err != nil {
		log.Fatalln("Error parsing initialize command:", err)
	}

	// Access the first argument
	workingDirectory := initCmd.Arg(0)
	options := map[string]any{
		"quiet":            quiet,
		"bare":             bare,
		"separate-jit-dir": separateJitDir,
		"template":         template,
		"object-format":    objectFormat,
		"initial-branch":   branch,
		"perm":             permission,
	}
	_, initErr := internal.InitializeJitRepository(options, workingDirectory)
	if initErr != nil {
		log.Fatalln(initErr)
	}
}
