package util

import (
	"embed"
	"io/fs"
	"log"
	"os"
)

//go:embed help_docs/*
var helpDocs embed.FS

func DisplayHelpDocs(topic string) {

	file := topic + HelpDocExtension
	data, readErr := fs.ReadFile(helpDocs, "help_docs/"+file)
	if readErr != nil {
		log.Fatalln(readErr)
	}

	if _, writeErr := os.Stdout.Write(data); writeErr != nil {
		log.Fatalln(writeErr)
	}
}
