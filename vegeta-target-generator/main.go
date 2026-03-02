package main

import (
	"bufio"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"net/url"
	"os"
	"path"
	"strconv"
	"strings"

	urlGen "github.com/oleshko-g/url-minifier/internal/url-generator"
)

type app struct {
	state
	stateMachine    map[state]state
	authrity        string
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
	defaultAuthority   = `http://`
	defaultHost        = `localhost`
	defaultPort        = `8080`
	defaultTargetsPath = `./vegeta-targets/`
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
	authrity: defaultAuthority,
	host:     defaultHost,
	port:     defaultPort,
	headers:  make(map[string][]string),
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
	input = strings.ToUpper(input)
	switch input {
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
	targetsDirPath := path.Join(defaultTargetsPath, fmt.Sprintf("%s:%s", a.host, a.port), fmt.Sprintf("%s-%s", a.method, a.uri))
	err := os.RemoveAll(defaultTargetsPath)
	if err != nil {
		return err
	}

	err = os.MkdirAll(targetsDirPath, defaultDirPerms)
	if err != nil {
		return err
	}
	a.targetsDirPath = targetsDirPath
	fmt.Printf("Created dir \"%s\"\n", targetsDirPath)
	a.setCurrentState()
	return nil
}

func handleCreatingTargets(a *app, _ string) error {
	var (
		ctx, cancel   = context.WithCancel(context.Background())
		errs          = make(chan error, a.targetsNumber)
		targetBodies  = make(chan string, a.targetsNumber)
		bodyFilePaths = make(chan string, a.targetsNumber)
		targets       = make(chan string, a.targetsNumber)
		urlGenerator  = urlGen.NewURLGenerator(2)
		targetsFile   *os.File
	)
	defer cancel()

	targetsFile, err := openFile(path.Join(a.targetsDirPath, "targets.txt"))
	if err != nil {
		return err
	}
	defer targetsFile.Close()

	bodiesDirPath := path.Join(a.targetsDirPath, "bodies")
	err = os.MkdirAll(bodiesDirPath, defaultDirPerms)
	if err != nil {
		return err
	}

	for targetIdx := range a.targetsNumber {
		go func() {
			select {
			case <-ctx.Done():
				return
			case targetBodies <- urlGenerator.Generate():
			}
		}()

		go func() {
			select {
			case <-ctx.Done():
				return
			case body := <-targetBodies:
				targetBodyFilePath := path.Join(bodiesDirPath, fmt.Sprintf("target-%d-body.txt", targetIdx))
				fp, err := openFile(targetBodyFilePath)
				if err != nil {
					errs <- err
				}

				if _, err = fp.Write([]byte(body)); err != nil {
					errs <- err
				}

				bodyFilePaths <- fp.Name()
			}
		}()

		go func() {
			select {
			case <-ctx.Done():
				return
			case bodyFilePath := <-bodyFilePaths:
				targetBuilder := &strings.Builder{}
				err = a.buildTarget(targetBuilder, bodyFilePath)
				if err != nil {
					errs <- err
				}
				targets <- targetBuilder.String()
			}
		}()

		target := <-targets
		_, err = targetsFile.Write([]byte(target))
		if err != nil {
			errs <- err
		}

	}

	cancel()

	select {
	case err := <-errs:
		cancel()
		return err
	case <-ctx.Done():
	}

	a.setCurrentState()
	return nil
}

func (a *app) setCurrentState() {
	nextState := a.stateMachine[a.state]
	a.state = nextState
}

func (a *app) buildTarget(builder *strings.Builder, bodyFilePath string) error {
	// write first line
	builder.WriteString(a.method)
	builder.WriteRune(' ')
	builder.WriteString(a.authrity)
	builder.WriteString(a.host)
	builder.WriteRune(':')
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

	if bodyFilePath != "" {
		// write the path to the body
		builder.WriteRune('@')
		builder.WriteString(bodyFilePath)
		builder.WriteRune('\n')
	}

	builder.WriteRune('\n')

	return nil
}

// openFile creates or opens and truncates the named file for writing.
func openFile(name string) (*os.File, error) {
	fp, err := os.OpenFile(name, os.O_WRONLY|os.O_CREATE|os.O_TRUNC|os.O_APPEND, 0o655)
	if err != nil {
		return nil, err
	}
	return fp, nil
}
