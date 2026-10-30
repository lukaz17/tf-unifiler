// Copyright (C) 2025  T-Force I/O
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package engine

import (
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

const (
	sensitiveAnnotation = "unifiler_sensitive" // flag annotation marking a value RedactArgs must mask.
	redactedValue       = "******"             // replaces the value of a sensitive flag.
)

// Mark the flag with given name on cmd as sensitive so RedactArgs masks its value.
func MarkFlagSensitive(cmd *cobra.Command, name string) {
	if err := cmd.Flags().SetAnnotation(name, sensitiveAnnotation, []string{"true"}); err != nil {
		panic(err)
	}
}

// Return a copy of args with the value of every sensitive flag of the target command masked, so args can be logged or reported.
// Flags are read the way pflag parses them; args after "--" and args of an unknown command are kept as-is.
func RedactArgs(args []string) []string {
	redacted := append([]string(nil), args...)
	cmd, _, err := BuildRootCmd().Find(args)
	if err != nil {
		return redacted
	}
	flags := cmd.Flags()
	flags.AddFlagSet(cmd.InheritedFlags())

	for i := 0; i < len(redacted); i++ {
		a := redacted[i]
		if a == "--" {
			break
		}
		f, valueAt := flagValueAt(flags, a)
		if f == nil {
			continue
		}
		_, sensitive := f.Annotations[sensitiveAnnotation]
		switch {
		case valueAt > 0 && sensitive:
			redacted[i] = a[:valueAt] + redactedValue
		case valueAt == 0 && i+1 < len(redacted):
			// The value is the next arg: mask it if needed, and skip it so it is not read as a flag.
			if sensitive {
				redacted[i+1] = redactedValue
			}
			i++
		}
	}
	return redacted
}

// Return the flag in arg a that takes a value, with the index in a where its inline value starts, or 0 when the value is the next arg.
// Return nil when a is positional, sets only bool-like flags, or names an unknown flag.
func flagValueAt(flags *pflag.FlagSet, a string) (*pflag.Flag, int) {
	if len(a) < 2 || a[0] != '-' {
		return nil, 0
	}
	if a[1] == '-' {
		name, _, inline := strings.Cut(a[2:], "=")
		f := flags.Lookup(name)
		switch {
		case f == nil:
			return nil, 0
		case inline:
			return f, len("--") + len(name) + len("=")
		case f.NoOptDefVal != "":
			return nil, 0
		}
		return f, 0
	}
	for j := 1; j < len(a); j++ {
		f := flags.ShorthandLookup(a[j : j+1])
		switch {
		case f == nil:
			return nil, 0
		case j+1 < len(a) && a[j+1] == '=':
			return f, j + 2
		case f.NoOptDefVal != "":
			continue
		case j+1 < len(a):
			return f, j + 1
		}
		return f, 0
	}
	return nil, 0
}
