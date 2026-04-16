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
	"strconv"

	"github.com/tforceaio/tf-unifiler-go/x/nullable"
)

type FFmpegArgs struct {
	Options *FFmpegArgsOptions
}

type FFmpegArgsOptions struct {
	InputFile      string
	InputStartTime nullable.Int

	OutputFile       string
	OutputFrameCount nullable.Int
	OutputStartTime  nullable.Int
	QualityFactor    nullable.Int
	VideoFilter      string
	OverwiteOutput   bool
}

func (args FFmpegArgs) Compile() []string {
	results := []string{}
	if args.Options.InputStartTime.IsValid {
		results = append(results, "-ss", strconv.Itoa(args.Options.InputStartTime.RealValue))
	}
	if args.Options.InputFile != "" {
		results = append(results, "-i", args.Options.InputFile)
	}
	if args.Options.OutputStartTime.IsValid {
		results = append(results, "-ss", strconv.Itoa(args.Options.OutputStartTime.RealValue))
	}
	if args.Options.OutputFrameCount.IsValid {
		results = append(results, "-frames", strconv.Itoa(args.Options.OutputFrameCount.RealValue))
	}
	if args.Options.QualityFactor.IsValid {
		results = append(results, "-q", strconv.Itoa(args.Options.QualityFactor.RealValue))
	}
	if args.Options.VideoFilter != "" {
		results = append(results, "-vf", args.Options.VideoFilter)
	}
	if args.Options.OverwiteOutput {
		results = append(results, "-y")
	}
	if args.Options.OutputFile != "" {
		results = append(results, args.Options.OutputFile)
	}
	return results
}

func NewFFmpegArgs(options *FFmpegArgsOptions) FFmpegArgs {
	return FFmpegArgs{options}
}
