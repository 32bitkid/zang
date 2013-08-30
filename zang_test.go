package main

import (
	"bufio"
	"bytes"
	"testing"
)

func TestTrimSpaces(t *testing.T) {
	lines := []string{
		"    This is some",
		"       content",
		"    that needs to be trimmed",
	}

	expected, actual := 4, calculateAmountToTrim(lines)

	if actual != expected {
		t.Errorf("Expected %d got %d", expected, actual)
	}

}

func TestTrimTabs(t *testing.T) {
	lines := []string{
		"\tThis is some",
		"\t\tcontent",
		"\tthat needs to be trimmed",
	}

	expected, actual := 1, calculateAmountToTrim(lines)

	if actual != expected {
		t.Errorf("Expected %d got %d", expected, actual)
	}
}

func TestHandleDanglingIndents(t *testing.T) {
	lines := []string{
		"    This is some",
		"        content",
		"  that needs to be trimmed",
	}

	expected, actual := 2, calculateAmountToTrim(lines)

	if actual != expected {
		t.Errorf("Expected %d got %d", expected, actual)
	}
}

func TestWriteTrimmedLines(t *testing.T) {
	lines := []string{
		"\tThis is some",
		"\t\tcontent",
		"\tthat needs to be trimmed",
	}

	var output bytes.Buffer

	writeTrimmedLines(&output, lines...)

	scanner := bufio.NewScanner(&output)

	// TODO probaby a better way to test this!
	if scanner.Scan(); scanner.Text() != "This is some" {
		t.Error("Ack!")
	}
	if scanner.Scan(); scanner.Text() != "\tcontent" {
		t.Error("Ack!")
	}
	if scanner.Scan(); scanner.Text() != "that needs to be trimmed" {
		t.Error("Ack!")
	}
}

func TestFilterLines(t *testing.T) {
	var output bytes.Buffer
	output.WriteString("line 1\nline 2\nline 3\nline 4")

	scanner := bufio.NewScanner(&output)

	filterFn := func(line int) bool {
		return line > 1 && line < 4
	}

	filteredLines := filterLines(scanner, filterFn)

	if len(filteredLines) != 2 {
		t.Error("Expected only two lines")
	}

	expectedLines := []string { "line 2", "line 3" }

	for i, val := range filteredLines {
		if val != expectedLines[i] {
			t.Errorf("Content was not right. Expected `%s`. Got `%s`", expectedLines[i], val)
		}
	}

}
