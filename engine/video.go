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
	"errors"
	"fmt"
	"math"
	"path"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler-go/config"
	"github.com/tforceaio/tf-unifiler-go/filesystem"
	"github.com/tforceaio/tf-unifiler-go/filesystem/exec"
	"github.com/tforceaio/tf-unifiler-go/media"
	"github.com/tforceaio/tf-unifiler-go/x/nullable"
)

// VideoModule handles user requests related to batch processing of video files.
type VideoModule struct {
	cfg    *config.RootConfig
	logger zerolog.Logger
}

// Return new VideoModule.
func NewVideoModule(c *Controller, cmdName string) *VideoModule {
	return &VideoModule{
		cfg:    c.Root,
		logger: c.CommandLogger("video", cmdName),
	}
}

// Analyze video file and store metadata in JSON format.
func (m *VideoModule) Info(file string) error {
	if file == "" {
		return errors.New("input is not set")
	} else if !filesystem.IsFileExist(file) {
		return errors.New("input file not found")
	}
	m.logger.Info().
		Str("input", file).
		Msg("Start analyzing file information.")

	inputFile, _ := filesystem.GetAbsPath(file)
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
func (m *VideoModule) Screenshot(file string, interval, offset, limit float64, quality int, outputDir string) error {
	if file == "" {
		return errors.New("input is not set")
	} else if !filesystem.IsFileExist(file) {
		return errors.New("input file not found")
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

	inputFile, _ := filesystem.CreateEntry(file)
	outputRoot := opx.Ternary(outputDir == "", path.Dir(inputFile.AbsolutePath), outputDir)
	if filesystem.IsFileExist(outputRoot) {
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
	limitMs := uint64(limitF64 * float64(1000))

	if !filesystem.IsDirectoryExist(outputRoot) {
		err = filesystem.CreateDirectoryRecursive(outputRoot)
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
	offsetMs := opx.Ternary(offset == 0, offsetDef, uint64(offset*1000))
	intervalMs := opx.Ternary(interval == 0, intervalDef, uint64(interval*1000))
	qualityFactor := opx.Ternary(quality == 0, 1, quality)
	outputFilenameFormat := opx.Ternary(quality <= 2, path.Join(outputRoot, inputFile.Name+"_%s"+".jpg"), path.Join(outputRoot, inputFile.Name+"_%s_q%d"+".jpg"))
	for t := offsetMs; t <= limitMs; t = t + intervalMs {
		outFile := opx.Ternary(quality == 1, fmt.Sprintf(outputFilenameFormat, m.ConvertSecondToTimeCode(t)), fmt.Sprintf(outputFilenameFormat, m.ConvertSecondToTimeCode(t), quality))
		ffmOptions := &exec.FFmpegArgsOptions{
			InputFile:      inputFile.AbsolutePath,
			InputStartTime: nullable.FromInt(int(t)),

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
			Float64("time", float64(t)/1000).
			Str("output", outFile).
			Msg("Created screenshot.")
	}

	return nil
}

// Return timecode string from timeMs in miliseconds.
func (m *VideoModule) ConvertSecondToTimeCode(timeMs uint64) string {
	hr := timeMs / 3600000
	timeMs = timeMs % 3600000
	mm := timeMs / 60000
	timeMs = timeMs % 60000
	sc := timeMs / 1000
	ms := timeMs % 1000

	return fmt.Sprintf("%d_%02d_%02d_%03d", hr, mm, sc, ms)
}

// Return offset/interval paramteters for video with lengthMs in miliseconds.
func (m *VideoModule) DefaultScreenshotParameter(lengthMs uint64) (uint64, uint64) {
	defaults := []struct {
		duration   uint64
		offset     uint64
		interval   uint64
		xrf        uint64
		breakpoint uint64
	}{
		{10800000, 5000, 24000, 945, 450},
		{2700000, 3000, 15000, 450, 180},
		{15000, 2000, 10000, 340, 10},
	}
	for _, d := range defaults {
		if lengthMs > d.duration {
			return d.offset, media.DurationToInterval(lengthMs, d.interval, d.xrf, d.breakpoint)
		}
	}
	return 3, lengthMs / 2 // capture of middle of clip
}

// Decorator to log error occurred when calling handlers.
func (m *VideoModule) logError(err error) {
	if err != nil {
		m.logger.Err(err).Msg("Unexpected error has occurred. Program will exit.")
	}
}

// Define Cobra Command for Video module.
func VideoCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "video",
		Short: "Batch video files processing.",
	}

	infoCmd := &cobra.Command{
		Use:   "info",
		Short: "Analyze video file and store metadata in JSON format.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseVideoFlags(cmd)
			m := NewVideoModule(c, "info")
			m.logError(m.Info(flags.File))
		},
	}
	infoCmd.Flags().StringP("file", "i", "", "Video file to generate info.")
	rootCmd.AddCommand(infoCmd)

	screenshotCmd := &cobra.Command{
		Use:   "screenshot",
		Short: "Scan and compute hashes files/directories then create hardlink to workspace.",
		Run: func(cmd *cobra.Command, args []string) {
			c := InitApp()
			defer c.Close()
			flags := ParseVideoFlags(cmd)
			m := NewVideoModule(c, "screenshot")
			m.logError(m.Screenshot(flags.File, flags.Interval, flags.Offset, flags.Limit, flags.Quality, flags.OutputDir))
		},
	}
	screenshotCmd.Flags().StringP("file", "i", "", "Video file to take screenshot.")
	screenshotCmd.Flags().IntP("quality", "q", 1, "Quality factor for screenshot in scale 2-31.")
	screenshotCmd.Flags().StringP("output", "o", "", "Directory to save screenshots.")
	rootCmd.AddCommand(screenshotCmd)

	return rootCmd
}

// Struct VideoFlags contains all flags used by Video module.
type VideoFlags struct {
	File      string
	Interval  float64
	Limit     float64
	Offset    float64
	OutputDir string
	Quality   int
}

// Extract all flags from a Cobra Command.
func ParseVideoFlags(cmd *cobra.Command) *VideoFlags {
	file, _ := cmd.Flags().GetString("file")
	interval, _ := cmd.Flags().GetFloat64("interval")
	limit, _ := cmd.Flags().GetFloat64("limit")
	offset, _ := cmd.Flags().GetFloat64("offset")
	outputDir, _ := cmd.Flags().GetString("output")
	quality, _ := cmd.Flags().GetInt("quality")

	return &VideoFlags{
		File:      file,
		Interval:  interval,
		Limit:     limit,
		Offset:    offset,
		OutputDir: outputDir,
		Quality:   quality,
	}
}
