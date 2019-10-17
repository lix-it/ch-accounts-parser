package main

import (
	"fmt"
	"os"
	"path/filepath"
)

const (
	HTML = iota
	XML  = iota
)

func detectFileType(file *os.File) (int, error) {
	ext := filepath.Ext(file.Name())
	switch ext {
	case ".html":
		return HTML, nil
	case ".xml":
		return XML, nil
	}
	return 0, fmt.Errorf("could not get file ext: %v", ext)
}
