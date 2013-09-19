package main

import (
	"regexp"
	"strconv"
	"strings"
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

func (args GitCommandArgs) displayLine(line int) bool {
	return !args.hasFrom && !args.hasTo ||
		args.hasFrom && !args.hasTo && line == args.from ||
		args.hasFrom && args.hasTo && line >= args.from && line <= args.to
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
