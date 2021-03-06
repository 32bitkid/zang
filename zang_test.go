package main

import (
	"bufio"
	"bytes"
	"io"
	"testing"
)

var testFile = "line 1\n\tline 2\n\tline 3\nline 4"

func TestTrimSpaces(t *testing.T) {
	lines := ContentLines{
		"    This is some",
		"       content",
		"    that needs to be trimmed",
	}

	expected, actual := 4, lines.trimAmount()

	if actual != expected {
		t.Errorf("Expected %d got %d", expected, actual)
	}

}

func TestTrimTabs(t *testing.T) {
	lines := ContentLines{
		"\tThis is some",
		"\t\tcontent",
		"\tthat needs to be trimmed",
	}

	expected, actual := 1, lines.trimAmount()

	if actual != expected {
		t.Errorf("Expected %d got %d", expected, actual)
	}
}

func TestHandleDanglingIndents(t *testing.T) {
	lines := ContentLines{
		"    This is some",
		"        content",
		"  that needs to be trimmed",
	}

	expected, actual := 2, lines.trimAmount()

	if actual != expected {
		t.Errorf("Expected %d got %d", expected, actual)
	}
}

func TestWriteTrimmedLines(t *testing.T) {
	lines := ContentLines{
		"\tThis is some",
		"\t\tcontent",
		"\tthat needs to be trimmed",
	}

	var output bytes.Buffer

	lines.writeTrimmedTo(&output)

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
	output.WriteString(testFile)

	scanner := bufio.NewScanner(&output)

	filterFn := func(line int) bool {
		return line > 1 && line < 4
	}

	filteredLines := filterLines(scanner, filterFn)

	if len(filteredLines) != 2 {
		t.Error("Expected only two lines")
	}

	expectedLines := []string{"\tline 2", "\tline 3"}

	for i, val := range filteredLines {
		if val != expectedLines[i] {
			t.Errorf("Content was not right. Expected `%s`. Got `%s`", expectedLines[i], val)
		}
	}
}

func TestProcessGitFullFile(t *testing.T) {
	var output bytes.Buffer
	execGit := func(cmdOutput io.Writer, args ...string) error {
		io.WriteString(cmdOutput, testFile)
		return nil
	}

	line := "<!-- {{csharp|git|developing|file.txt}} -->"
	args, success := parseAsGitCommand(line)

	if !success {
		t.Errorf("Parse was not parsed correctly!")
		return
	}

	args.process(&output, execGit)

	// TODO this could probably be done better
	expected := "<!-- {{csharp|git|developing|file.txt}} -->\n<!-- Begin generated code reference. DO NOT EDIT -->\n```csharp\nline 1\n\tline 2\n\tline 3\nline 4\n```\n> Commit: developing  \n> File: file.txt  \n<!-- End generated code reference. -->\n"

	if actual := output.String(); actual != expected {
		t.Errorf("Expected result was not correct\n<<< Actual \n%q\n---\n%q\n>>> Expected\n", actual, expected)
		return
	}
}

func TestProcessGitRange(t *testing.T) {
	var output bytes.Buffer
	execGit := func(cmdOutput io.Writer, args ...string) error {
		io.WriteString(cmdOutput, testFile)
		return nil
	}

	line := "<!-- {{csharp|git|developing|file.txt:2:3}} -->"
	args, success := parseAsGitCommand(line)

	if !success {
		t.Errorf("Parse was not parsed correctly!")
		return
	}

	args.process(&output, execGit)

	// TODO this could probably be done better
	expected := "<!-- {{csharp|git|developing|file.txt:2:3}} -->\n<!-- Begin generated code reference. DO NOT EDIT -->\n```csharp\nline 2\nline 3\n```\n> Commit: developing  \n> File: file.txt  \n> Lines: 2 to 3  \n<!-- End generated code reference. -->\n"

	if actual := output.String(); actual != expected {
		t.Errorf("Expected result was not correct\n<<<\n%q\n---\n%q\n>>>\n", actual, expected)
	}
}

func TestProcessGitSingleLine(t *testing.T) {
	var output bytes.Buffer
	execGit := func(cmdOutput io.Writer, args ...string) error {
		io.WriteString(cmdOutput, testFile)
		return nil
	}

	line := "<!-- {{csharp|git|developing|file.txt:2}} -->"
	args, success := parseAsGitCommand(line)

	if !success {
		t.Errorf("Parse was not parsed correctly!")
		return
	}

	args.process(&output, execGit)

	// TODO this could probably be done better
	expected := "<!-- {{csharp|git|developing|file.txt:2}} -->\n<!-- Begin generated code reference. DO NOT EDIT -->\n```csharp\nline 2\n```\n> Commit: developing  \n> File: file.txt  \n> Line: 2  \n<!-- End generated code reference. -->\n"

	if actual := output.String(); actual != expected {
		t.Errorf("Expected result was not correct\n<<<\n%q\n---\n%q\n>>>\n", actual, expected)
	}
}
