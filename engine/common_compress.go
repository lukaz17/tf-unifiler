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
	"context"
	"errors"
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/rs/zerolog"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/core/compression"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
	"github.com/tforceaio/tf-unifiler/filesys/exec"
	"github.com/tforceaio/tf-unifiler/internal/nullable"
	"github.com/tforceaio/tf-unifiler/tui"
)

// Pack inputs (files/folders) into an archive.
func createArchive(
	inputs []string, archive string, format string, level compression.Level, solid nullable.Bool, dictSize int, password string, threads nullable.Int, move bool,
	pathConfig *config.PathConfig, logger zerolog.Logger, notifier diag.Notifier, tuiMode bool,
) error {
	p := diag.NewProgressTracker("Pack files", notifier)
	defer p.Done()
	p.Status(archive)

	var args exec.CommandArgs
	var executable string

	dictSizeVal := opx.Ternary(dictSize > 0, strconv.Itoa(dictSize)+"M", "")
	switch format {
	case "7z":
		args = exec.New7zArgs(&exec.X7zOptions{
			ArchiveName:    archive,
			ArchiveType:    "7z",
			Command:        "a",
			CompressLevel:  nullable.FromInt(to7zCompressLevel(level)),
			CompressMethod: "LZMA2",
			DictSize:       dictSizeVal,
			MoveToArchive:  move,
			FileNames:      inputs,
			Password:       password,
			Recurse:        true,
			Solid:          opx.Ternary(solid.IsValid, solid.RealValue, false),
			Threads:        threads,
		})
		executable = pathConfig.X7zPath
	case "rar":
		args = exec.NewRarArgs(&exec.RarOptions{
			ArchiveName:   archive,
			Command:       "a",
			CompressLevel: nullable.FromInt(toRarCompressLevel(level)),
			DictSize:      dictSizeVal,
			MoveToArchive: move,
			FileNames:     inputs,
			Password:      password,
			Recurse:       true,
			Solid:         solid,
			Threads:       threads,
		})
		executable = pathConfig.RarPath
	default:
		return errors.New("unsupported archive format: " + format)
	}

	_, err := exec.Run(executable, args, !tuiMode)
	if err != nil {
		return err
	}

	if !tuiMode {
		logger.Info().
			Str("archive", archive).
			Msg("Packed files.")
	}
	return nil
}

// Derive archive name from input.
func getArchiveName(input string, ext string) string {
	base := filepath.Base(input)
	isDir, err := filesys.IsDirectory(input)
	if err != nil {
		isDir = false
	}

	if isDir {
		return base + ext
	}

	nameWithoutExt := strings.TrimSuffix(base, filepath.Ext(base))
	return nameWithoutExt + ext
}

// Multi-pack inputs (files/folders) into an archives.
func packFiles(
	inputs []string, output string, format string, level compression.Level, solid nullable.Bool, dictSize int, password string, threadNum nullable.Int, move bool, normalize bool, separate bool,
	pathConfig *config.PathConfig, logger zerolog.Logger, notifier diag.Notifier, tuiMode bool,
) ([]string, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if n, ok := notifier.(*tui.BubbleteaNotifier); ok && tuiMode {
		ps := tui.RunProcessStatus(n, cancel)
		defer ps.Stop()
		inputCount := opx.Ternary(separate, len(inputs), 1)
		n.SetTotal(uint64(inputCount))
	}

	var results []string

	if separate {
		for _, input := range inputs {
			if err := ctx.Err(); err != nil {
				return nil, err
			}
			archiveName := getArchiveName(input, "."+format)
			archivePath := filesys.Join(output, archiveName)
			sInput := input
			if isDir, err := filesys.IsDirectory(input); isDir && err == nil {
				sInput += string(os.PathSeparator) + "**"
			}
			if normalize {
				contents, err := filesys.List([]string{input}, true)
				if err != nil {
					return nil, err
				}
				for _, c := range contents {
					if err := normalizeAttributes(c.RelativePath); err != nil {
						return nil, err
					}
				}
			}
			if err := createArchive([]string{sInput}, archivePath, format, level, solid, dictSize, password, threadNum, move, pathConfig, logger, notifier, tuiMode); err != nil {
				return nil, err
			}
			if move {
				if isDir, err := filesys.IsDirectory(input); isDir && err == nil {
					if isEmpty, err := filesys.IsDirectoryEmpty(input); isEmpty && err == nil {
						os.Remove(input)
						logger.Info().Str("path", input).Msg("Removed empty directory.")
					}
				}
			}
			results = append(results, archivePath)
		}
	} else {
		if normalize {
			contents, err := filesys.List(inputs, true)
			if err != nil {
				return nil, err
			}
			for _, c := range contents {
				if err := normalizeAttributes(c.RelativePath); err != nil {
					return nil, err
				}
			}
		}
		if err := createArchive(inputs, output, format, level, solid, dictSize, password, threadNum, move, pathConfig, logger, notifier, tuiMode); err != nil {
			return nil, err
		}
		results = append(results, output)
	}

	return results, nil
}

func to7zCompressLevel(level compression.Level) int {
	switch level {
	case compression.None:
		return 0
	case compression.Fast:
		return 3
	case compression.Normal:
		return 5
	case compression.High:
		return 7
	case compression.Ultra:
		return 9
	default:
		return 5
	}
}

func toRarCompressLevel(level compression.Level) int {
	switch level {
	case compression.None:
		return 0
	case compression.Fast:
		return 2
	case compression.Normal:
		return 3
	case compression.High:
		return 4
	case compression.Ultra:
		return 5
	default:
		return 3
	}
}
