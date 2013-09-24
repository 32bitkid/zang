package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os/exec"
	"strings"
)

type (
	ExecGitFn    func(io.Writer, ...string) error
	cachedResult struct {
		reader io.Reader
		error
	}
)

// Retrieve the file contents from git
func (exec ExecGitFn) showFile(result io.Writer, refspec, file string) error {
	fileRef := fmt.Sprintf(`%s:%s`, refspec, file)
	return exec(result, `show`, fileRef)
}

// Retrieve the list of files that have changed between commit1 and commit2
func (exec ExecGitFn) changedFiles(commit1, commit2 string) (map[string]bool, error) {
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

func memoizeExecGitFn(fn ExecGitFn) ExecGitFn {

	cache := make(map[string]cachedResult)

	return func(writer io.Writer, args ...string) error {
		key := strings.Join(args, " ")
		val, exists := cache[key]

		if !exists {
			var buffer bytes.Buffer
			error := fn(&buffer, args...)

			val = cachedResult{&buffer, error}
			cache[key] = val
		}

		io.Copy(writer, val.reader)
		return val.error
	}
}

func execGit(cmdOutput io.Writer, args ...string) error {
	cmd := exec.Command(`git`, args...)

	cmd.Dir = repoFlag

	cmd.Stdout = cmdOutput
	cmd.Stderr = cmdOutput

	return cmd.Run()
}
