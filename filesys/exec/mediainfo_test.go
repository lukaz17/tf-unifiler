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
	"testing"
)

var apocalypseCode2007JSON = "{\"creatingLibrary\":{\"name\":\"MediaInfoLib\",\"version\":\"24.00\",\"url\":\"https://mediaarea.net/MediaInfo\"},\"media\":{\"@ref\":\"/var/media/The.Apocalypse.Code.2007.1080p.BluRay.DTS.x265-TFX.mp4\",\"track\":[{\"@type\":\"General\",\"UniqueID\":\"208543189653730980696786440587063431138\",\"VideoCount\":\"1\",\"AudioCount\":\"1\",\"TextCount\":\"1\",\"MenuCount\":\"1\",\"FileExtension\":\"mp4\",\"Format\":\"Matroska\",\"Format_Version\":\"4\",\"FileSize\":\"5655516759\",\"Duration\":\"6601.547\",\"OverallBitRate\":\"6853565\",\"FrameRate\":\"23.976\",\"FrameCount\":\"158276\",\"IsStreamable\":\"Yes\",\"Encoded_Date\":\"2023-02-16 16:20:15 UTC\",\"File_Created_Date\":\"2023-07-24 17:17:52.498 UTC\",\"File_Created_Date_Local\":\"2023-07-25 00:17:52.498\",\"File_Modified_Date\":\"2023-02-16 16:23:34.967 UTC\",\"File_Modified_Date_Local\":\"2023-02-16 23:23:34.967\",\"Encoded_Application\":\"mkvmerge v64.0.0 ('Willows') 64-bit\",\"Encoded_Library\":\"libebml v1.4.2 + libmatroska v1.6.4\",\"extra\":{\"FileExtension_Invalid\":\"mkv mk3d mka mks\"}},{\"@type\":\"Video\",\"StreamOrder\":\"0\",\"ID\":\"1\",\"UniqueID\":\"13569045751096588468\",\"Format\":\"HEVC\",\"Format_Profile\":\"Main\",\"Format_Level\":\"4.1\",\"Format_Tier\":\"High\",\"CodecID\":\"V_MPEGH/ISO/HEVC\",\"Duration\":\"6601.428000000\",\"BitRate\":\"5337680\",\"Width\":\"1920\",\"Height\":\"816\",\"Sampled_Width\":\"1920\",\"Sampled_Height\":\"816\",\"PixelAspectRatio\":\"1.000\",\"DisplayAspectRatio\":\"2.353\",\"FrameRate_Mode\":\"CFR\",\"FrameRate\":\"23.976\",\"FrameRate_Num\":\"24000\",\"FrameRate_Den\":\"1001\",\"FrameCount\":\"158276\",\"ColorSpace\":\"YUV\",\"ChromaSubsampling\":\"4:2:0\",\"BitDepth\":\"8\",\"Delay\":\"0.000\",\"Delay_Source\":\"Container\",\"StreamSize\":\"4404539228\",\"Encoded_Library\":\"x265 - 3.5+69-dc12b9de0:[Windows][GCC 12.2.0][64 bit] 8bit+10bit+12bit\",\"Encoded_Library_Name\":\"x265\",\"Encoded_Library_Version\":\"3.5+69-dc12b9de0:[Windows][GCC 12.2.0][64 bit] 8bit+10bit+12bit\",\"Encoded_Library_Settings\":\"cpuid=1111039 / frame-threads=1 / numa-pools=16 / wpp / no-pmode / no-pme / no-psnr / no-ssim / log-level=2 / input-csp=1 / input-res=1920x816 / interlace=0 / total-frames=0 / level-idc=41 / high-tier=1 / uhd-bd=0 / ref=4 / no-allow-non-conformance / no-repeat-headers / annexb / no-aud / no-eob / no-eos / no-hrd / info / hash=0 / no-temporal-layers / no-open-gop / min-keyint=12 / keyint=50 / gop-lookahead=0 / bframes=8 / b-adapt=2 / b-pyramid / bframe-bias=0 / rc-lookahead=25 / lookahead-slices=4 / scenecut=40 / no-hist-scenecut / radl=0 / no-splice / no-intra-refresh / ctu=64 / min-cu-size=8 / rect / amp / max-tu-size=32 / tu-inter-depth=3 / tu-intra-depth=3 / limit-tu=4 / rdoq-level=2 / dynamic-rd=0.00 / no-ssim-rd / signhide / no-tskip / nr-intra=0 / nr-inter=0 / no-constrained-intra / strong-intra-smoothing / max-merge=4 / limit-refs=1 / limit-modes / me=2 / subme=7 / merange=51 / temporal-mvp / no-frame-dup / no-hme / weightp / weightb / no-analyze-src-pics / deblock=-3:-3 / sao / no-sao-non-deblock / rd=5 / selective-sao=4 / no-early-skip / rskip / no-fast-intra / no-tskip-fast / no-cu-lossless / no-b-intra / no-splitrd-skip / rdpenalty=0 / psy-rd=2.00 / psy-rdoq=1.00 / no-rd-refine / no-lossless / cbqpoffs=0 / crqpoffs=0 / rc=crf / crf=17.5 / qcomp=0.60 / qpstep=4 / stats-write=0 / stats-read=0 / vbv-maxrate=50000 / vbv-bufsize=50000 / vbv-init=0.9 / min-vbv-fullness=50.0 / max-vbv-fullness=80.0 / crf-max=0.0 / crf-min=0.0 / ipratio=1.40 / pbratio=1.30 / aq-mode=2 / aq-strength=1.00 / cutree / zone-count=0 / no-strict-cbr / qg-size=32 / no-rc-grain / qpmax=69 / qpmin=0 / no-const-vbv / sar=1 / overscan=0 / videoformat=5 / range=0 / colorprim=2 / transfer=2 / colormatrix=2 / chromaloc=0 / display-window=0 / cll=0,0 / min-luma=0 / max-luma=255 / log2-max-poc-lsb=8 / vui-timing-info / vui-hrd-info / slices=1 / no-opt-qp-pps / opt-ref-list-length-pps / no-multi-pass-opt-rps / scenecut-bias=0.05 / no-opt-cu-delta-qp / no-aq-motion / sbrc / no-hdr10 / no-hdr10-opt / no-dhdr10-opt / no-idr-recovery-sei / analysis-reuse-level=0 / analysis-save-reuse-level=0 / analysis-load-reuse-level=0 / scale-factor=0 / refine-intra=0 / refine-inter=0 / refine-mv=1 / refine-ctu-distortion=0 / no-limit-sao / ctu-info=0 / no-lowpass-dct / refine-analysis-type=0 / copy-pic=1 / max-ausize-factor=1.0 / no-dynamic-refine / no-single-sei / no-hevc-aq / no-svt / no-field / qp-adaptation-range=1.00 / scenecut-aware-qp=0conformance-window-offsets / right=0 / bottom=0 / decoder-max-rate=0 / no-vbv-live-multi-pass / no-mcstf\",\"Language\":\"en\",\"Default\":\"Yes\",\"Forced\":\"No\",\"colour_description_present\":\"Yes\",\"colour_description_present_Source\":\"Stream\",\"colour_range\":\"Limited\",\"colour_range_Source\":\"Stream\",\"colour_primaries_Source\":\"Stream\",\"transfer_characteristics_Source\":\"Stream\",\"matrix_coefficients_Source\":\"Stream\"},{\"@type\":\"Audio\",\"StreamOrder\":\"1\",\"ID\":\"2\",\"UniqueID\":\"8433993259508358331\",\"Format\":\"DTS\",\"Format_Settings_Mode\":\"16\",\"Format_Settings_Endianness\":\"Big\",\"CodecID\":\"A_DTS\",\"Duration\":\"6601.547000000\",\"BitRate_Mode\":\"CBR\",\"BitRate\":\"1509000\",\"Channels\":\"6\",\"ChannelPositions\":\"Front: L C R, Side: L R, LFE\",\"ChannelLayout\":\"C L R Ls Rs LFE\",\"SamplesPerFrame\":\"512\",\"SamplingRate\":\"48000\",\"SamplingCount\":\"316874256\",\"FrameRate\":\"93.750\",\"FrameCount\":\"618895\",\"BitDepth\":\"24\",\"Compression_Mode\":\"Lossy\",\"Delay\":\"0.000\",\"Delay_Source\":\"Container\",\"Video_Delay\":\"0.000\",\"StreamSize\":\"1245216740\",\"Language\":\"ru\",\"Default\":\"Yes\",\"Forced\":\"No\"},{\"@type\":\"Text\",\"StreamOrder\":\"2\",\"ID\":\"3\",\"UniqueID\":\"17648032518559112747\",\"Format\":\"PGS\",\"MuxingMode\":\"zlib\",\"CodecID\":\"S_HDMV/PGS\",\"Duration\":\"6165.576000000\",\"BitRate\":\"17655\",\"FrameRate\":\"0.202\",\"FrameCount\":\"1247\",\"ElementCount\":\"1247\",\"StreamSize\":\"13607205\",\"Title\":\"English\",\"Language\":\"en\",\"Default\":\"Yes\",\"Forced\":\"No\"},{\"@type\":\"Menu\",\"extra\":{\"_00_00_00_000\":\"en:Chapter 1\",\"_00_07_46_591\":\"en:Chapter 2\",\"_00_19_34_840\":\"en:Chapter 3\",\"_00_30_17_566\":\"en:Chapter 4\",\"_00_40_08_615\":\"en:Chapter 5\",\"_00_49_01_105\":\"en:Chapter 6\",\"_01_01_02_868\":\"en:Chapter 7\",\"_01_09_20_281\":\"en:Chapter 8\",\"_01_19_58_836\":\"en:Chapter 9\",\"_01_29_47_257\":\"en:Chapter 10\",\"_01_37_58_623\":\"en:Chapter 11\",\"_01_40_21_015\":\"en:Chapter 12\"}}]}}"
var beyondTheBlackFreeMeJSON = "{\"creatingLibrary\":{\"name\":\"MediaInfoLib\",\"version\":\"24.00\",\"url\":\"https://mediaarea.net/MediaInfo\"},\"media\":{\"@ref\":\"/var/media/beyond_the_black-free_me.mkv\",\"track\":[{\"@type\":\"General\",\"UniqueID\":\"105131622467761036355380856157015977195\",\"VideoCount\":\"1\",\"AudioCount\":\"1\",\"FileExtension\":\"mkv\",\"Format\":\"Matroska\",\"Format_Version\":\"4\",\"FileSize\":\"394899090\",\"Duration\":\"246.901\",\"OverallBitRate\":\"12795382\",\"FrameRate\":\"25.000\",\"FrameCount\":\"6172\",\"StreamSize\":\"90016\",\"IsStreamable\":\"Yes\",\"Encoded_Date\":\"2023-12-29 05:02:28 UTC\",\"File_Created_Date\":\"2023-12-29 05:02:28.143 UTC\",\"File_Created_Date_Local\":\"2023-12-29 12:02:28.143\",\"File_Modified_Date\":\"2023-12-29 05:02:32.633 UTC\",\"File_Modified_Date_Local\":\"2023-12-29 12:02:32.633\",\"Encoded_Application\":\"mkvmerge v64.0.0 ('Willows') 64-bit\",\"Encoded_Library\":\"libebml v1.4.2 + libmatroska v1.6.4 / IDMmkvlib0.1\",\"extra\":{\"Language\":\"und\"}},{\"@type\":\"Video\",\"StreamOrder\":\"0\",\"ID\":\"1\",\"UniqueID\":\"1\",\"Format\":\"VP9\",\"Format_Profile\":\"0\",\"CodecID\":\"V_VP9\",\"Duration\":\"246.840000000\",\"BitRate\":\"12666286\",\"Width\":\"3840\",\"Height\":\"1600\",\"Sampled_Width\":\"3840\",\"Sampled_Height\":\"1600\",\"PixelAspectRatio\":\"1.000\",\"DisplayAspectRatio\":\"2.400\",\"FrameRate_Mode\":\"CFR\",\"FrameRate\":\"25.000\",\"FrameRate_Num\":\"25\",\"FrameRate_Den\":\"1\",\"FrameCount\":\"6172\",\"ColorSpace\":\"YUV\",\"ChromaSubsampling\":\"4:2:0\",\"ChromaSubsampling_Position\":\"Type 1\",\"BitDepth\":\"8\",\"Delay\":\"0.000\",\"Delay_Source\":\"Container\",\"StreamSize\":\"390818269\",\"Language\":\"en\",\"Default\":\"Yes\",\"Forced\":\"No\",\"colour_description_present\":\"Yes\",\"colour_description_present_Source\":\"Container\",\"colour_range\":\"Limited\",\"colour_range_Source\":\"Container / Stream\",\"colour_primaries\":\"BT.709\",\"colour_primaries_Source\":\"Container\",\"transfer_characteristics\":\"BT.709\",\"transfer_characteristics_Source\":\"Container\",\"matrix_coefficients\":\"BT.709\",\"matrix_coefficients_Source\":\"Container / Stream\"},{\"@type\":\"Audio\",\"StreamOrder\":\"1\",\"ID\":\"2\",\"UniqueID\":\"2\",\"Format\":\"Opus\",\"CodecID\":\"A_OPUS\",\"Duration\":\"246.901000000\",\"BitRate\":\"129308\",\"Channels\":\"2\",\"ChannelPositions\":\"Front: L R\",\"ChannelLayout\":\"L R\",\"SamplesPerFrame\":\"960\",\"SamplingRate\":\"48000\",\"SamplingCount\":\"11851248\",\"FrameRate\":\"50.000\",\"FrameCount\":\"12345\",\"BitDepth\":\"16\",\"Compression_Mode\":\"Lossy\",\"Delay\":\"0.000\",\"Delay_Source\":\"Container\",\"Video_Delay\":\"0.000\",\"StreamSize\":\"3990805\",\"Language\":\"en\",\"Default\":\"Yes\",\"Forced\":\"No\"}]}}"

func TestDecodeMediaInfoJson(t *testing.T) {
	tests := []struct {
		name         string
		json         string
		generalCount int
		videoCount   int
		audioCount   int
		textCount    int
	}{
		{"The.Apocalypse.Code.2007.1080p.BluRay.DTS.x265-TFX.mp4", apocalypseCode2007JSON, 1, 1, 1, 1},
		{"Beyond The Black - Free Me.mkv", beyondTheBlackFreeMeJSON, 1, 1, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			report, err := DecodeMediaInfoJson(tt.json)
			if err != nil {
				t.Error(err)
			}
			if report == nil {
				t.Errorf("Report must not be null")
			}
			if len(report.Media.GeneralTracks) != tt.generalCount {
				t.Errorf("Wrong number of general tracks. Expected '%d' Actual '%d'", tt.generalCount, len(report.Media.GeneralTracks))
			}
			if len(report.Media.VideoTracks) != tt.videoCount {
				t.Errorf("Wrong number of video tracks. Expected '%d' Actual '%d'", tt.videoCount, len(report.Media.VideoTracks))
			}
			if len(report.Media.AudioTracks) != tt.audioCount {
				t.Errorf("Wrong number of audio tracks. Expected '%d' Actual '%d'", tt.audioCount, len(report.Media.AudioTracks))
			}
			if len(report.Media.TextTracks) != tt.textCount {
				t.Errorf("Wrong number of text tracks. Expected '%d' Actual '%d'", tt.textCount, len(report.Media.TextTracks))
			}
		})
	}
}
