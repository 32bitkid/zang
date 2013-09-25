package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"regexp"
	"strconv"
	"strings"
)

const (
	startCodeGate  string = "```%s\n"
	endCodeGate    string = "```\n"
	commitRefBlock string = "> Commit: %s  \n"
	fileRefBlock   string = "> File: %s  \n"
	linesRefBlock  string = "> Lines: %d to %d  \n"
	lineRefBlock   string = "> Line: %d  \n"
	staleRefBlock  string = "> *WARNING* This file has changed since the referenced commit. This documentation may be out of date. \n"
	beginMarker    string = "<!-- Begin generated code reference. DO NOT EDIT -->"
	endMarker      string = "<!-- End generated code reference. -->"
)

var (
	gitCodeReference *regexp.Regexp = regexp.MustCompile("^\\s*<!--\\s*\\{\\{(\\w+)\\|git\\|(.*?)\\|(.*?):?(\\d+)?:?(\\d+)?\\}\\}\\s*-->\\s*$")
)

type GitCommandArgs struct {
	from, to              int
	hasFrom, hasTo        bool
	format, refspec, file string
	source                string
}

func (args *GitCommandArgs) displayLine(line int) bool {
	return !args.hasFrom && !args.hasTo ||
		args.hasFrom && !args.hasTo && line == args.from ||
		args.hasFrom && args.hasTo && line >= args.from && line <= args.to
}

func (args *GitCommandArgs) process(output io.Writer, execGit ExecGitFn) error {
	var cmdOutput bytes.Buffer

	fmt.Fprintln(output, args.source)
	fmt.Fprintln(output, beginMarker)

	defer func() {
		fmt.Fprintln(output, endMarker)
	}()

	if err := execGit.showFile(&cmdOutput, args.refspec, args.file); err == nil {

		fmt.Fprintf(output, startCodeGate, args.format)

		cmdScanner := bufio.NewScanner(&cmdOutput)
		filterLines(cmdScanner, args.displayLine).writeTrimmedTo(output)

		fmt.Fprint(output, endCodeGate)
		fmt.Fprintf(output, commitRefBlock, args.refspec)
		fmt.Fprintf(output, fileRefBlock, args.file)

		if args.hasFrom && !args.hasTo {
			fmt.Fprintf(output, lineRefBlock, args.from)
		} else if args.hasFrom && args.hasTo {
			fmt.Fprintf(output, linesRefBlock, args.from, args.to)
		}
		return nil
	} else {
		return errors.New(strings.TrimRight(cmdOutput.String(), "\r\n"))
	}
}

func (args *GitCommandArgs) checkGitChanges(output io.Writer, git ExecGitFn) bool {

	results, error := git.changedFiles(args.refspec, headFlag)
	if error == nil {
		if _, exists := results[args.file]; exists {
			fmt.Fprintf(os.Stderr, "WARN: \"%s\" has changed since %s. This documentation may be out of date.\n", args.file, args.refspec)
			fmt.Fprintf(output, staleRefBlock)
			return true
		}
	} else {
		fmt.Fprintf(os.Stderr, "WARN: Unable to get history of \"%s\".\n", args.file)
	}
	return false
}

func parseAsGitCommand(text string) (*GitCommandArgs, bool) {
	if parts := gitCodeReference.FindStringSubmatch(text); len(parts) > 0 {

		from, fromErr := strconv.Atoi(parts[4])
		to, toErr := strconv.Atoi(parts[5])

		return &GitCommandArgs{
			from,
			to,
			fromErr == nil,
			toErr == nil,
			parts[1],
			parts[2],
			strings.Replace(parts[3], `\`, `/`, -1),
			text,
		}, true

	} else {
		return new(GitCommandArgs), false
	}
}
