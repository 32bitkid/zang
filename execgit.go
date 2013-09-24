package main

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"strings"
)

type (
	execGitFn    func(*bytes.Buffer, ...string) error
	cachedResult struct {
		string
		error
	}
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

func memoizeExecGitFn(fn execGitFn) execGitFn {

	cache := make(map[string]cachedResult)

	return func(buffer *bytes.Buffer, args ...string) error {
		key := strings.Join(args, " ")
		val, exists := cache[key]

		if exists {
			buffer.WriteString(val.string)
			return val.error
		} else {
			error := fn(buffer, args...)
			cache[key] = cachedResult{buffer.String(), error}
			return error
		}

	}
}

func gitShowFile(result *bytes.Buffer, exec execGitFn, refspec, file string) error {
	fileRef := fmt.Sprintf(`%s:%s`, refspec, file)
	return exec(result, `show`, fileRef)
}

func gitChangedFiles(exec execGitFn, commit1, commit2 string) (map[string]bool, error) {
	var result bytes.Buffer

	lookupTable := make(map[string]bool)

	error := exec(&result, `diff`, `--name-only`, commit1, commit2)

	if error == nil {
		scanner := bufio.NewScanner(&result)
		for scanner.Scan() {
			lookupTable[scanner.Text()] = true
		}

	}

	return lookupTable, error
}

func execGit(cmdOutput *bytes.Buffer, args ...string) error {
	cmd := exec.Command(`git`, args...)

	cmd.Dir = repoFlag

	cmd.Stdout = cmdOutput
	cmd.Stderr = cmdOutput

	return cmd.Run()
}

func processGit(output io.Writer, execGit execGitFn, args *GitCommandArgs) error {
	var cmdOutput bytes.Buffer

	fmt.Fprintln(output, args.source)
	fmt.Fprintln(output, beginMarker)

	defer func() {
		fmt.Fprintln(output, endMarker)
	}()

	if err := gitShowFile(&cmdOutput, execGit, args.refspec, args.file); err == nil {

		fmt.Fprintf(output, startCodeGate, args.format)

		cmdScanner := bufio.NewScanner(&cmdOutput)
		writeTrimmedLines(output, filterLines(cmdScanner, args.displayLine)...)

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
		return errors.New(cmdOutput.String())
	}
}

func checkGitChanges(output io.Writer, git execGitFn, args *GitCommandArgs) bool {

	results, error := gitChangedFiles(git, args.refspec, headFlag)
	if error == nil {
		if _, exists := results[args.file]; exists {
			fmt.Fprintf(os.Stderr, "WARN: \"%s\" has changed since %s. This documentation may be out of date.\n", args.file, args.refspec)
			fmt.Fprintf(output, staleRefBlock)
			return true
		}
	} else {
		fmt.Fprintf(os.Stderr, "WARN: Unable to get history of %s.\n", args.file)
	}
	return false
}
