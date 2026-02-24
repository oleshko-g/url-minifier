package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"os"
)

type app struct {
	targetsNum   int
	method       string
	url          string
	headers      map[string][]string
	currentState int
	stateMachine map[int]int
}

var a = app{
	currentState: enteringMethod,
	stateMachine: map[int]int{
		// currentState 			 // nextState
		enteringMethod:        enteringPath,
		enteringPath:          enteringHeaders,
		enteringHeaders:       enteringBody,
		enteringBody:          enteringBodyPath,
		enteringBodyPath:      enteringTargetsNumber,
		enteringTargetsNumber: enteringMethod,
	},
}

const (
	// Initial state
	enteringMethod = iota
	enteringPath
	enteringHeaders
	enteringBody
	enteringBodyPath
	enteringTargetsNumber
)

func main() {
	s := bufio.NewScanner(os.Stdin)
	for {
		a.prompt()
		if !s.Scan() {
			if err := s.Err(); err != nil {
				slog.Error(err.Error())
			}
		}
		a.setCurrentState(s.Text())
	}
}

func (a *app) prompt() {
	switch a.currentState {
	case enteringMethod:
		fmt.Print("Enter HTTP Method >")
	case enteringPath:
		fmt.Print("Enter URL Path >")
	case enteringHeaders:
		fmt.Print("Enter HTTP Headers")
	case enteringBody:
		fmt.Print("Enter Request Body")
	case enteringBodyPath:
		fmt.Print("Enter Request Body Path")
	case enteringTargetsNumber:
		fmt.Print("Enter a number targets to generate and save")
	}
}

func (a *app) setCurrentState(input string) {
	a.currentState = a.stateMachine[a.currentState]
}
