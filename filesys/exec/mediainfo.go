// Copyright (C) 2024 T-Force I/O
// This file is part of TF Unifiler
//
// TF Unifiler is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// TF Unifiler is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE. See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with TF Unifiler. If not, see <https://www.gnu.org/licenses/>.

package exec

import (
	"fmt"
)

type MediaInfoArgs struct {
	Options *MediaInfoOptions
}

type MediaInfoOptions struct {
	InputFile    string
	OutputFormat string
	OutputFile   string
}

func (args MediaInfoArgs) Compile() []string {
	results := []string{}
	if args.Options.OutputFormat != "" {
		results = append(results, fmt.Sprintf("--output=%s", args.Options.OutputFormat))
	}
	if args.Options.OutputFile != "" {
		results = append(results, fmt.Sprintf("--logfile=%s", args.Options.OutputFile))
	}
	results = append(results, args.Options.InputFile)
	return results
}

func NewMediaInfoArgs(options *MediaInfoOptions) MediaInfoArgs {
	return MediaInfoArgs{options}
}
