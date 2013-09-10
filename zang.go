package main

import (
	"bufio"
	"flag"
	"fmt"
	"io"
	"os"
	"regexp"
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
	repoFlag         string
	headFlag         string
	checkStaleFlag   bool
	gitCodeReference *regexp.Regexp = regexp.MustCompile("^\\s*```(\\w+)\\|git\\|(.*?)\\|(.*?):?(\\d+)?:?(\\d+)?```\\s*$")
)

func init() {
	flag.StringVar(&repoFlag, "repo", "", "the path to the repository")
	flag.StringVar(&headFlag, "head", "master", "the commit to check for stale documentation")
	flag.BoolVar(&checkStaleFlag, "check", true, "search the repository for changes since the documentation was written")
}

func main() {
	flag.Parse()

	var (
		inFile  *os.File = os.Stdin
		outFile *os.File = os.Stdout
		fileErr error
	)

	if len(flag.Arg(0)) > 0 {
		inFile, fileErr = os.Open(flag.Arg(0))

		if fileErr != nil {
			panic(fmt.Sprintf("Could not open input file \"%s\"\n", flag.Arg(0)))
		}
	}

	if len(flag.Arg(1)) > 0 {
		outFile, fileErr = os.Create(flag.Arg(1))

		if fileErr != nil {
			panic(fmt.Sprintf("Could not write to file \"%s\"\n", flag.Arg(1)))
		}
	}

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
		if len(str) == 0 {
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
