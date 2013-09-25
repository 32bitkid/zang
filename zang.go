package main

import (
	"bufio"
	"bytes"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
)

var (
	repoFlag       string
	headFlag       string
	checkStaleFlag bool
)

func init() {
	flag.StringVar(&repoFlag, "repo", "", "the path to the repository")
	flag.StringVar(&headFlag, "head", "master", "the commit to check for stale documentation")
	flag.BoolVar(&checkStaleFlag, "check", true, "search the repository for changes since the documentation was written")
}

func main() {
	flag.Parse()

	start := time.Now()
	if err := modeSwitch(flag.Arg(0), flag.Arg(1)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
		os.Exit(1)
	}

	fmt.Printf("%v\n", time.Since(start))
}

func modeSwitch(inName, outName string) error {

	if len(inName) == 0 {
		return pipeMode()
	}

	if len(outName) == 0 {
		return errors.New("A output file must be provided")
	}

	inputFileInfo, statErr := os.Stat(inName)
	if statErr != nil {
		return statErr
	}

	if inputFileInfo.IsDir() {
		return dirMode(inName, outName)
	} else {
		return fileMode(inName, outName)
	}
}

func deferredCreate(filename string) func() (io.WriteCloser, error) {
	return func() (io.WriteCloser, error) {
		return os.Create(filepath.Clean(filename))
	}
}

func getStdout() (io.WriteCloser, error) {
	return os.Stdout, nil
}

func pipeMode() error {
	resultChannel := make(chan Result, 1)

	go processFile(os.Stdin, getStdout, resultChannel)

	return (<-resultChannel).Execute()
}

func dirMode(inFolder, outFolder string) error {

	inFolder = filepath.Clean(inFolder)
	outFolder = filepath.Clean(outFolder)

	expectedResults := 0

	resultChannel := make(chan Result)

	dirWalker := func(path string, info os.FileInfo, err error) error {
		if filepath.Ext(path) == ".md" {
			expectedResults += 1
			go handleMarkdown(path, inFolder, outFolder, resultChannel)
		}
		return err
	}

	walkErr := filepath.Walk(inFolder, dirWalker)

	for expectedResults > 0 {
		if err := (<-resultChannel).Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
		expectedResults--
	}

	return walkErr
}

func handleMarkdown(path, inFolder, outFolder string, resultChannel chan<- Result) {

	var inFile *os.File
	var relativePath string
	var err error

	if inFile, err = os.Open(path); err != nil {
		resultChannel <- ErrorResult{err}
		return
	}

	defer func() {
		inFile.Close()
	}()

	if relativePath, err = filepath.Rel(inFolder, path); err != nil {
		resultChannel <- ErrorResult{err}
		return
	}

	destFile := filepath.Join(outFolder, relativePath)

	if err = os.MkdirAll(filepath.Dir(destFile), os.ModePerm); err != nil {
		resultChannel <- ErrorResult{err}
		return
	}

	processFile(inFile, deferredCreate(destFile), resultChannel)
}

func fileMode(inFileName, outFileName string) error {

	inFile, openErr := os.Open(inFileName)

	defer func() {
		inFile.Close()
	}()

	if openErr != nil {
		return openErr
	}

	resultChannel := make(chan Result, 1)

	go processFile(inFile, deferredCreate(outFileName), resultChannel)

	return (<-resultChannel).Execute()
}

func processFile(in io.Reader, getOutFile func() (io.WriteCloser, error), resultChannel chan<- Result) {
	var buffer bytes.Buffer
	var err error

	input := bufio.NewScanner(in)

	git := execGit.memoize()

	skipScan := false

	for skipScan || input.Scan() {
		skipScan = false

		text := input.Text()

		if args, success := parseAsGitCommand(text); success {
			if err = args.process(&buffer, git); err != nil {
				resultChannel <- ErrorResult{err}
				return
			}

			if checkStaleFlag {
				args.checkGitChanges(&buffer, git)
			}

			if skipScan, err = skipExistingCode(input); err != nil {
				resultChannel <- ErrorResult{err}
			}

		} else {
			fmt.Fprintln(&buffer, text)
		}
	}

	if err := input.Err(); err != nil {
		resultChannel <- ErrorResult{err}
	} else {
		resultChannel <- WriteFileResult{getFile: getOutFile, content: &buffer}
	}
}

func skipExistingCode(input TextScanner) (bool, error) {
	// Scan the next line
	if input.Scan() == false {
		return false, nil
	}

	// Check for a generated start marker
	if input.Text() != beginMarker {
		return true, nil
	}

	// Keep scanning until the end marker
	for input.Scan() {
		if input.Text() == endMarker {
			return false, nil
		}
	}

	// Something went wrong, and we reached the end of the file.
	return false, errors.New("End marker was not found before end of file...")
}
