package main

import "testing"

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
