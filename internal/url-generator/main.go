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
	state
	stateMachine map[state]state
	handlers     map[state]handler
}

var a = app{
	state: enteringMethod,
	stateMachine: map[state]state{
		// currentState 			 // nextState
		enteringMethod:        enteringPath,
		enteringPath:          enteringHeaders,
		enteringHeaders:       enteringBody,
		enteringBody:          enteringBodyPath,
		enteringBodyPath:      enteringTargetsNumber,
		enteringTargetsNumber: enteringMethod,
	},
	handlers: map[state]handler{
		// currentState 			 // handler
		enteringMethod:        handleMethod,
		enteringPath:          handlePath,
		enteringHeaders:       handleHeaders,
		enteringBody:          handleBody,
		enteringBodyPath:      handleBodyPath,
		enteringTargetsNumber: handleTargetsNumber,
	},
}


type state int

const (
	// Initial state
	enteringMethod        state = 0
	enteringPath          state = 1
	enteringHeaders       state = 2
	enteringBody          state = 3
	enteringBodyPath      state = 4
	enteringTargetsNumber state = 5
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
	switch a.state {
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
	a.state = a.stateMachine[a.state]
}

type handler func(input string) error
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
