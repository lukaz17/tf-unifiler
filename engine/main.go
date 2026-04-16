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

package engine

import (
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler-go/filesys"
	"github.com/tforceaio/tf-unifiler-go/filesys/exec"
)

var majorVersion = 0
var minorVersion = 5
var patchVersion = 0
var gitCommit, gitDate, gitBranch string

func version() string {
	originDate := time.Date(2024, time.August, 13, 0, 0, 0, 0, time.UTC)
	gitDate2, _ := time.Parse("20060102", gitDate)
	buildDate := opx.Ternary(gitDate == "", time.Now().UTC(), gitDate2)
	duration := buildDate.Sub(originDate)
	minor := minorVersion
	patch := strconv.Itoa(patchVersion)
	if gitBranch == "master" {
		// do nothing
	} else if gitBranch == "release" {
		minor += 1
		patch = patch + "-rc"
	} else if strings.Contains(gitBranch, "feat/") {
		minor += 1
		patch = patch + "-dev"
	} else {
		patch = strconv.Itoa(patchVersion+1) + "-dev"
	}
	if gitCommit != "" {
		return fmt.Sprintf("%d.%d.%s.%d-%s", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000), gitCommit)
	}
	return fmt.Sprintf("%d.%d.%s.%d", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000))
}

// Initialize configurations, loggings for internal modules, and display basic
// information about this invocation.
func InitApp() *Controller {
	cfg := NewController(true)

	filesys.SetLogger(cfg.ModuleLogger("filesystem"))
	exec.SetLogger(cfg.ModuleLogger("exec"))

	pwd, _ := os.Getwd()
	pwd, _ = filesys.GetAbsPath(pwd)
	exec, _ := os.Executable()
	exec, _ = filesys.GetAbsPath(exec)

	cfg.Logger.Info().Msgf("TF UNIFILER v%s", version())
	gitDate2, _ := time.Parse("20060102", gitDate)
	buildDate := opx.Ternary(gitDate == "", time.Now().UTC(), gitDate2)
	cfg.Logger.Info().Msgf("Copyright (C) %d T-Force I/O", buildDate.Year())
	cfg.Logger.Info().Msg("Licensed under GPL-3.0 license. See COPYING file along with this program for more details.")
	cfg.Logger.Info().Msgf("Working directory %s", pwd)
	cfg.Logger.Info().Msgf("Config directory %s", cfg.Root.ConfigDir)
	cfg.Logger.Info().Msgf("Executable file %s", exec)
	cfg.Logger.Info().Msgf("Portable mode %t", cfg.Root.IsPortable)
	cfg.Logger.Info().Msg("-----------------")

	return cfg
}
