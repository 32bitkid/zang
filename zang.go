package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"time"
	"unicode"
)

const (
	startCodeGate  string = "```%s\n"
	endCodeGate    string = "```\n"
	commitRefBlock string = "> Commit: %s  \n"
	fileRefBlock   string = "> File: %s  \n"
	linesRefBlock  string = "> Lines: %d to %d  \n"
	lineRefBlock   string = "> Line: %d  \n"
	staleRefBlock  string = "> *WARNING* This file has changed since the referenced commit. This documentation may be out of date. \n"
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

func main() {
	flag.Parse()

	start := time.Now()
	if err := modeSwitch(flag.Arg(0), flag.Arg(1)); err != nil {
		fmt.Fprintf(os.Stderr, "Error: %s\n", err)
	}

	fmt.Printf("%v", time.Since(start))
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

	dirWalker := func(path string, info os.FileInfo, err error) error {

		if filepath.Ext(path) == ".md" {

			relativePath, relErr := filepath.Rel(inFolder, path)

			if relErr != nil {
				return relErr
			}

			destFile := filepath.Join(outFolder, relativePath)

			os.MkdirAll(filepath.Dir(destFile), os.ModePerm)

			if fileErr := fileMode(path, destFile); fileErr != nil {
				return fileErr
			}
		}
		return err
	}

	return filepath.Walk(inFolder, dirWalker)
}

func fileMode(inFileName, outFileName string) error {
	var (
		inFile  *os.File = safeFile(os.Open, inFileName, os.Stdin)
		outFile *os.File = safeFile(os.Create, outFileName, os.Stdout)
	)

	return processFile(bufio.NewScanner(inFile), bufio.NewWriter(outFile))
}

func processFile(input *bufio.Scanner, output *bufio.Writer) error {
	git := memoizeExecGitFn(execGit)

	for input.Scan() {
		text := input.Text()

		if args, success := parseAsGitCommand(text); success {
			if err := processGit(output, git, args); err != nil {
				return err
			}

			if checkStaleFlag {
				checkGitChanges(output, git, args)
			}

		} else {
			fmt.Fprintln(output, text)
		}
	}

	defer output.Flush()

	return input.Err()
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
