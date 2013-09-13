package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
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
)

func init() {
	flag.StringVar(&repoFlag, "repo", "", "the path to the repository")
	flag.StringVar(&headFlag, "head", "master", "the commit to check for stale documentation")
	flag.BoolVar(&checkStaleFlag, "check", true, "search the repository for changes since the documentation was written")
}

func safeFile(method func(string) (*os.File, error), fileName string, defaultFile *os.File) *os.File {

	if len(fileName) == 0 {
		return defaultFile
	}

	file, fileErr := method(fileName)

	if fileErr != nil {
		fmt.Fprintf(os.Stderr, "Could not open input file \"%s\"\n", fileName)
		os.Exit(1)
	}

	return file
}

func main() {
	flag.Parse()

	var (
		inFile  *os.File = safeFile(os.Open, flag.Arg(0), os.Stdin)
		outFile *os.File = safeFile(os.Create, flag.Arg(1), os.Stdout)
	)

	err := processFile(bufio.NewScanner(inFile), bufio.NewWriter(outFile))

	if err != nil {
		fmt.Fprintln(os.Stderr, "reading standard input:", err)
	}
}

func processFile(input *bufio.Scanner, output *bufio.Writer) error {
	git := memoizeExecGitFn(execGit)

	for input.Scan() {
		text := input.Text()

		if args, success := parseAsGitCommand(text); success {
			processGit(output, git, args)

			if checkStaleFlag {
				checkGitChanges(output, git, args)
			}

		} else {
			fmt.Fprintln(output, text)
		}
	}

	defer output.Flush()

	if scannerError := input.Err(); scannerError != nil {
		return scannerError
	}

	return nil
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
