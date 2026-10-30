// Copyright (C) 2025 T-Force I/O
// This file is part of TFunifiler
//
// TFunifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TFunifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TFunifiler. If not, see <https://www.gnu.org/licenses/>.

package exec

import (
	"io"
	"os/exec"
)

type CommandArgs interface {
	Compile() []string
}

// Run an app and wait for it to finish, reading its stdin from in (nil means no input).
// A non-nil out receives the app's stdout and stderr as they are written; otherwise its stdout is captured and returned.
func Run(app string, arg CommandArgs, in io.Reader, out io.Writer) (string, error) {
	cmd := exec.Command(app, arg.Compile()...)
	cmd.Stdin = in
	if out != nil {
		cmd.Stdout = out
		cmd.Stderr = out
		err := cmd.Run()
		return "", err
	}

	stdout, err := cmd.Output()
	return string(stdout), err
}
