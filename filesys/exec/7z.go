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
	"strconv"

	"github.com/tforceaio/tf-unifiler/internal/nullable"
)

type X7zArgs struct {
	Options *X7zOptions
}

type X7zOptions struct {
	Command string

	ArchiveType    string
	CompressLevel  nullable.Int
	CompressMethod string
	DictSize       string
	EncryptHeaders bool
	MoveToArchive  bool
	Password       string
	Recurse        bool
	Solid          bool
	SolidFlags     string
	Threads        nullable.Int

	ArchiveName string
	FileNames   []string
}

func (args X7zArgs) Compile() []string {
	results := []string{}
	if args.Options.Command != "" {
		results = append(results, args.Options.Command)
	}
	if args.Options.ArchiveType != "" {
		results = append(results, "-t"+args.Options.ArchiveType)
	}
	if args.Options.CompressLevel.IsValid {
		results = append(results, "-mx"+strconv.Itoa(args.Options.CompressLevel.RealValue))
		results = append(results, "-myx"+strconv.Itoa(args.Options.CompressLevel.RealValue))
	}
	if args.Options.CompressMethod != "" {
		results = append(results, "-m0="+args.Options.CompressMethod)
	}
	if args.Options.DictSize != "" {
		results = append(results, "-md"+args.Options.DictSize)
	}
	if args.Options.EncryptHeaders {
		results = append(results, "-mhe=on")
	}
	if args.Options.MoveToArchive {
		results = append(results, "-sdel")
	}
	if args.Options.Password != "" {
		results = append(results, "-p"+args.Options.Password)
	}
	if args.Options.Recurse {
		results = append(results, "-r")
	}
	if args.Options.SolidFlags != "" {
		results = append(results, "-ms="+args.Options.SolidFlags)
	} else if args.Options.Solid {
		results = append(results, "-ms=on")
	}
	if args.Options.Threads.IsValid {
		results = append(results, "-mmt"+strconv.Itoa(args.Options.Threads.RealValue))
	}
	results = append(results, "--")
	if args.Options.ArchiveName != "" {
		results = append(results, args.Options.ArchiveName)
	}
	results = append(results, args.Options.FileNames...)
	return results
}

func New7zArgs(options *X7zOptions) X7zArgs {
	return X7zArgs{options}
}
