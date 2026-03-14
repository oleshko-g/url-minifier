package main

import "log"

func fLogFatal() {
	log.Fatal("fatal") // want "calls log.Fatal outside of main func"
}
