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
	"context"
	"errors"
	"fmt"
	"math"
	"math/big"
	"path/filepath"
	"strconv"

	"github.com/rs/zerolog"
	"github.com/spf13/cobra"
	"github.com/tforce-io/tf-golib/opx"
	"github.com/tforceaio/tf-unifiler/config"
	"github.com/tforceaio/tf-unifiler/diag"
	"github.com/tforceaio/tf-unifiler/filesys"
	"github.com/tforceaio/tf-unifiler/filesys/exec"
	"github.com/tforceaio/tf-unifiler/internal/nullable"
	"github.com/tforceaio/tf-unifiler/tui"
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
		cfg:      c.Config,
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
		Str("input", filesys.NormalizePath(file, true)).
		Msg("Start analyzing file information.")

	inputFile, err := filesys.GetAbsPath(file)
	if err != nil {
		return fmt.Errorf("failed to resolve input path: %w", err)
	}
	miFile := inputFile + ".mediainfo.json"
	miOptions := &exec.MediaInfoOptions{
		InputFile:    inputFile,
		OutputFormat: "JSON",
		OutputFile:   miFile,
	}

	_, err = exec.Run(m.cfg.Path.MediaInfoPath, exec.NewMediaInfoArgs(miOptions), nil, nil)
	if err != nil {
		return err
	}

	m.logger.Info().
		Str("path", filesys.NormalizePath(inputFile, true)).
		Msg("Analyzed video file.")
	m.logger.Info().
		Str("path", filesys.NormalizePath(miFile, true)).
		Msg("Saved video info.")

	return nil
}

// Take screenshots for videos file from offset of the video file, for a limit duration, every interval.
// All time are in seconds. Quality is FFmpeg's JPEG scale, from 1 (best) to 31.
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
		Str("file", filesys.NormalizePath(file, true)).
		Floats64("interval/offset/limit", []float64{interval, offset, limit}).
		Str("output", filesys.NormalizePath(outputDir, true)).
		Msg("Extracting frames for video file.")

	inputFile, err := filesys.CreateEntry(file)
	if err != nil {
		return fmt.Errorf("failed to read input file: %w", err)
	}
	outputRoot := opx.Ternary(outputDir == "", filepath.Dir(inputFile.AbsolutePath), outputDir)
	if filesys.IsFileExist(outputRoot) {
		return errors.New("a file with same name with target root existed")
	}

	count, err := m.extractFrames(inputFile, outputRoot, interval, offset, limit, quality)
	if err != nil {
		return err
	}
	m.logger.Info().
		Uint64("count", count).
		Str("file", filesys.NormalizePath(file, true)).
		Str("output", filesys.NormalizePath(outputDir, true)).
		Msg("Extracted frames for video file.")
	return nil
}

// Return seconds in milliseconds.
func (m *VideoModule) convertSecondToMs(seconds float64) *big.Int {
	// Rounded, since seconds*1000 can land just below a whole millisecond, such as 1.001 to 1000.9999999999999.
	return big.NewInt(int64(math.Round(seconds * 1000)))
}

// Return timecode string from timeMs in milliseconds.
func (m *VideoModule) convertSecondToTimeCode(timeMs *big.Int) string {
	hr := new(big.Int).Div(timeMs, big.NewInt(3600000))
	timeMs = new(big.Int).Mod(timeMs, big.NewInt(3600000))
	mm := new(big.Int).Div(timeMs, big.NewInt(60000))
	timeMs = new(big.Int).Mod(timeMs, big.NewInt(60000))
	sc := new(big.Int).Div(timeMs, big.NewInt(1000))
	ms := new(big.Int).Mod(timeMs, big.NewInt(1000))

	return fmt.Sprintf("%d_%02d_%02d_%03d", hr.Int64(), mm.Int64(), sc.Int64(), ms.Int64())
}

// Return offset/interval parameters for video with lengthMs in milliseconds.
func (m *VideoModule) defaultScreenshotParameter(lengthMs *big.Int) (*big.Int, *big.Int) {
	defaults := []struct {
		duration *big.Int
		offset   *big.Int
		interval *big.Int
	}{
		{big.NewInt(120000), big.NewInt(1000), big.NewInt(2500)},    // 0 -> 47
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

func (m *VideoModule) extractFrames(inputFile *filesys.FsEntry, outputRoot string, interval, offset, limit float64, quality int) (uint64, error) {
	miOptions := &exec.MediaInfoOptions{
		InputFile:    inputFile.AbsolutePath,
		OutputFormat: "JSON",
	}
	stdout, err := exec.Run(m.cfg.Path.MediaInfoPath, exec.NewMediaInfoArgs(miOptions), nil, nil)
	if err != nil {
		return 0, err
	}
	fileMI, err := exec.DecodeMediaInfoJson(stdout)
	if err != nil {
		return 0, err
	}
	if len(fileMI.Media.GeneralTracks) == 0 {
		return 0, errors.New("media info has no general track")
	}
	if len(fileMI.Media.VideoTracks) == 0 {
		return 0, errors.New("media info has no video track")
	}

	duration, err := strconv.ParseFloat(fileMI.Media.GeneralTracks[0].Duration, 64)
	if err != nil {
		m.logger.Warn().Msg("Invalid video file duration.")
		return 0, err
	}
	limitF64 := opx.Ternary(limit == 0, duration, math.Min(duration, limit))
	limitMs := m.convertSecondToMs(limitF64)

	if !filesys.IsDirectoryExist(outputRoot) {
		err = filesys.CreateDirectoryRecursive(outputRoot)
		if err != nil {
			return 0, err
		}
	}

	isHDR := fileMI.Media.VideoTracks[0].HDRFormat != ""
	// Convert from BT2020 HDR to BT709 using ffmpeg
	// Reference https://web.archive.org/web/20190722004804/https://stevens.li/guides/video/converting-hdr-to-sdr-with-ffmpeg/
	vfHDR := "zscale=t=linear:npl=100,format=gbrpf32le,zscale=p=bt709,tonemap=tonemap=hable:desat=0,zscale=t=bt709:m=bt709:r=tv,format=yuv420p"
	if isHDR {
		m.logger.Info().Str("param", vfHDR).Msg("The video is HDR, Unifiler will attempt to apply colorspace conversion.")
	}
	offsetDef, intervalDef := m.defaultScreenshotParameter(limitMs)
	offsetMs := opx.Ternary(offset == 0, offsetDef, m.convertSecondToMs(offset))
	intervalMs := opx.Ternary(interval == 0, intervalDef, m.convertSecondToMs(interval))

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	if n, ok := m.notifier.(*tui.BubbleteaNotifier); ok {
		ps := tui.RunProcessStatus(n, cancel)
		defer ps.Stop()
	}

	p := diag.NewProgressTracker("Extract frames", m.notifier)
	defer p.Done()
	p.Total(limitMs.Int64())
	qualityFactor := opx.Ternary(quality == 0, 1, quality)
	outputFilenameFormat := opx.Ternary(quality == 1, filepath.Join(outputRoot, inputFile.Name+"_%s"+".jpg"), filepath.Join(outputRoot, inputFile.Name+"_%s_q%d"+".jpg"))
	count := uint64(0)
	for t := offsetMs; t.Cmp(limitMs) <= 0; t = new(big.Int).Add(t, intervalMs) {
		if err := ctx.Err(); err != nil {
			return count, err
		}
		outFile := opx.Ternary(quality == 1, fmt.Sprintf(outputFilenameFormat, m.convertSecondToTimeCode(t)), fmt.Sprintf(outputFilenameFormat, m.convertSecondToTimeCode(t), quality))
		ffmOptions := &exec.FFmpegArgsOptions{
			InputFile:        inputFile.AbsolutePath,
			InputStartTimeMs: nullable.FromInt(int(t.Int64())),

			OutputFile:       outFile,
			OutputFrameCount: nullable.FromInt(1),
			QualityFactor:    nullable.FromInt(qualityFactor),
			OverwriteOutput:  true,
		}
		if isHDR {
			ffmOptions.VideoFilter = vfHDR
		}

		_, err := exec.Run(m.cfg.Path.FFMpegPath, exec.NewFFmpegArgs(ffmOptions), nil, nil)
		if err != nil {
			m.logger.Info().Msg("Failed to take video screenshot.")
			return 0, err
		}
		p.Progress(t.Int64())
		p.Status(outFile)
		count++
	}
	return count, nil
}

// Decorator to log error occurred when calling handlers.
func (m *VideoModule) logError(err error) error {
	return LogCommandError(m.logger, err)
}

// Define Cobra Command for Video module.
func VideoCmd() *cobra.Command {
	rootCmd := &cobra.Command{
		Use:   "video",
		Short: "Work with video files.",
	}
	MarkCommandQueueable(rootCmd)
	rootCmd.PersistentFlags().StringP("file", "i", "", "Video file; same as the <file> argument.")

	infoCmd := &cobra.Command{
		Use:   "info <file>",
		Short: "Save a video's media information to <file>.mediainfo.json.",
		Long: `Run MediaInfo on the video and save its report, as JSON, to
<file>.mediainfo.json next to it.

Needs MediaInfo CLI (mediainfo) on PATH, or its location in the
paths.mediainfo setting.`,
		Example: `  unifiler video info movie.mkv`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitAppWithBanner(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			flags := ParseVideoFlags(cmd, args)
			m := NewVideoModule(c, "info")
			if err := validateSingleInput(flags.Inputs); err != nil {
				return m.logError(err)
			}
			return m.logError(m.Info(flags.Inputs[0]))
		},
	}
	rootCmd.AddCommand(infoCmd)

	extractFramesCmd := &cobra.Command{
		Use:   "extract-frames <file>",
		Short: "Save JPEG frames taken across a video.",
		Long: `Save one JPEG frame every --interval seconds, from --offset to --limit, as
<file name>_<h>_<mm>_<ss>_<ms>_q<quality>.jpg in --output; the _q<quality> part
is left out at quality 1. Left at 0, the offset and interval are picked from the
length of the video, longer videos getting fewer frames per minute. HDR video
is converted to SDR colors.

Needs FFmpeg (ffmpeg) and MediaInfo CLI (mediainfo) on PATH, or their
locations in the paths.ffmpeg and paths.mediainfo settings.`,
		Example: `  unifiler video extract-frames movie.mkv -o ./frames
  unifiler video extract-frames movie.mkv --offset 60 --interval 30 -q 5`,
		RunE: func(cmd *cobra.Command, args []string) error {
			c, err := InitAppWithBanner(cmd.Context())
			if err != nil {
				return err
			}
			defer c.Close()
			flags := ParseVideoFlags(cmd, args)
			m := NewVideoModule(c, "extract-frames")
			if err := validateSingleInput(flags.Inputs); err != nil {
				return m.logError(err)
			}
			return m.logError(m.ExtractFrames(flags.Inputs[0], flags.Interval, flags.Offset, flags.Limit, flags.Quality, flags.OutputDir))
		},
	}
	extractFramesCmd.Flags().Float64("interval", 0, "Seconds between frames; 0 picks it from the length of the video.")
	extractFramesCmd.Flags().Float64("limit", 0, "Second to stop at; 0 goes to the end of the video.")
	extractFramesCmd.Flags().Float64("offset", 0, "Second of the first frame; 0 picks it from the length of the video.")
	extractFramesCmd.Flags().StringP("output", "o", "", "Folder to save the frames in, created when missing; the video's folder when empty.")
	extractFramesCmd.Flags().IntP("quality", "q", 3, "JPEG quality, from 1 (best, largest files) to 31 (smallest files).")
	rootCmd.AddCommand(extractFramesCmd)

	return rootCmd
}

// VideoFlags contains all flags used by Video module.
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
