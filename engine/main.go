// Copyright (C) 2024  T-Force I/O
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
	"fmt"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/filesys"
)

var majorVersion = 0
var minorVersion = 7
var patchVersion = 0
var gitCommit, gitDate, gitBranch string

func version() string {
	originDate := time.Date(2024, time.August, 13, 0, 0, 0, 0, time.UTC)
	gitDate2, err := time.Parse("20060102", gitDate)
	buildDate := opx.Ternary(err == nil, gitDate2, time.Now().UTC())
	duration := buildDate.Sub(originDate)
	minor := minorVersion
	patch := strconv.Itoa(patchVersion)
	if gitBranch == "master" || gitBranch == "HEAD" {
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
	if gitCommit != "" && len(gitCommit) >= 8 {
		return fmt.Sprintf("%d.%d.%s.%d-%s", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000), gitCommit[:8])
	}
	return fmt.Sprintf("%d.%d.%s.%d", majorVersion, minor, patch, duration.Milliseconds()/int64(86400000))
}

// Initialize configurations, loggings for internal modules, and display basic
// information about this invocation.
func InitApp() *Controller {
	ctrl := NewController(true)
	if execViaQueue {
		return ctrl
	}

	pwd, _ := os.Getwd()
	pwd, _ = filesys.GetAbsPath(pwd)
	exec, _ := os.Executable()
	exec, _ = filesys.GetAbsPath(exec)

	fmt.Printf("\nTFunifiler v%s\n", version())
	gitDate2, _ := time.Parse("20060102", gitDate)
	buildDate := opx.Ternary(gitDate == "", time.Now().UTC(), gitDate2)
	fmt.Printf("Copyright (C) %d T-Force I/O\n", buildDate.Year())
	fmt.Printf("Licensed under GPL-3.0 license. See COPYING file along with this program for more details.\n\n")
	ctrl.Logger.Info().Msg("-----------------")
	ctrl.Logger.Info().Msgf("Working directory %s", pwd)
	ctrl.Logger.Info().Msgf("Config directory %s", ctrl.Config.ConfigDir)
	ctrl.Logger.Info().Msgf("Executable file %s", exec)
	ctrl.Logger.Info().Msgf("Portable mode %t", ctrl.Config.IsPortable)
	ctrl.Logger.Info().Msg("-----------------")

	return ctrl
}

// Execute the program.
func Execute() {
	rootCmd := newRootCmd()
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

// Build the root command with all registered subcommands. Callers which
// execute commands programmatically must create a fresh instance per
// invocation to avoid leaking parsed flag states across executions.
func newRootCmd() *cobra.Command {
	gitDate2, _ := time.Parse("20060102", gitDate)
	buildDate := opx.Ternary(gitDate == "", time.Now().UTC(), gitDate2)
	rootCmd := &cobra.Command{
		Use: "unifiler",
		Long: fmt.Sprintf(
			`TFunifiler v%s.
Copyright (C) %d T-Force I/O.
Licensed under GPL-3.0 license. See COPYING file along with this program for more details.`,
			version(),
			buildDate.Year()),
		Short:   "Cross platform file managements command line utility.",
		Version: version(),
	}
	rootCmd.AddCommand(ChecksumCmd())
	rootCmd.AddCommand(ConfigCmd())
	rootCmd.AddCommand(FileCmd())
	rootCmd.AddCommand(MetadataCmd())
	rootCmd.AddCommand(MirrorCmd())
	rootCmd.AddCommand(ServeCmd())
	rootCmd.AddCommand(VideoCmd())
	return rootCmd
}
