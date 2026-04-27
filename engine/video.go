// Copyright (C) 2024 T-Force I/O
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

package engine

import (
	"errors"
	"fmt"
	"math"
	"math/big"
	"path"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
	"github.com/tforceaio/tf-unifiler/filesys/exec"
	"github.com/tforceaio/tf-unifiler/internal/nullable"
)

// VideoModule handles user requests related to batch processing video files.
type VideoModule struct {
	cfg      *config.RootConfig
	logger   zerolog.Logger
	notifier diag.Notifier
}

// Return new VideoModule.
func NewVideoModule(c *Controller, cmdName string) *VideoModule {
	return &VideoModule{
		cfg:      c.Root,
		logger:   c.CommandLogger("video", cmdName),
		notifier: c.Notifier,
	}
}

// Analyze video file and store metadata in JSON format.
func (m *VideoModule) Info(file string) error {
	if err := validateInput(file, "input"); err != nil {
		return err
	}
	m.logger.Info().
		Str("input", file).
		Msg("Start analyzing file information.")

	inputFile, _ := filesys.GetAbsPath(file)
	miFile := inputFile + ".mediainfo.json"
	miOptions := &exec.MediaInfoOptions{
		InputFile:    inputFile,
		OutputFormat: "JSON",
		OutputFile:   miFile,
	}

	stdout, err := exec.Run(m.cfg.Path.MediaInfoPath, exec.NewMediaInfoArgs(miOptions))
	if err != nil {
		return err
	}

	m.logger.Info().
		Str("path", inputFile).
		Msg("Analyzed video file.")
	fmt.Println(stdout)
	m.logger.Info().
		Str("path", miFile).
		Msg("Saved video info.")

	return nil
}

// Take screenshots for videos file from offet of the video file, for a limit duration, every interval.
// All time are in seconds. Quality factor range from 1-100.
func (m *VideoModule) ExtractFrames(file string, interval, offset, limit float64, quality int, outputDir string) error {
	if err := validateInput(file, "input"); err != nil {
		return err
	}
	if outputDir == "" {
		m.logger.Warn().Msg("Output directory is not specified, screenshot will be saved in same directory as input.")
	}
	if interval == 0 {
		m.logger.Warn().Msg("Interval is not specified, default value will be used.")
	}
	if quality == 0 {
		m.logger.Warn().Msg("Quality is not specified, default value will be used.")
	}
	m.logger.Info().
		Str("file", file).
		Floats64("interval/offset/limit", []float64{interval, offset, limit}).
		Str("output", outputDir).
		Msg("Taking screenshot for video file.")

	inputFile, _ := filesys.CreateEntry(file)
	outputRoot := opx.Ternary(outputDir == "", path.Dir(inputFile.AbsolutePath), outputDir)
	if filesys.IsFileExist(outputRoot) {
		return errors.New("a file with same name with target root existed")
	}
	miOptions := &exec.MediaInfoOptions{
		InputFile:    inputFile.AbsolutePath,
		OutputFormat: "JSON",
	}
	stdout, err := exec.Run(m.cfg.Path.MediaInfoPath, exec.NewMediaInfoArgs(miOptions))
	if err != nil {
		return err
	}
	fileMI, _ := exec.DecodeMediaInfoJson(stdout)

	duration, err := strconv.ParseFloat(fileMI.Media.GeneralTracks[0].Duration, 64)
	if err != nil {
		m.logger.Warn().Msg("Invalid video file duration.")
		return err
	}
	limitF64 := opx.Ternary(limit == 0, duration, math.Min(duration, limit))
	limitMs := big.NewInt(int64(limitF64 * float64(1000)))

	if !filesys.IsDirectoryExist(outputRoot) {
		err = filesys.CreateDirectoryRecursive(outputRoot)
		if err != nil {
			return err
		}
	}

	isHDR := fileMI.Media.VideoTracks[0].HDRFormat != ""
	// Convert from BT2020 HDR to BT709 using ffmpeg
	// Reference https://web.archive.org/web/20190722004804/https://stevens.li/guides/video/converting-hdr-to-sdr-with-ffmpeg/
	vfHDR := "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv,format=yuv420p"
	if isHDR {
		m.logger.Info().Str("param", vfHDR).Msg("The video is HDR, Unifiler will attempt to apply colorspace conversion.")
	}
	offsetDef, intervalDef := m.DefaultScreenshotParameter(limitMs)
	offsetMs := opx.Ternary(offset == 0, offsetDef, big.NewInt(int64(offset*1000)))
	intervalMs := opx.Ternary(interval == 0, intervalDef, big.NewInt(int64(interval*1000)))
	qualityFactor := opx.Ternary(quality == 0, 1, quality)
	outputFilenameFormat := opx.Ternary(quality == 1, path.Join(outputRoot, inputFile.Name+"_%s"+".jpg"), path.Join(outputRoot, inputFile.Name+"_%s_q%d"+".jpg"))
	for t := offsetMs; t.Cmp(limitMs) <= 0; t = new(big.Int).Add(t, intervalMs) {
		outFile := opx.Ternary(quality == 1, fmt.Sprintf(outputFilenameFormat, m.ConvertSecondToTimeCode(t)), fmt.Sprintf(outputFilenameFormat, m.ConvertSecondToTimeCode(t), quality))
		ffmOptions := &exec.FFmpegArgsOptions{
			InputFile:      inputFile.AbsolutePath,
			InputStartTime: nullable.FromInt(int(t.Int64()) / 1000),

			OutputFile:       outFile,
			OutputFrameCount: nullable.FromInt(1),
			QualityFactor:    nullable.FromInt(qualityFactor),
			OverwiteOutput:   true,
		}
		if isHDR {
			ffmOptions.VideoFilter = vfHDR
		}

		_, err := exec.Run(m.cfg.Path.FFMpegPath, exec.NewFFmpegArgs(ffmOptions))
		if err != nil {
			m.logger.Info().Msg("Failed to take video screenshot.")
			return err
		}
		log.Info().
			Float64("time", float64(t.Int64())/float64(1000)).
			Str("output", outFile).
			Msg("Created screenshot.")
	}

	return nil
}

// Return timecode string from timeMs in miliseconds.
func (m *VideoModule) ConvertSecondToTimeCode(timeMs *big.Int) string {
	hr := new(big.Int).Div(timeMs, big.NewInt(3600000))
	timeMs = new(big.Int).Mod(timeMs, big.NewInt(3600000))
	mm := new(big.Int).Div(timeMs, big.NewInt(60000))
	timeMs = new(big.Int).Mod(timeMs, big.NewInt(60000))
	sc := new(big.Int).Div(timeMs, big.NewInt(1000))
	ms := new(big.Int).Mod(timeMs, big.NewInt(1000))

	return fmt.Sprintf("%d_%02d_%02d_%03d", hr.Int64(), mm.Int64(), sc.Int64(), ms.Int64())
}

// Return offset/interval paramteters for video with lengthMs in miliseconds.
func (m *VideoModule) DefaultScreenshotParameter(lengthMs *big.Int) (*big.Int, *big.Int) {
	defaults := []struct {
		duration *big.Int
		offset   *big.Int
		interval *big.Int
	}{
		{big.NewInt(120), big.NewInt(1000), big.NewInt(2500)},       // 0 -> 47
		{big.NewInt(420000), big.NewInt(1300), big.NewInt(4300)},    // 27 -> 97
		{big.NewInt(1200000), big.NewInt(1700), big.NewInt(7100)},   // 59 -> 168
		{big.NewInt(3600000), big.NewInt(2300), big.NewInt(12300)},  // 97 -> 292
		{big.NewInt(10800000), big.NewInt(2700), big.NewInt(12700)}, // 283 -> 850
	}
	for _, d := range defaults {
		if lengthMs.Cmp(d.duration) < 0 {
			return d.offset, d.interval
		}
	}
	return big.NewInt(3400), big.NewInt(17100) // 631 -> max
}

// Decorator to log error occurred when calling handlers.
func (m *VideoModule) logError(err error) {
	logProgramError(m.logger, err)
}

// Define Cobra Command for Video module.
func VideoCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "video",
		Short: "Batch processing video file.",
	}
	rootCmd.PersistentFlags().StringP("file", "i", "", "Input video file.")

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Analyze video file.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseVideoFlags(cmd, args)
			m := NewVideoModule(c, "info")
			if err := validateSingleInput(flags.Inputs); err != nil {
				m.logError(err)
				return
			}
			m.logError(m.Info(flags.Inputs[0]))
		},
	}
	rootCmd.AddCommand(infoCmd)

	extractFramesCmd := &cobra.Command{
		Use:   "extract-frames",
		Short: "Extract multiple frames in video file.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseVideoFlags(cmd, args)
			m := NewVideoModule(c, "screenshot")
			if err := validateSingleInput(flags.Inputs); err != nil {
				m.logError(err)
				return
			}
			m.logError(m.ExtractFrames(flags.Inputs[0], flags.Interval, flags.Offset, flags.Limit, flags.Quality, flags.OutputDir))
		},
	}
	extractFramesCmd.Flags().IntP("quality", "q", 90, "Quality factor for screenshot in scale 1-100.")
	extractFramesCmd.Flags().StringP("output", "o", "", "Directory to save screenshots.")
	rootCmd.AddCommand(extractFramesCmd)

	return rootCmd
}

// Struct VideoFlags contains all flags used by Video module.
type VideoFlags struct {
	Inputs    []string
	Interval  float64
	Limit     float64
	Offset    float64
	OutputDir string
	Quality   int
}

// Extract all flags from a Cobra Command.
func ParseVideoFlags(cmd *cobra.Command, args []string) *VideoFlags {
	file, _ := cmd.Flags().GetString("file")
	inputs := args
	if file != "" {
		inputs = append(inputs, file)
	}
	interval, _ := cmd.Flags().GetFloat64("interval")
	limit, _ := cmd.Flags().GetFloat64("limit")
	offset, _ := cmd.Flags().GetFloat64("offset")
	outputDir, _ := cmd.Flags().GetString("output")
	quality, _ := cmd.Flags().GetInt("quality")

	return &VideoFlags{
		Inputs:    inputs,
		Interval:  interval,
		Limit:     limit,
		Offset:    offset,
		OutputDir: outputDir,
		Quality:   quality,
	}
}
