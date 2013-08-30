package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var repoFlag = flag.String("repo", "", "the path to the repository")

var gitCodeReference *regexp.Regexp = regexp.MustCompile("^\\s*```(\\w+)\\|git\\|(.*?)\\|(.*?):?(\\d+)?:?(\\d+)?```\\s*$")

var startCodeGate string = "```%s\n"
var endCodeGate string = "```\n"
var commitRefBlock string = "> Commit: %s\n"
var fileRefBlock string = "> File: %s\n"
var linesRefBlock string = "> Lines: %d to %d\n"
var lineRefBlock string = "> Line: %d\n"

func main() {
	flag.Parse()

	input, output, err := bufio.NewScanner(os.Stdin), bufio.NewWriter(os.Stdout), os.Stderr

	for input.Scan() {
		text := input.Text()

		if matches := gitCodeReference.FindStringSubmatch(text); len(matches) > 0 {
			processGit(output, matches)
		} else {
			fmt.Fprintln(output, text)
		}
	}

	if scannerError := input.Err(); scannerError != nil {
		fmt.Fprintln(err, "reading standard input:", scannerError)
	}

	output.Flush()
}

func processGit(output io.Writer, parts []string) {
	format, refspec, file := parts[1], parts[2], parts[3]
	from, fromErr := strconv.Atoi(parts[4])
	to, toErr := strconv.Atoi(parts[5])
	hasFrom, hasTo := fromErr == nil, toErr == nil

	filterFn := func(line int) bool {
		return !hasFrom && !hasTo ||
			hasFrom && !hasTo && line == from ||
			hasFrom && hasTo && line >= from && line <= to
	}

	gitArgs := fmt.Sprintf(`%s:%s`, refspec, strings.Replace(file, `\`, `/`, -1))

	cmd := exec.Command(`git`, `show`, gitArgs)
	cmd.Dir = *repoFlag

	var cmdOutput bytes.Buffer
	cmd.Stdout = &cmdOutput
	cmd.Stderr = &cmdOutput

	err := cmd.Run()

	if err == nil {
		fmt.Fprintf(output, startCodeGate, format)

		cmdScanner := bufio.NewScanner(&cmdOutput)
		writeTrimmedLines(output, filterLines(cmdScanner, filterFn)...)

		fmt.Fprint(output, endCodeGate)
		fmt.Fprintf(output, commitRefBlock, refspec)
		fmt.Fprintf(output, fileRefBlock, file)

		if hasFrom && !hasTo {
			fmt.Fprintf(output, lineRefBlock, from)
		} else if hasFrom && hasTo {
			fmt.Fprintf(output, linesRefBlock, from, to)
		}
	} else {
		fmt.Fprintf(output, "    Unable to render code: %s\n", cmdOutput.String())
	}
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
