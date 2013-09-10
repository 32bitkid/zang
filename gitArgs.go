package main

import (
	"strconv"
	"strings"
)

type GitCommandArgs struct {
	from, to              int
	hasFrom, hasTo        bool
	format, refspec, file string
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
			from:    from,
			to:      to,
			hasFrom: fromErr == nil,
			hasTo:   toErr == nil,
			format:  parts[1],
			refspec: parts[2],
			file:    strings.Replace(parts[3], `\`, `/`, -1),
		}, true

	} else {
		return new(GitCommandArgs), false
	}
}
