package main

import (
	"fmt"
	"io"
	"unicode"
)

type ContentLines []string

func filterLines(scanner TextScanner, filterFn func(line int) bool) ContentLines {
	lines := make([]string, 0, 30)

	for line := 1; scanner.Scan(); line++ {
		if filterFn(line) {
			text := scanner.Text()

			// TODO This might not be necessary
			if line == 1 && text[0] == '\xEF' && text[1] == '\xBB' && text[2] == '\xBF' {
				text = text[3:]
			}

			lines = append(lines, text)
		}
	}

	return lines
}

func (lines ContentLines) writeTrimmedTo(output io.Writer) {

	trimAmount := lines.trimAmount()

	for _, str := range lines {
		if len(str) <= trimAmount {
			fmt.Fprintln(output, str)
		} else {
			fmt.Fprintln(output, str[trimAmount:])
		}
	}
}

func (lines ContentLines) trimAmount() int {
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
