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

type RarArgs struct {
	Options *RarOptions
}

type RarOptions struct {
	Command string

	ArchiveType    string
	CompressLevel  nullable.Int
	DictSize       string
	EncryptHeaders bool
	MoveToArchive  bool
	Password       string
	Recurse        bool
	Solid          nullable.Bool
	Threads        nullable.Int

	ArchiveName string
	FileNames   []string
}

func (args RarArgs) Compile() []string {
	results := []string{}
	if args.Options.Command != "" {
		results = append(results, args.Options.Command)
	}
	if args.Options.ArchiveType != "" {
		results = append(results, "-ma"+args.Options.ArchiveType)
	}
	if args.Options.CompressLevel.IsValid {
		results = append(results, "-m"+strconv.Itoa(args.Options.CompressLevel.RealValue))
	}
	if args.Options.DictSize != "" {
		results = append(results, "-md"+args.Options.DictSize)
	}
	if args.Options.MoveToArchive {
		results = append(results, "-df")
	}
	if args.Options.Password != "" {
		if args.Options.EncryptHeaders {
			results = append(results, "-hp"+args.Options.Password)
		} else {
			results = append(results, "-p"+args.Options.Password)
		}
	}
	if args.Options.Recurse {
		results = append(results, "-r")
	}
	if args.Options.Solid.IsValid {
		if args.Options.Solid.RealValue {
			results = append(results, "-s")
		} else {
			results = append(results, "-s-")
		}
	}
	if args.Options.Threads.IsValid {
		results = append(results, "-mt"+strconv.Itoa(args.Options.Threads.RealValue))
	}
	results = append(results, "--")
	if args.Options.ArchiveName != "" {
		results = append(results, args.Options.ArchiveName)
	}
	results = append(results, args.Options.FileNames...)
	return results
}

func NewRarArgs(options *RarOptions) RarArgs {
	return RarArgs{options}
}
