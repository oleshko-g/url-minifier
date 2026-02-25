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
	currentState appState
	stateMachine map[appState]appState
	handlers     map[appState]handlerFunc
}

var a = app{
	currentState: enteringMethod,
	stateMachine: map[appState]appState{
		// currentState 			 // nextState
		enteringMethod:        enteringPath,
		enteringPath:          enteringHeaders,
		enteringHeaders:       enteringBody,
		enteringBody:          enteringBodyPath,
		enteringBodyPath:      enteringTargetsNumber,
		enteringTargetsNumber: enteringMethod,
	},
	handlers: map[appState]handlerFunc{
		// currentState 			 // handler
		enteringMethod:        handleMethod,
		enteringPath:          handlePath,
		enteringHeaders:       handleHeaders,
		enteringBody:          handleBody,
		enteringBodyPath:      handleBodyPath,
		enteringTargetsNumber: handleTargetsNumber,
	},
}

type handlerFunc func(input string) error
type appState int

const (
	// Initial state
	enteringMethod        appState = 0
	enteringPath          appState = 1
	enteringHeaders       appState = 2
	enteringBody          appState = 3
	enteringBodyPath      appState = 4
	enteringTargetsNumber appState = 5
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
		a.handle(s.Text())
		a.setCurrentState()
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

func (a *app) setCurrentState() {
	a.currentState = a.stateMachine[a.currentState]
}

func (a *app) handle(input string) error {
	switch {

	}
	return nil
}

func handleMethod(input string) error {
	return nil
}

func handlePath(input string) error {
	return nil
}

func handleHeaders(input string) error {
	return nil
}

func handleBody(input string) error {
	return nil
}

func handleBodyPath(input string) error {
	return nil
}

func handleTargetsNumber(input string) error {
	return nil
}
