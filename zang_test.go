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
