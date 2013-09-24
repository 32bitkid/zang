package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unicode"
)

var (
	repoFlag       string
	headFlag       string
	checkStaleFlag bool

	inFlag  string
	outFlag string
)

func init() {
	flag.StringVar(&repoFlag, "repo", "", "the path to the repository")
	flag.StringVar(&headFlag, "head", "master", "the commit to check for stale documentation")
	flag.BoolVar(&checkStaleFlag, "check", true, "search the repository for changes since the documentation was written")

	flag.StringVar(&inFlag, "in", "", "input folder to process")
	flag.StringVar(&outFlag, "out", "", "output folder")
}

func main() {
	flag.Parse()

	start := time.Now()
	if err := modeSwitch(flag.Arg(0), flag.Arg(1)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}

	fmt.Printf("%v\n", time.Since(start))
}

func modeSwitch(inName, outName string) error {

	if len(inName) == 0 {
		return fileMode(inName, outName)
	}

	fileInfo, statErr := os.Stat(inName)
	if statErr != nil {
		return statErr
	}

	if fileInfo.IsDir() {
		return dirMode(inName, outName)
	} else {
		return fileMode(inName, outName)
	}
}

func dirMode(inFolder, outFolder string) error {

	inFolder = filepath.Clean(inFolder)
	outFolder = filepath.Clean(outFolder)

	expectedResults := 0

	resultChannel := make(chan Result)

	dirWalker := func(path string, info os.FileInfo, err error) error {

		if filepath.Ext(path) == ".md" {

			relativePath, relErr := filepath.Rel(inFolder, path)

			if relErr != nil {
				return relErr
			}

			destFile := filepath.Join(outFolder, relativePath)

			os.MkdirAll(filepath.Dir(destFile), os.ModePerm)

			inFile := safeFile(os.Open, path, os.Stdin)

			expectedResults += 1
			go processFile(bufio.NewScanner(inFile), destFile, resultChannel)
		}
		return err
	}

	walkErr := filepath.Walk(inFolder, dirWalker)

	for resultIndex := 0; resultIndex < expectedResults; resultIndex++ {
		if err := (<-resultChannel).Execute(); err != nil {
			fmt.Fprintln(os.Stderr, err)
		}
	}

	return walkErr
}

func fileMode(inFileName, outFileName string) error {
	var inFile *os.File = safeFile(os.Open, inFileName, os.Stdin)

	resultChannel := make(chan Result, 1)

	go processFile(bufio.NewScanner(inFile), outFileName, resultChannel)

	return (<-resultChannel).Execute()
}

func processFile(input *bufio.Scanner, outfile string, resultChannel chan<- Result) {
	var reportedError error
	var buffer bytes.Buffer

	defer func() {
		if reportedError == nil {
			resultChannel <- WriteFileResult{fileName: outfile, content: &buffer}
		} else {
			resultChannel <- ErrorResult{reportedError}
		}
	}()

	git := memoizeExecGitFn(execGit)

	skipScan := false

	for skipScan || input.Scan() {
		skipScan = false

		text := input.Text()

		if args, success := parseAsGitCommand(text); success {
			if err := processGit(&buffer, git, args); err != nil {
				reportedError = err
				return
			}

			if checkStaleFlag {
				checkGitChanges(&buffer, git, args)
			}

			skipScan = skipExistingCode(input)
		} else {
			fmt.Fprintln(&buffer, text)
		}
	}

	reportedError = input.Err()
}

func skipExistingCode(input *bufio.Scanner) bool {
	// Scan the next line
	if input.Scan() == false {
		return false
	}

	// Check for a generated start marker
	if input.Text() == beginMarker {
		// Keep scanning until the end marker
		for input.Scan() {
			if input.Text() == endMarker {
				return false
			}
		}
	}

	return true
}

func filterLines(scanner *bufio.Scanner, filterFn func(line int) bool) []string {
	lines := make([]string, 0, 30)

	for line := 1; scanner.Scan(); line++ {
		if filterFn(line) {
			text := scanner.Text()

			if line == 1 && text[0] == '\xEF' && text[1] == '\xBB' && text[2] == '\xBF' {
				text = text[3:]
			}

			lines = append(lines, text)
		}
	}

	return lines
}

func writeTrimmedLines(output io.Writer, lines ...string) {
	trimAmount := calculateAmountToTrim(lines)

	for _, str := range lines {
		if len(str) <= trimAmount {
			fmt.Fprintln(output, str)
		} else {
			fmt.Fprintln(output, str[trimAmount:])
		}
	}
}

func calculateAmountToTrim(lines []string) int {
	amountToTrim := int(^uint(0) >> 1)

	for _, lineContent := range lines {
		for characterPosition, rune := range lineContent {
			if characterPosition >= amountToTrim {
				break
			}
			if !unicode.IsSpace(rune) {
				if amountToTrim > characterPosition {
					amountToTrim = characterPosition
				}
				break
			}
		}
		if amountToTrim == 0 {
			break
		}
	}

	return amountToTrim
}

func safeFile(method func(string) (*os.File, error), fileName string, defaultFile *os.File) *os.File {

	if len(fileName) == 0 {
		return defaultFile
	}

	cleanFileName := filepath.Clean(fileName)

	file, fileErr := method(cleanFileName)

	if fileErr != nil {
		fmt.Fprintf(os.Stderr, "Could not open file \"%s\"\n", cleanFileName)
		os.Exit(1)
	}

	return file
}
