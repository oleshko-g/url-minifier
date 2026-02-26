package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"strings"
)

type app struct {
	state
	stateMachine map[state]state
	method       string
	url          string
	bodyPath     string
	targetsNum   int
	headers      map[string][]string
	handlers     map[state]handler
}

type handler func(a *app, input string) error

var urlGen = app{
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
		urlGen.prompt()
		if !s.Scan() {
			if err := s.Err(); err != nil {
				slog.Error(err.Error())
			}
		}
		urlGen.handle(s.Text())
		urlGen.setCurrentState()
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
	if a.method == "" {
		return
	}

	if a.url == "" {
		return
	}

	if a.headers == nil {
		return
	}

	if a.bodyPath == "" {
		return
	}

	a.state = a.stateMachine[a.state]
}

func (a *app) handle(input string) error {
	switch a.state {
	case enteringMethod:
		if err := handleMethod(a, input); err != nil {
			return err
		}
	}
	return nil
}

func handleMethod(a *app, input string) error {
	switch input := strings.ToUpper(input); input {
	case http.MethodPost:
	default:
		return fmt.Errorf("unsupported method")
	}
	a.method = input
	return nil
}

func handlePath(a *app, input string) error {
	a.url = input
	return nil
}

func handleHeaders(a *app, input string) error {
	h := strings.Split(input, ":")
	header := h[0]
	value := h[1]
	a.headers[header] = append(a.headers[header], value)
	return nil
}

func handleBodyPath(a *app, input string) error {
	a.bodyPath = input
	return nil
}

func handleTargetsNumber(a *app, input string) error {
	fileName := fmt.Sprintf("vegeta targets %s %s", a.method, a.url)
	f, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o655)
	if err != nil {
		return err
	}
	_ = f
	return nil
}
