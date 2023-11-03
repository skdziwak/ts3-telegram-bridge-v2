package main

import (
	"strings"
	"unicode"
)

func parseCommand(input string) (cmd string, args []string) {
	// Trim the leading slash and any space
	command := strings.TrimLeftFunc(input, func(r rune) bool {
		return !unicode.IsLetter(r)
	})

	// Split the command and arguments
	split := strings.FieldsFunc(command, func(r rune) bool {
		return unicode.IsSpace(r) && r != '"'
	})

	cmd = split[0] // The command is the first element
	var currentArg strings.Builder
	inQuotes := false

	// Iterate over each element to handle quoted arguments
	for _, arg := range split[1:] {
		if strings.HasPrefix(arg, "\"") && strings.HasSuffix(arg, "\"") && len(arg) > 1 {
			// Argument is a single quoted word
			args = append(args, strings.Trim(arg, "\""))
		} else if strings.HasPrefix(arg, "\"") {
			// Beginning of a quoted argument
			currentArg.WriteString(strings.TrimPrefix(arg, "\""))
			inQuotes = true
		} else if strings.HasSuffix(arg, "\"") && inQuotes {
			// End of a quoted argument
			currentArg.WriteString(" ")
			currentArg.WriteString(strings.TrimSuffix(arg, "\""))
			args = append(args, currentArg.String())
			currentArg.Reset()
			inQuotes = false
		} else if inQuotes {
			// Middle of a quoted argument
			currentArg.WriteString(" ")
			currentArg.WriteString(arg)
		} else {
			// Unquoted argument
			args = append(args, arg)
		}
	}

	// Handle the case where the last argument is still in quotes
	if inQuotes {
		args = append(args, currentArg.String())
	}

	return cmd, args
}
