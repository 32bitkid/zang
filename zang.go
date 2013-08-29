package main

import (
	"bufio"
	"bytes"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"regexp"
	"strconv"
	"strings"
	"unicode"
)

var repoFlag = flag.String("repo", "", "the path to the repository")

var gitCodeReference *regexp.Regexp = regexp.MustCompile("^\\s*```(\\w+)\\|git\\|(.*?)\\|(.*?):?(\\d+)?:?(\\d+)?```\\s*$")

var commitRefBlock string = `> Commit: %s`
var fileRefBlock string = `> File: %s`
var linesRefBlock string = `> Lines: %d to %d`
var lineRefBlock string = `> Line: %d`

func main() {
	flag.Parse()

	in, out, err := os.Stdin, os.Stdout, os.Stderr
	scanner := bufio.NewScanner(in)

	for scanner.Scan() {
		text := scanner.Text()

		if matches := gitCodeReference.FindStringSubmatch(text); len(matches) > 0 {
			for _, data := range processGit(matches) {
				fmt.Fprintln(out, data)
			}
		} else {
			fmt.Fprintln(out, text) // Println will add back the final '\n'
		}
	}

	if scannerError := scanner.Err(); scannerError != nil {
		fmt.Fprintln(err, "reading standard input:", scannerError)
	}
}

func processGit(parts []string) []string {
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

	var out bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &out

	err := cmd.Run()

	if err == nil {
		var lines []string = make([]string, 0, 30)

		lines = append(lines, fmt.Sprintf("```%s", format))

		lines = append(lines, filterLines(&out, filterFn)...)

		lines = append(lines, "```")
		lines = append(lines, fmt.Sprintf(commitRefBlock, refspec))
		lines = append(lines, fmt.Sprintf(fileRefBlock, file))

		if hasFrom && !hasTo {
			lines = append(lines, fmt.Sprintf(lineRefBlock, from))
		} else if hasFrom && hasTo {
			lines = append(lines, fmt.Sprintf(linesRefBlock, from, to))
		}

		return lines
	}

	return []string{fmt.Sprintf("    Unable to render code: %s", out.String())}
}

func filterLines(out *bytes.Buffer, filterFn func(line int) bool) []string {
	scanner := bufio.NewScanner(out)
	lines := make([]string, 0, 30)

	for line := 1; scanner.Scan(); line++ {
		if filterFn(line) {
			lines = append(lines, scanner.Text())
		}
	}

	return trimLeadingWhitespace(lines)
}

func trimLeadingWhitespace(lines []string) []string {
	ammountToTrim := int(^uint(0) >> 1)

	for _, lineContent := range lines {
		for characterPosition, rune := range lineContent {
			if characterPosition >= ammountToTrim {
				break
			}
			if !unicode.IsSpace(rune) {
				if ammountToTrim > characterPosition {
					ammountToTrim = characterPosition
				}
				break
			}
		}
		if ammountToTrim == 0 {
			return lines
		}
	}

	trimmedLines := make([]string, len(lines))
	for i, str := range lines {
		if len(str) == 0 {
			trimmedLines[i] = str
		} else {
			trimmedLines[i] = str[ammountToTrim:]
		}
	}

	return trimmedLines
}
