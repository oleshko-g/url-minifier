package main

import (
	"bufio"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	urlGen "github.com/oleshko-g/url-minifier/internal/url-generator/internal/url-generator"
)

type app struct {
	state
	stateMachine    map[state]state
	host            string
	port            string
	method          string
	uri             string
	targetsNumber   int
	targetsDirPath  string
	targetsFilePath *os.File
	headers         map[string][]string
	handlers        map[state]handler
}

const (
	defaultHost        = `localhost`
	defaultPort        = `8080`
	defaultTargetsPath = `./testdata/vegeta-targets/`
	// wrx _rx _rx
	defaultDirPerms os.FileMode = 0o0755
)

type state int

const (
	// Initial state
	enteringMethod state = iota
	enteringPath
	enteringHeaders
	enteringTargetsNumber
	creatingTargetsDir
	creatingTargets
)

type handler func(a *app, input string) error

var targetGenerator = app{
	// initital statte
	state: enteringMethod,
	stateMachine: map[state]state{
		// currentState 			 // nextState
		enteringMethod:        enteringPath,
		enteringPath:          enteringHeaders,
		enteringHeaders:       enteringTargetsNumber,
		enteringTargetsNumber: creatingTargetsDir,
		creatingTargetsDir:    creatingTargets,
		creatingTargets:       enteringPath,
	},
	handlers: map[state]handler{
		enteringMethod:        handleMethod,
		enteringPath:          handlePath,
		enteringHeaders:       handleHeader,
		enteringTargetsNumber: handleTargetsNumber,
		creatingTargetsDir:    handleCreatingTargetsDir,
		creatingTargets:       handleCreatingTargets,
	},
	host:    defaultHost,
	port:    defaultPort,
	headers: make(map[string][]string),
}

func main() {
	s := bufio.NewScanner(os.Stdin)
	for {
		targetGenerator.prompt()

		for !s.Scan() {
			if err := s.Err(); err != nil {
				slog.Error(err.Error())
			}
		}

		err := targetGenerator.handle(s.Text())
		if err != nil {
			slog.Error(err.Error())
		}
	}
}

func (a *app) prompt() {
	switch a.state {
	case enteringMethod:
		fmt.Print("Enter HTTP Method >")
	case enteringPath:
		fmt.Print("Enter URL Path >")
	case enteringHeaders:
		fmt.Print("Enter HTTP Headers or leave empty >")
	case enteringTargetsNumber:
		fmt.Print("Enter a number targets to generate and save >")
	case creatingTargetsDir:
		fmt.Println("Creating targets dir...")
	case creatingTargets:
		fmt.Println("Creating targets...")
	}
}

func (a *app) handle(input string) error {
	handler := a.handlers[a.state]
	if err := handler(a, input); err != nil {
		return err
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
	a.setCurrentState()
	return nil
}

func handlePath(a *app, input string) error {
	if _, err := url.Parse(input); err != nil {
		return err
	}

	a.uri = input
	a.setCurrentState()
	return nil
}

func handleHeader(a *app, input string) error {
	// empty line is th end of headers
	if input == "" {
		a.setCurrentState()
		return nil
	}

	h := strings.Split(input, ":")
	if len(h) < 2 {
		return fmt.Errorf("malformed header")
	}
	header := http.CanonicalHeaderKey(h[0])
	value := strings.TrimSpace(h[1])
	a.headers[header] = append(a.headers[header], value)

	return nil
}

func handleTargetsNumber(a *app, input string) error {
	targetsNum, err := strconv.Atoi(input)
	if err != nil {
		return err
	}

	if targetsNum <= 0 {
		return fmt.Errorf("targets number must be an integer greater then 0")
	}

	a.targetsNumber = targetsNum
	a.setCurrentState()
	return nil

}

func handleCreatingTargetsDir(a *app, _ string) error {
	targetsDirPath := path.Join(defaultTargetsPath, fmt.Sprintf("%s-%s", a.method, a.uri))
	err := os.MkdirAll(targetsDirPath, defaultDirPerms)
	if err != nil {
		return err
	}
	a.targetsDirPath = targetsDirPath
	fmt.Printf("Created dir \"%s\"\n", targetsDirPath)
	a.setCurrentState()
	return nil
}

func handleCreatingTargets(a *app, _ string) error {
	targetBuilder := &strings.Builder{}
	if err := a.buildTargets(targetBuilder); err != nil {
		return err
	}
	targets := targetBuilder.String()
	_ = targets

	a.setCurrentState()
	return nil
}

func (a *app) setCurrentState() {
	nextState := a.stateMachine[a.state]
	a.state = nextState
}

func (a *app) generateTargets(number int) error {
	return nil
}

func (a *app) setFile(name string) error {
	fp, err := openFile(name)
	if err != nil {
		return err
	}

	a.targetsFilePath = fp
	return nil
}

func (a *app) buildTargets(builder *strings.Builder) error {
	for targetIdx := range a.targetsNumber {
		// write first line
		builder.WriteString(a.method)
		builder.WriteRune(' ')
		builder.WriteString(a.host)
		builder.WriteString(a.port)
		builder.WriteString(a.uri)
		builder.WriteRune('\n')

		// write headers
		for header, values := range a.headers {
			if len(values) == 0 {
				continue
			}

			builder.WriteString(header)
			builder.WriteString(": ")
			for i, value := range values {
				if value == "" {
					continue
				}
				builder.WriteString(value)

				// if not the last value
				if i+1 < len(values) {
					builder.WriteString(", ")
				}
			}
			builder.WriteRune('\n')
		}

		targetBodyFilePath := path.Join(a.targetsFilePath.Name(), fmt.Sprintf("target-%d-body.txt", targetIdx))
		fp, err := openFile(targetBodyFilePath)
		if err != nil {
			return err
		}

		bodyGen := urlGen.NewURLGenerator(fp, 2)
		_, err = bodyGen.Write(bodyGen.Generate())
		if err != nil {
			return err
		}

		// write the path to the body
		builder.WriteRune('@')
		builder.WriteString(targetBodyFilePath)
		builder.WriteRune('\n')
		builder.WriteRune('\n')
	}

	return nil
}

func openFile(name string) (*os.File, error) {
	fp, err := os.OpenFile(name, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0o655)
	if err != nil {
		return nil, err
	}
	return fp, nil
}
