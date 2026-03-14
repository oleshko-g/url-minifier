package main

import "os"

func fOSExit() {
	os.Exit(1) // want "calls os.Exit outside of main func"
}
