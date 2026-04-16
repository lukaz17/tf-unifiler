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

import "encoding/json"

type MediaInfoReport struct {
	CreatingLibrary *MediaInfoLibrary `json:"creatingLibrary,omitempty"`
	Media           *MediaInfoMedia   `json:"media,omitempty"`
}

type MediaInfoLibrary struct {
	Name    string `json:"name,omitempty"`
	URL     string `json:"url,omitempty"`
	Version string `json:"version,omitempty"`
}

type MediaInfoMedia struct {
	Ref string `json:"@ref,omitempty"`

	AudioTracks   []*MediaInfoTrackAudio   `json:"audios,omitempty"`
	GeneralTracks []*MediaInfoTrackGeneral `json:"generals,omitempty"`
	GenericTracks []*MediaInfoTrackGeneric `json:"generics,omitempty"`
	ImageTracks   []*MediaInfoTrackImage   `json:"images,omitempty"`
	MenuTracks    []*MediaInfoTrackMenu    `json:"menus,omitempty"`
	OtherTracks   []*MediaInfoTrackOther   `json:"others,omitempty"`
	TextTracks    []*MediaInfoTrackText    `json:"texts,omitempty"`
	VideoTracks   []*MediaInfoTrackVideo   `json:"videos,omitempty"`
}

type MediaInfoTrack struct {
	Audio   *MediaInfoTrackAudio
	General *MediaInfoTrackGeneral
	Generic *MediaInfoTrackGeneric
	Image   *MediaInfoTrackImage
	Menu    *MediaInfoTrackMenu
	Other   *MediaInfoTrackOther
	Text    *MediaInfoTrackText
	Video   *MediaInfoTrackVideo
}

func (t *MediaInfoTrack) UnmarshalJSON(data []byte) error {
	if string(data) == "null" || string(data) == `""` {
		return nil
	}
	var mapping map[string]interface{}
	if err := json.Unmarshal(data, &mapping); err != nil {
		return err
	}

	switch mapping["@type"].(string) {
	case "General":
		var st *MediaInfoTrackGeneral
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		t.General = st
	case "Video":
		var st *MediaInfoTrackVideo
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		t.Video = st
	case "Audio":
		var st *MediaInfoTrackAudio
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		t.Audio = st
	case "Text":
		var st *MediaInfoTrackText
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		t.Text = st
	case "Image":
		var st *MediaInfoTrackImage
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		t.Image = st
	case "Menu":
		var st *MediaInfoTrackMenu
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		t.Menu = st
	case "Other":
		var st *MediaInfoTrackOther
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		t.Other = st
	default:
		var st *MediaInfoTrackGeneric
		if err := json.Unmarshal(data, &st); err != nil {
			return err
		}
		t.Generic = st
	}

	return nil
}

/*
 * Reference: https://github.com/MediaArea/MediaInfoLib/blob/master/Source/Resource/Text/Stream/Audio.csv
 */
type MediaInfoTrackAudio struct {
	TrackType string `json:"@type,omitempty"`

	Alignment                         string `json:"Alignment,omitempty"`
	AlternateGroup                    string `json:"AlternateGroup,omitempty"`
	BitDepth                          string `json:"BitDepth,omitempty"`
	BitDepthDetected                  string `json:"BitDepth_Detected,omitempty"`
	BitDepthStored                    string `json:"BitDepth_Stored,omitempty"`
	BitRate                           string `json:"BitRate,omitempty"`
	BitRateEncoded                    string `json:"BitRate_Encoded,omitempty"`
	BitRateMaximum                    string `json:"BitRate_Maximum,omitempty"`
	BitRateMinimum                    string `json:"BitRate_Minimum,omitempty"`
	BitRateMode                       string `json:"BitRate_Mode,omitempty"`
	BitRateNominal                    string `json:"BitRate_Nominal,omitempty"`
	Channels                          string `json:"Channel(s),omitempty"`
	ChannelsOriginal                  string `json:"Channel(s)_Original,omitempty"`
	ChannelLayout                     string `json:"ChannelLayout,omitempty"`
	ChannelLayoutOriginal             string `json:"ChannelLayout_Original,omitempty"`
	ChannelLayoutID                   string `json:"ChannelLayoutID,omitempty"`
	ChannelPositions                  string `json:"ChannelPositions,omitempty"`
	ChannelPositionsOriginal          string `json:"ChannelPositions_Original,omitempty"`
	Codec                             string `json:"Codec,omitempty"`
	CodecDescription                  string `json:"Codec_Description,omitempty"`
	CodecProfile                      string `json:"Codec_Profile,omitempty"`
	CodecSettings                     string `json:"Codec_Settings,omitempty"`
	CodecSettingsAutomatic            string `json:"Codec_Settings_Automatic,omitempty"`
	CodecSettingsEndianness           string `json:"Codec_Settings_Endianness,omitempty"`
	CodecSettingsFirm                 string `json:"Codec_Settings_Firm,omitempty"`
	CodecSettingsFloor                string `json:"Codec_Settings_Floor,omitempty"`
	CodecSettingsITU                  string `json:"Codec_Settings_ITU,omitempty"`
	CodecSettingsLaw                  string `json:"Codec_Settings_Law,omitempty"`
	CodecSettingsSign                 string `json:"Codec_Settings_Sign,omitempty"`
	CodecID                           string `json:"CodecID,omitempty"`
	CodecIDDescription                string `json:"CodecID_Description,omitempty"`
	CompressionMode                   string `json:"Compression_Mode,omitempty"`
	CompressionRatio                  string `json:"Compression_Ratio,omitempty"`
	Count                             string `json:"Count,omitempty"`
	Default                           string `json:"Default,omitempty"`
	Delay                             string `json:"Delay,omitempty"`
	DelayDropFrame                    string `json:"Delay_DropFrame,omitempty"`
	DelayOriginal                     string `json:"Delay_Original,omitempty"`
	DelayOriginalDropFrame            string `json:"Delay_Original_DropFrame,omitempty"`
	DelayOriginalSettings             string `json:"Delay_Original_Settings,omitempty"`
	DelayOriginalSource               string `json:"Delay_Original_Source,omitempty"`
	DelaySettings                     string `json:"Delay_Settings,omitempty"`
	DelaySource                       string `json:"Delay_Source,omitempty"`
	Disabled                          string `json:"Disabled,omitempty"`
	Duration                          string `json:"Duration,omitempty"`
	DurationFirstFrame                string `json:"Duration_FirstFrame,omitempty"`
	DurationLastFrame                 string `json:"Duration_LastFrame,omitempty"`
	EncodedApplication                string `json:"Encoded_Application,omitempty"`
	EncodedApplicationCompanyName     string `json:"Encoded_Application_CompanyName,omitempty"`
	EncodedApplicationName            string `json:"Encoded_Application_Name,omitempty"`
	EncodedApplicationUrl             string `json:"Encoded_Application_Url,omitempty"`
	EncodedApplicationVersion         string `json:"Encoded_Application_Version,omitempty"`
	EncodedDate                       string `json:"Encoded_Date,omitempty"`
	EncodedLibrary                    string `json:"Encoded_Library,omitempty"`
	EncodedLibraryCompanyName         string `json:"Encoded_Library_CompanyName,omitempty"`
	EncodedLibraryDate                string `json:"Encoded_Library_Date,omitempty"`
	EncodedLibraryName                string `json:"Encoded_Library_Name,omitempty"`
	EncodedLibrarySettings            string `json:"Encoded_Library_Settings,omitempty"`
	EncodedLibraryVersion             string `json:"Encoded_Library_Version,omitempty"`
	EncodedOperatingSystem            string `json:"Encoded_OperatingSystem,omitempty"`
	Encryption                        string `json:"Encryption,omitempty"`
	FirstPacketOrder                  string `json:"FirstPacketOrder,omitempty"`
	Forced                            string `json:"Forced,omitempty"`
	Format                            string `json:"Format,omitempty"`
	FormatAdditionalFeatures          string `json:"Format_AdditionalFeatures,omitempty"`
	FormatCommercial                  string `json:"Format_Commercial,omitempty"`
	FormatCommercialIfAny             string `json:"Format_Commercial_IfAny,omitempty"`
	FormatCompression                 string `json:"Format_Compression,omitempty"`
	FormatLevel                       string `json:"Format_Level,omitempty"`
	FormatProfile                     string `json:"Format_Profile,omitempty"`
	FormatSettings                    string `json:"Format_Settings,omitempty"`
	FormatSettingsEmphasis            string `json:"Format_Settings_Emphasis,omitempty"`
	FormatSettingsEndianness          string `json:"Format_Settings_Endianness,omitempty"`
	FormatSettingsFirm                string `json:"Format_Settings_Firm,omitempty"`
	FormatSettingsFloor               string `json:"Format_Settings_Floor,omitempty"`
	FormatSettingsITU                 string `json:"Format_Settings_ITU,omitempty"`
	FormatSettingsLaw                 string `json:"Format_Settings_Law,omitempty"`
	FormatSettingsMode                string `json:"Format_Settings_Mode,omitempty"`
	FormatSettingsModeExtension       string `json:"Format_Settings_ModeExtension,omitempty"`
	FormatSettingsPS                  string `json:"Format_Settings_PS,omitempty"`
	FormatSettingsSBR                 string `json:"Format_Settings_SBR,omitempty"`
	FormatSettingsSign                string `json:"Format_Settings_Sign,omitempty"`
	FormatSettingsWrapping            string `json:"Format_Settings_Wrapping,omitempty"`
	FormatVersion                     string `json:"Format_Version,omitempty"`
	FrameCount                        string `json:"FrameCount,omitempty"`
	FrameRate                         string `json:"FrameRate,omitempty"`
	FrameRateDen                      string `json:"FrameRate_Den,omitempty"`
	FrameRateNum                      string `json:"FrameRate_Num,omitempty"`
	ID                                string `json:"ID,omitempty"`
	Inform                            string `json:"Inform,omitempty"`
	InterleaveDuration                string `json:"Interleave_Duration,omitempty"`
	InterleavePreload                 string `json:"Interleave_Preload,omitempty"`
	InterleaveVideoFrames             string `json:"Interleave_VideoFrames,omitempty"`
	InternetMediaType                 string `json:"InternetMediaType,omitempty"`
	Language                          string `json:"Language,omitempty"`
	LanguageMore                      string `json:"Language_More,omitempty"`
	MatrixChannels                    string `json:"Matrix_Channel(s),omitempty"`
	MatrixChannelPositions            string `json:"Matrix_ChannelPositions,omitempty"`
	MatrixFormat                      string `json:"Matrix_Format,omitempty"`
	MenuID                            string `json:"MenuID,omitempty"`
	MuxingMode                        string `json:"MuxingMode,omitempty"`
	MuxingModeMoreInfo                string `json:"MuxingMode_MoreInfo,omitempty"`
	OriginalSourceMediumID            string `json:"OriginalSourceMedium_ID,omitempty"`
	ReplayGainGain                    string `json:"ReplayGain_Gain,omitempty"`
	ReplayGainPeak                    string `json:"ReplayGain_Peak,omitempty"`
	Resolution                        string `json:"Resolution,omitempty"`
	SamplesPerFrame                   string `json:"SamplesPerFrame,omitempty"`
	SamplingCount                     string `json:"SamplingCount,omitempty"`
	SamplingRate                      string `json:"SamplingRate,omitempty"`
	ServiceKind                       string `json:"ServiceKind,omitempty"`
	SourceDuration                    string `json:"Source_Duration,omitempty"`
	SourceDurationFirstFrame          string `json:"Source_Duration_FirstFrame,omitempty"`
	SourceDurationLastFrame           string `json:"Source_Duration_LastFrame,omitempty"`
	SourceFrameCount                  string `json:"Source_FrameCount,omitempty"`
	SourceSamplingCount               string `json:"Source_SamplingCount,omitempty"`
	SourceStreamSize                  string `json:"Source_StreamSize,omitempty"`
	SourceStreamSizeEncoded           string `json:"Source_StreamSize_Encoded,omitempty"`
	SourceStreamSizeEncodedProportion string `json:"Source_StreamSize_Encoded_Proportion,omitempty"`
	SourceStreamSizeProportion        string `json:"Source_StreamSize_Proportion,omitempty"`
	Status                            string `json:"Status,omitempty"`
	StreamCount                       string `json:"StreamCount,omitempty"`
	StreamKind                        string `json:"StreamKind,omitempty"`
	StreamKindID                      string `json:"StreamKindID,omitempty"`
	StreamKindPos                     string `json:"StreamKindPos,omitempty"`
	StreamOrder                       string `json:"StreamOrder,omitempty"`
	StreamSize                        string `json:"StreamSize,omitempty"`
	StreamSizeDemuxed                 string `json:"StreamSize_Demuxed,omitempty"`
	StreamSizeEncoded                 string `json:"StreamSize_Encoded,omitempty"`
	StreamSizeEncodedProportion       string `json:"StreamSize_Encoded_Proportion,omitempty"`
	StreamSizeProportion              string `json:"StreamSize_Proportion,omitempty"`
	TaggedDate                        string `json:"Tagged_Date,omitempty"`
	TimeCodeDropFrame                 string `json:"TimeCode_DropFrame,omitempty"`
	TimeCodeFirstFrame                string `json:"TimeCode_FirstFrame,omitempty"`
	TimeCodeLastFrame                 string `json:"TimeCode_LastFrame,omitempty"`
	TimeCodeSettings                  string `json:"TimeCode_Settings,omitempty"`
	TimeCodeSource                    string `json:"TimeCode_Source,omitempty"`
	Title                             string `json:"Title,omitempty"`
	UniqueID                          string `json:"UniqueID,omitempty"`
	VideoDelay                        string `json:"Video_Delay,omitempty"`
	Video0Delay                       string `json:"Video0_Delay,omitempty"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

/*
 * Reference: https://github.com/MediaArea/MediaInfoLib/blob/master/Source/Resource/Text/Stream/General.csv
 */
type MediaInfoTrackGeneral struct {
	TrackType string `json:"@type,omitempty"`

	Accompaniment                  string `json:"Accompaniment,omitempty"`
	Actor                          string `json:"Actor,omitempty"`
	ActorCharacter                 string `json:"Actor_Character,omitempty"`
	AddedDate                      string `json:"Added_Date,omitempty"`
	Album                          string `json:"Album,omitempty"`
	AlbumMore                      string `json:"Album_More,omitempty"`
	AlbumReplayGainGain            string `json:"Album_ReplayGain_Gain,omitempty"`
	AlbumReplayGainPeak            string `json:"Album_ReplayGain_Peak,omitempty"`
	ArchivalLocation               string `json:"Archival_Location,omitempty"`
	Arranger                       string `json:"Arranger,omitempty"`
	ArtDirector                    string `json:"ArtDirector,omitempty"`
	AssistantDirector              string `json:"AssistantDirector,omitempty"`
	AudioChannelsTotal             string `json:"Audio_Channels_Total,omitempty"`
	AudioCodecList                 string `json:"Audio_Codec_List,omitempty"`
	AudioFormatList                string `json:"Audio_Format_List,omitempty"`
	AudioFormatWithHintList        string `json:"Audio_Format_WithHint_List,omitempty"`
	AudioLanguageList              string `json:"Audio_Language_List,omitempty"`
	AudioCount                     string `json:"AudioCount,omitempty"`
	BarCode                        string `json:"BarCode,omitempty"`
	BPM                            string `json:"BPM,omitempty"`
	CatalogNumber                  string `json:"CatalogNumber,omitempty"`
	Chapter                        string `json:"Chapter,omitempty"`
	Choreographer                  string `json:"Choreographer,omitempty"`
	Codec                          string `json:"Codec,omitempty"`
	CodecSettings                  string `json:"Codec_Settings,omitempty"`
	CodecSettingsAutomatic         string `json:"Codec_Settings_Automatic,omitempty"`
	CodecID                        string `json:"CodecID,omitempty"`
	CodecIDCompatible              string `json:"CodecID_Compatible,omitempty"`
	CodecIDDescription             string `json:"CodecID_Description,omitempty"`
	CodecIDVersion                 string `json:"CodecID_Version,omitempty"`
	CoDirector                     string `json:"CoDirector,omitempty"`
	Collection                     string `json:"Collection,omitempty"`
	Comic                          string `json:"Comic,omitempty"`
	ComicMore                      string `json:"Comic_More,omitempty"`
	Comment                        string `json:"Comment,omitempty"`
	CommissionedBy                 string `json:"CommissionedBy,omitempty"`
	Compilation                    string `json:"Compilation,omitempty"`
	CompleteName                   string `json:"CompleteName,omitempty"`
	CompleteNameLast               string `json:"CompleteName_Last,omitempty"`
	Composer                       string `json:"Composer,omitempty"`
	Conductor                      string `json:"Conductor,omitempty"`
	ContentType                    string `json:"ContentType,omitempty"`
	CoProducer                     string `json:"CoProducer,omitempty"`
	Copyright                      string `json:"Copyright,omitempty"`
	CostumeDesigner                string `json:"CostumeDesigner,omitempty"`
	Count                          string `json:"Count,omitempty"`
	Country                        string `json:"Country,omitempty"`
	Cover                          string `json:"Cover,omitempty"`
	CoverData                      string `json:"Cover_Data,omitempty"`
	CoverDescription               string `json:"Cover_Description,omitempty"`
	CoverMime                      string `json:"Cover_Mime,omitempty"`
	CoverType                      string `json:"Cover_Type,omitempty"`
	Cropped                        string `json:"Cropped,omitempty"`
	DataSize                       string `json:"DataSize,omitempty"`
	Delay                          string `json:"Delay,omitempty"`
	DelayDropFrame                 string `json:"Delay_DropFrame,omitempty"`
	DelaySettings                  string `json:"Delay_Settings,omitempty"`
	DelaySource                    string `json:"Delay_Source,omitempty"`
	Description                    string `json:"Description,omitempty"`
	Dimensions                     string `json:"Dimensions,omitempty"`
	Director                       string `json:"Director,omitempty"`
	DirectorOfPhotography          string `json:"DirectorOfPhotography,omitempty"`
	DistributedBy                  string `json:"DistributedBy,omitempty"`
	Domain                         string `json:"Domain,omitempty"`
	DotsPerInch                    string `json:"DotsPerInch,omitempty"`
	Duration                       string `json:"Duration,omitempty"`
	DurationEnd                    string `json:"Duration_End,omitempty"`
	DurationStart                  string `json:"Duration_Start,omitempty"`
	EditedBy                       string `json:"EditedBy,omitempty"`
	EncodedApplication             string `json:"Encoded_Application,omitempty"`
	EncodedApplicationCompanyName  string `json:"Encoded_Application_CompanyName,omitempty"`
	EncodedApplicationName         string `json:"Encoded_Application_Name,omitempty"`
	EncodedApplicationUrl          string `json:"Encoded_Application_Url,omitempty"`
	EncodedApplicationVersion      string `json:"Encoded_Application_Version,omitempty"`
	EncodedDate                    string `json:"Encoded_Date,omitempty"`
	EncodedLibrary                 string `json:"Encoded_Library,omitempty"`
	EncodedLibraryCompanyName      string `json:"Encoded_Library_CompanyName,omitempty"`
	EncodedLibraryDate             string `json:"Encoded_Library_Date,omitempty"`
	EncodedLibraryName             string `json:"Encoded_Library_Name,omitempty"`
	EncodedLibrarySettings         string `json:"Encoded_Library_Settings,omitempty"`
	EncodedLibraryVersion          string `json:"Encoded_Library_Version,omitempty"`
	EncodedOperatingSystem         string `json:"Encoded_OperatingSystem,omitempty"`
	EncodedBy                      string `json:"EncodedBy,omitempty"`
	Encryption                     string `json:"Encryption,omitempty"`
	EncryptionFormat               string `json:"Encryption_Format,omitempty"`
	EncryptionInitializationVector string `json:"Encryption_InitializationVector,omitempty"`
	EncryptionLength               string `json:"Encryption_Length,omitempty"`
	EncryptionMethod               string `json:"Encryption_Method,omitempty"`
	EncryptionMode                 string `json:"Encryption_Mode,omitempty"`
	EncryptionPadding              string `json:"Encryption_Padding,omitempty"`
	EPGPositionsBegin              string `json:"EPG_Positions_Begin,omitempty"`
	EPGPositionsEnd                string `json:"EPG_Positions_End,omitempty"`
	ExecutiveProducer              string `json:"ExecutiveProducer,omitempty"`
	FileCreatedDate                string `json:"File_Created_Date,omitempty"`
	FileCreatedDateLocal           string `json:"File_Created_Date_Local,omitempty"`
	FileModifiedDate               string `json:"File_Modified_Date,omitempty"`
	FileModifiedDateLocal          string `json:"File_Modified_Date_Local,omitempty"`
	FileExtension                  string `json:"FileExtension,omitempty"`
	FileExtensionLast              string `json:"FileExtension_Last,omitempty"`
	FileName                       string `json:"FileName,omitempty"`
	FileNameLast                   string `json:"FileName_Last,omitempty"`
	FileNameExtension              string `json:"FileNameExtension,omitempty"`
	FileNameExtensionLast          string `json:"FileNameExtension_Last,omitempty"`
	FileSize                       string `json:"FileSize,omitempty"`
	FirstPacketOrder               string `json:"FirstPacketOrder,omitempty"`
	FolderName                     string `json:"FolderName,omitempty"`
	FolderNameLast                 string `json:"FolderName_Last,omitempty"`
	FooterSize                     string `json:"FooterSize,omitempty"`
	Format                         string `json:"Format,omitempty"`
	FormatAdditionalFeatures       string `json:"Format_AdditionalFeatures,omitempty"`
	FormatCommercial               string `json:"Format_Commercial,omitempty"`
	FormatCommercialIfAny          string `json:"Format_Commercial_IfAny,omitempty"`
	FormatCompression              string `json:"Format_Compression,omitempty"`
	FormatLevel                    string `json:"Format_Level,omitempty"`
	FormatProfile                  string `json:"Format_Profile,omitempty"`
	FormatSettings                 string `json:"Format_Settings,omitempty"`
	FormatVersion                  string `json:"Format_Version,omitempty"`
	FrameCount                     string `json:"FrameCount,omitempty"`
	FrameRate                      string `json:"FrameRate,omitempty"`
	FrameRateDen                   string `json:"FrameRate_Den,omitempty"`
	FrameRateNum                   string `json:"FrameRate_Num,omitempty"`
	GeneralCount                   string `json:"GeneralCount,omitempty"`
	Genre                          string `json:"Genre,omitempty"`
	Grouping                       string `json:"Grouping,omitempty"`
	HeaderSize                     string `json:"HeaderSize,omitempty"`
	ICRA                           string `json:"ICRA,omitempty"`
	ID                             string `json:"ID,omitempty"`
	ImageCodecList                 string `json:"Image_Codec_List,omitempty"`
	ImageFormatList                string `json:"Image_Format_List,omitempty"`
	ImageFormatWithHintList        string `json:"Image_Format_WithHint_List,omitempty"`
	ImageLanguageList              string `json:"Image_Language_List,omitempty"`
	ImageCount                     string `json:"ImageCount,omitempty"`
	Inform                         string `json:"Inform,omitempty"`
	Interleaved                    string `json:"Interleaved,omitempty"`
	InternetMediaType              string `json:"InternetMediaType,omitempty"`
	ISAN                           string `json:"ISAN,omitempty"`
	ISBN                           string `json:"ISBN,omitempty"`
	ISRC                           string `json:"ISRC,omitempty"`
	IsStreamable                   string `json:"IsStreamable,omitempty"`
	Keywords                       string `json:"Keywords,omitempty"`
	Label                          string `json:"Label,omitempty"`
	LabelCode                      string `json:"LabelCode,omitempty"`
	LawRating                      string `json:"LawRating,omitempty"`
	LawRatingReason                string `json:"LawRating_Reason,omitempty"`
	LCCN                           string `json:"LCCN,omitempty"`
	Lightness                      string `json:"Lightness,omitempty"`
	Lyricist                       string `json:"Lyricist,omitempty"`
	Lyrics                         string `json:"Lyrics,omitempty"`
	MasteredDate                   string `json:"Mastered_Date,omitempty"`
	MasteredBy                     string `json:"MasteredBy,omitempty"`
	MenuCodecList                  string `json:"Menu_Codec_List,omitempty"`
	MenuFormatList                 string `json:"Menu_Format_List,omitempty"`
	MenuFormatWithHintList         string `json:"Menu_Format_WithHint_List,omitempty"`
	MenuLanguageList               string `json:"Menu_Language_List,omitempty"`
	MenuCount                      string `json:"MenuCount,omitempty"`
	MenuID                         string `json:"MenuID,omitempty"`
	Mood                           string `json:"Mood,omitempty"`
	Movie                          string `json:"Movie,omitempty"`
	MovieMore                      string `json:"Movie_More,omitempty"`
	MusicBy                        string `json:"MusicBy,omitempty"`
	NetworkName                    string `json:"NetworkName,omitempty"`
	OriginalNetworkName            string `json:"OriginalNetworkName,omitempty"`
	OriginalSourceForm             string `json:"OriginalSourceForm,omitempty"`
	OriginalSourceMedium           string `json:"OriginalSourceMedium,omitempty"`
	OriginalSourceMediumID         string `json:"OriginalSourceMedium_ID,omitempty"`
	OtherCodecList                 string `json:"Other_Codec_List,omitempty"`
	OtherFormatList                string `json:"Other_Format_List,omitempty"`
	OtherFormatWithHintList        string `json:"Other_Format_WithHint_List,omitempty"`
	OtherLanguageList              string `json:"Other_Language_List,omitempty"`
	OtherCount                     string `json:"OtherCount,omitempty"`
	OverallBitRate                 string `json:"OverallBitRate,omitempty"`
	OverallBitRateMaximum          string `json:"OverallBitRate_Maximum,omitempty"`
	OverallBitRateMinimum          string `json:"OverallBitRate_Minimum,omitempty"`
	OverallBitRateMode             string `json:"OverallBitRate_Mode,omitempty"`
	OverallBitRateNominal          string `json:"OverallBitRate_Nominal,omitempty"`
	Owner                          string `json:"Owner,omitempty"`
	PackageName                    string `json:"PackageName,omitempty"`
	Part                           string `json:"Part,omitempty"`
	Performer                      string `json:"Performer,omitempty"`
	Period                         string `json:"Period,omitempty"`
	PlayedCount                    string `json:"Played_Count,omitempty"`
	PlayedFirstDate                string `json:"Played_First_Date,omitempty"`
	PlayedLastDate                 string `json:"Played_Last_Date,omitempty"`
	PodcastCategory                string `json:"PodcastCategory,omitempty"`
	Producer                       string `json:"Producer,omitempty"`
	ProducerCopyright              string `json:"Producer_Copyright,omitempty"`
	ProductionDesigner             string `json:"ProductionDesigner,omitempty"`
	ProductionStudio               string `json:"ProductionStudio,omitempty"`
	Publisher                      string `json:"Publisher,omitempty"`
	Rating                         string `json:"Rating,omitempty"`
	RecordedDate                   string `json:"Recorded_Date,omitempty"`
	RecordedLocation               string `json:"Recorded_Location,omitempty"`
	Reel                           string `json:"Reel,omitempty"`
	ReleasedDate                   string `json:"Released_Date,omitempty"`
	RemixedBy                      string `json:"RemixedBy,omitempty"`
	ScreenplayBy                   string `json:"ScreenplayBy,omitempty"`
	Season                         string `json:"Season,omitempty"`
	SeasonPosition                 string `json:"Season_Position,omitempty"`
	SeasonPositionTotal            string `json:"Season_Position_Total,omitempty"`
	ServiceChannel                 string `json:"ServiceChannel,omitempty"`
	ServiceName                    string `json:"ServiceName,omitempty"`
	ServiceProvider                string `json:"ServiceProvider,omitempty"`
	ServiceType                    string `json:"ServiceType,omitempty"`
	SoundEngineer                  string `json:"SoundEngineer,omitempty"`
	Status                         string `json:"Status,omitempty"`
	StreamCount                    string `json:"StreamCount,omitempty"`
	StreamKind                     string `json:"StreamKind,omitempty"`
	StreamKindID                   string `json:"StreamKindID,omitempty"`
	StreamKindPos                  string `json:"StreamKindPos,omitempty"`
	StreamOrder                    string `json:"StreamOrder,omitempty"`
	StreamSize                     string `json:"StreamSize,omitempty"`
	StreamSizeDemuxed              string `json:"StreamSize_Demuxed,omitempty"`
	StreamSizeProportion           string `json:"StreamSize_Proportion,omitempty"`
	Subject                        string `json:"Subject,omitempty"`
	SubTrack                       string `json:"SubTrack,omitempty"`
	Summary                        string `json:"Summary,omitempty"`
	Synopsis                       string `json:"Synopsis,omitempty"`
	TaggedApplication              string `json:"Tagged_Application,omitempty"`
	TaggedDate                     string `json:"Tagged_Date,omitempty"`
	TermsOfUse                     string `json:"TermsOfUse,omitempty"`
	TextCodecList                  string `json:"Text_Codec_List,omitempty"`
	TextFormatList                 string `json:"Text_Format_List,omitempty"`
	TextFormatWithHintList         string `json:"Text_Format_WithHint_List,omitempty"`
	TextLanguageList               string `json:"Text_Language_List,omitempty"`
	TextCount                      string `json:"TextCount,omitempty"`
	ThanksTo                       string `json:"ThanksTo,omitempty"`
	TimeZone                       string `json:"TimeZone,omitempty"`
	Title                          string `json:"Title,omitempty"`
	TitleMore                      string `json:"Title_More,omitempty"`
	Track                          string `json:"Track,omitempty"`
	TrackMore                      string `json:"Track_More,omitempty"`
	UMID                           string `json:"UMID,omitempty"`
	UniqueID                       string `json:"UniqueID,omitempty"`
	UniversalAdIDRegistry          string `json:"UniversalAdID_Registry,omitempty"`
	UniversalAdIDValue             string `json:"UniversalAdID_Value,omitempty"`
	VideoCodecList                 string `json:"Video_Codec_List,omitempty"`
	VideoFormatList                string `json:"Video_Format_List,omitempty"`
	VideoFormatWithHintList        string `json:"Video_Format_WithHint_List,omitempty"`
	VideoLanguageList              string `json:"Video_Language_List,omitempty"`
	VideoCount                     string `json:"VideoCount,omitempty"`
	WrittenDate                    string `json:"Written_Date,omitempty"`
	WrittenLocation                string `json:"Written_Location,omitempty"`
	WrittenBy                      string `json:"WrittenBy,omitempty"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

/*
 * Reference: https://github.com/MediaArea/MediaInfoLib/blob/master/Source/Resource/Text/Stream/Generic.csv
 */
type MediaInfoTrackGeneric struct {
	TrackType string `json:"@type,omitempty"`

	BitDepth                          string `json:"BitDepth,omitempty"`
	BitRate                           string `json:"BitRate,omitempty"`
	BitRateEncoded                    string `json:"BitRate_Encoded,omitempty"`
	BitRateMaximum                    string `json:"BitRate_Maximum,omitempty"`
	BitRateMinimum                    string `json:"BitRate_Minimum,omitempty"`
	BitRateMode                       string `json:"BitRate_Mode,omitempty"`
	BitRateNominal                    string `json:"BitRate_Nominal,omitempty"`
	ChromaSubsampling                 string `json:"ChromaSubsampling,omitempty"`
	Codec                             string `json:"Codec,omitempty"`
	CodecID                           string `json:"CodecID,omitempty"`
	CodecIDDescription                string `json:"CodecID_Description,omitempty"`
	ColorSpace                        string `json:"ColorSpace,omitempty"`
	CompressionMode                   string `json:"Compression_Mode,omitempty"`
	CompressionRatio                  string `json:"Compression_Ratio,omitempty"`
	Delay                             string `json:"Delay,omitempty"`
	DelayDropFrame                    string `json:"Delay_DropFrame,omitempty"`
	DelayOriginal                     string `json:"Delay_Original,omitempty"`
	DelayOriginalDropFrame            string `json:"Delay_Original_DropFrame,omitempty"`
	DelayOriginalSettings             string `json:"Delay_Original_Settings,omitempty"`
	DelayOriginalSource               string `json:"Delay_Original_Source,omitempty"`
	DelaySettings                     string `json:"Delay_Settings,omitempty"`
	DelaySource                       string `json:"Delay_Source,omitempty"`
	Duration                          string `json:"Duration,omitempty"`
	Format                            string `json:"Format,omitempty"`
	FormatAdditionalFeatures          string `json:"Format_AdditionalFeatures,omitempty"`
	FormatCommercial                  string `json:"Format_Commercial,omitempty"`
	FormatCommercialIfAny             string `json:"Format_Commercial_IfAny,omitempty"`
	FormatCompression                 string `json:"Format_Compression,omitempty"`
	FormatLevel                       string `json:"Format_Level,omitempty"`
	FormatProfile                     string `json:"Format_Profile,omitempty"`
	FormatSettings                    string `json:"Format_Settings,omitempty"`
	FormatTier                        string `json:"Format_Tier,omitempty"`
	FormatVersion                     string `json:"Format_Version,omitempty"`
	FrameCount                        string `json:"FrameCount,omitempty"`
	FrameRate                         string `json:"FrameRate,omitempty"`
	FrameRateDen                      string `json:"FrameRate_Den,omitempty"`
	FrameRateNum                      string `json:"FrameRate_Num,omitempty"`
	InternetMediaType                 string `json:"InternetMediaType,omitempty"`
	Language                          string `json:"Language,omitempty"`
	Resolution                        string `json:"Resolution,omitempty"`
	ServiceName                       string `json:"ServiceName,omitempty"`
	ServiceProvider                   string `json:"ServiceProvider,omitempty"`
	SourceDuration                    string `json:"Source_Duration,omitempty"`
	SourceFrameCount                  string `json:"Source_FrameCount,omitempty"`
	SourceStreamSize                  string `json:"Source_StreamSize,omitempty"`
	SourceStreamSizeEncoded           string `json:"Source_StreamSize_Encoded,omitempty"`
	SourceStreamSizeEncodedProportion string `json:"Source_StreamSize_Encoded_Proportion,omitempty"`
	SourceStreamSizeProportion        string `json:"Source_StreamSize_Proportion,omitempty"`
	StreamSize                        string `json:"StreamSize,omitempty"`
	StreamSizeEncoded                 string `json:"StreamSize_Encoded,omitempty"`
	StreamSizeEncodedProportion       string `json:"StreamSize_Encoded_Proportion,omitempty"`
	StreamSizeProportion              string `json:"StreamSize_Proportion,omitempty"`
	TimeCodeDropFrame                 string `json:"TimeCode_DropFrame,omitempty"`
	TimeCodeFirstFrame                string `json:"TimeCode_FirstFrame,omitempty"`
	TimeCodeLastFrame                 string `json:"TimeCode_LastFrame,omitempty"`
	TimeCodeSettings                  string `json:"TimeCode_Settings,omitempty"`
	TimeCodeSource                    string `json:"TimeCode_Source,omitempty"`
	VideoDelay                        string `json:"Video_Delay,omitempty"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

/*
 * Reference: https://github.com/MediaArea/MediaInfoLib/blob/master/Source/Resource/Text/Stream/Image.csv
 */
type MediaInfoTrackImage struct {
	TrackType string `json:"@type,omitempty"`

	ActiveDisplayAspectRatio                     string `json:"Active_DisplayAspectRatio,omitempty"`
	ActiveHeight                                 string `json:"Active_Height,omitempty"`
	ActiveWidth                                  string `json:"Active_Width,omitempty"`
	AlternateGroup                               string `json:"AlternateGroup,omitempty"`
	BitDepth                                     string `json:"BitDepth,omitempty"`
	ChromaSubsampling                            string `json:"ChromaSubsampling,omitempty"`
	Codec                                        string `json:"Codec,omitempty"`
	CodecID                                      string `json:"CodecID,omitempty"`
	CodecIDDescription                           string `json:"CodecID_Description,omitempty"`
	ColorSpace                                   string `json:"ColorSpace,omitempty"`
	ColourDescriptionPresent                     string `json:"colour_description_present,omitempty"`
	ColourDescriptionPresentOriginal             string `json:"colour_description_present_Original,omitempty"`
	ColourDescriptionPresentOriginalSource       string `json:"colour_description_present_Original_Source,omitempty"`
	ColourDescriptionPresentSource               string `json:"colour_description_present_Source,omitempty"`
	ColourPrimaries                              string `json:"colour_primaries,omitempty"`
	ColourPrimariesOriginal                      string `json:"colour_primaries_Original,omitempty"`
	ColourPrimariesOriginalSource                string `json:"colour_primaries_Original_Source,omitempty"`
	ColourPrimariesSource                        string `json:"colour_primaries_Source,omitempty"`
	ColourRange                                  string `json:"colour_range,omitempty"`
	ColourRangeOriginal                          string `json:"colour_range_Original,omitempty"`
	ColourRangeOriginalSource                    string `json:"colour_range_Original_Source,omitempty"`
	ColourRangeSource                            string `json:"colour_range_Source,omitempty"`
	CompressionMode                              string `json:"Compression_Mode,omitempty"`
	CompressionRatio                             string `json:"Compression_Ratio,omitempty"`
	Count                                        string `json:"Count,omitempty"`
	Default                                      string `json:"Default,omitempty"`
	Disabled                                     string `json:"Disabled,omitempty"`
	DisplayAspectRatio                           string `json:"DisplayAspectRatio,omitempty"`
	DisplayAspectRatioOriginal                   string `json:"DisplayAspectRatio_Original,omitempty"`
	EncodedDate                                  string `json:"Encoded_Date,omitempty"`
	EncodedLibrary                               string `json:"Encoded_Library,omitempty"`
	EncodedLibraryDate                           string `json:"Encoded_Library_Date,omitempty"`
	EncodedLibraryName                           string `json:"Encoded_Library_Name,omitempty"`
	EncodedLibrarySettings                       string `json:"Encoded_Library_Settings,omitempty"`
	EncodedLibraryVersion                        string `json:"Encoded_Library_Version,omitempty"`
	Encryption                                   string `json:"Encryption,omitempty"`
	FirstPacketOrder                             string `json:"FirstPacketOrder,omitempty"`
	Forced                                       string `json:"Forced,omitempty"`
	Format                                       string `json:"Format,omitempty"`
	FormatAdditionalFeatures                     string `json:"Format_AdditionalFeatures,omitempty"`
	FormatCommercial                             string `json:"Format_Commercial,omitempty"`
	FormatCommercialIfAny                        string `json:"Format_Commercial_IfAny,omitempty"`
	FormatCompression                            string `json:"Format_Compression,omitempty"`
	FormatProfile                                string `json:"Format_Profile,omitempty"`
	FormatSettings                               string `json:"Format_Settings,omitempty"`
	FormatSettingsEndianness                     string `json:"Format_Settings_Endianness,omitempty"`
	FormatSettingsPacking                        string `json:"Format_Settings_Packing,omitempty"`
	FormatSettingsWrapping                       string `json:"Format_Settings_Wrapping,omitempty"`
	FormatVersion                                string `json:"Format_Version,omitempty"`
	HDRFormat                                    string `json:"HDR_Format,omitempty"`
	HDRFormatCommercial                          string `json:"HDR_Format_Commercial,omitempty"`
	HDRFormatCompatibility                       string `json:"HDR_Format_Compatibility,omitempty"`
	HDRFormatLevel                               string `json:"HDR_Format_Level,omitempty"`
	HDRFormatProfile                             string `json:"HDR_Format_Profile,omitempty"`
	HDRFormatSettings                            string `json:"HDR_Format_Settings,omitempty"`
	HDRFormatVersion                             string `json:"HDR_Format_Version,omitempty"`
	Height                                       string `json:"Height,omitempty"`
	HeightOffset                                 string `json:"Height_Offset,omitempty"`
	HeightOriginal                               string `json:"Height_Original,omitempty"`
	ID                                           string `json:"ID,omitempty"`
	Inform                                       string `json:"Inform,omitempty"`
	InternetMediaType                            string `json:"InternetMediaType,omitempty"`
	Language                                     string `json:"Language,omitempty"`
	LanguageMore                                 string `json:"Language_More,omitempty"`
	MasteringDisplayColorPrimaries               string `json:"MasteringDisplay_ColorPrimaries,omitempty"`
	MasteringDisplayColorPrimariesOriginal       string `json:"MasteringDisplay_ColorPrimaries_Original,omitempty"`
	MasteringDisplayColorPrimariesOriginalSource string `json:"MasteringDisplay_ColorPrimaries_Original_Source,omitempty"`
	MasteringDisplayColorPrimariesSource         string `json:"MasteringDisplay_ColorPrimaries_Source,omitempty"`
	MasteringDisplayLuminance                    string `json:"MasteringDisplay_Luminance,omitempty"`
	MasteringDisplayLuminanceOriginal            string `json:"MasteringDisplay_Luminance_Original,omitempty"`
	MasteringDisplayLuminanceOriginalSource      string `json:"MasteringDisplay_Luminance_Original_Source,omitempty"`
	MasteringDisplayLuminanceSource              string `json:"MasteringDisplay_Luminance_Source,omitempty"`
	MatrixCoefficients                           string `json:"matrix_coefficients,omitempty"`
	MatrixCoefficientsOriginal                   string `json:"matrix_coefficients_Original,omitempty"`
	MatrixCoefficientsOriginalSource             string `json:"matrix_coefficients_Original_Source,omitempty"`
	MatrixCoefficientsSource                     string `json:"matrix_coefficients_Source,omitempty"`
	MaxCLL                                       string `json:"MaxCLL,omitempty"`
	MaxCLLOriginal                               string `json:"MaxCLL_Original,omitempty"`
	MaxCLLOriginalSource                         string `json:"MaxCLL_Original_Source,omitempty"`
	MaxCLLSource                                 string `json:"MaxCLL_Source,omitempty"`
	MaxFALL                                      string `json:"MaxFALL,omitempty"`
	MaxFALLOriginal                              string `json:"MaxFALL_Original,omitempty"`
	MaxFALLOriginalSource                        string `json:"MaxFALL_Original_Source,omitempty"`
	MaxFALLSource                                string `json:"MaxFALL_Source,omitempty"`
	MenuID                                       string `json:"MenuID,omitempty"`
	OriginalSourceMediumID                       string `json:"OriginalSourceMedium_ID,omitempty"`
	PixelAspectRatio                             string `json:"PixelAspectRatio,omitempty"`
	PixelAspectRatioOriginal                     string `json:"PixelAspectRatio_Original,omitempty"`
	Resolution                                   string `json:"Resolution,omitempty"`
	ServiceKind                                  string `json:"ServiceKind,omitempty"`
	Status                                       string `json:"Status,omitempty"`
	StreamCount                                  string `json:"StreamCount,omitempty"`
	StreamKind                                   string `json:"StreamKind,omitempty"`
	StreamKindID                                 string `json:"StreamKindID,omitempty"`
	StreamKindPos                                string `json:"StreamKindPos,omitempty"`
	StreamOrder                                  string `json:"StreamOrder,omitempty"`
	StreamSize                                   string `json:"StreamSize,omitempty"`
	StreamSizeDemuxed                            string `json:"StreamSize_Demuxed,omitempty"`
	StreamSizeProportion                         string `json:"StreamSize_Proportion,omitempty"`
	Summary                                      string `json:"Summary,omitempty"`
	TaggedDate                                   string `json:"Tagged_Date,omitempty"`
	Title                                        string `json:"Title,omitempty"`
	TransferCharacteristics                      string `json:"transfer_characteristics,omitempty"`
	TransferCharacteristicsOriginal              string `json:"transfer_characteristics_Original,omitempty"`
	TransferCharacteristicsOriginalSource        string `json:"transfer_characteristics_Original_Source,omitempty"`
	TransferCharacteristicsSource                string `json:"transfer_characteristics_Source,omitempty"`
	UniqueID                                     string `json:"UniqueID,omitempty"`
	Width                                        string `json:"Width,omitempty"`
	WidthOffset                                  string `json:"Width_Offset,omitempty"`
	WidthOriginal                                string `json:"Width_Original,omitempty"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

/*
 * Reference: https://github.com/MediaArea/MediaInfoLib/blob/master/Source/Resource/Text/Stream/Menu.csv
 */
type MediaInfoTrackMenu struct {
	TrackType string `json:"@type,omitempty"`

	AlternateGroup           string `json:"AlternateGroup,omitempty"`
	ChaptersPosBegin         string `json:"Chapters_Pos_Begin,omitempty"`
	ChaptersPosEnd           string `json:"Chapters_Pos_End,omitempty"`
	Codec                    string `json:"Codec,omitempty"`
	CodecID                  string `json:"CodecID,omitempty"`
	CodecIDDescription       string `json:"CodecID_Description,omitempty"`
	Count                    string `json:"Count,omitempty"`
	Countries                string `json:"Countries,omitempty"`
	Default                  string `json:"Default,omitempty"`
	Delay                    string `json:"Delay,omitempty"`
	DelayDropFrame           string `json:"Delay_DropFrame,omitempty"`
	DelaySettings            string `json:"Delay_Settings,omitempty"`
	DelaySource              string `json:"Delay_Source,omitempty"`
	Disabled                 string `json:"Disabled,omitempty"`
	Duration                 string `json:"Duration,omitempty"`
	DurationEnd              string `json:"Duration_End,omitempty"`
	DurationStart            string `json:"Duration_Start,omitempty"`
	FirstPacketOrder         string `json:"FirstPacketOrder,omitempty"`
	Forced                   string `json:"Forced,omitempty"`
	Format                   string `json:"Format,omitempty"`
	FormatAdditionalFeatures string `json:"Format_AdditionalFeatures,omitempty"`
	FormatCommercial         string `json:"Format_Commercial,omitempty"`
	FormatCommercialIfAny    string `json:"Format_Commercial_IfAny,omitempty"`
	FormatCompression        string `json:"Format_Compression,omitempty"`
	FormatProfile            string `json:"Format_Profile,omitempty"`
	FormatSettings           string `json:"Format_Settings,omitempty"`
	FormatVersion            string `json:"Format_Version,omitempty"`
	FrameCount               string `json:"FrameCount,omitempty"`
	FrameRate                string `json:"FrameRate,omitempty"`
	FrameRateDen             string `json:"FrameRate_Den,omitempty"`
	FrameRateMode            string `json:"FrameRate_Mode,omitempty"`
	FrameRateNum             string `json:"FrameRate_Num,omitempty"`
	ID                       string `json:"ID,omitempty"`
	Inform                   string `json:"Inform,omitempty"`
	Language                 string `json:"Language,omitempty"`
	LanguageMore             string `json:"Language_More,omitempty"`
	LawRating                string `json:"LawRating,omitempty"`
	LawRatingReason          string `json:"LawRating_Reason,omitempty"`
	List                     string `json:"List,omitempty"`
	ListStreamKind           string `json:"List_StreamKind,omitempty"`
	ListStreamPos            string `json:"List_StreamPos,omitempty"`
	MenuID                   string `json:"MenuID,omitempty"`
	NetworkName              string `json:"NetworkName,omitempty"`
	OriginalSourceMediumID   string `json:"OriginalSourceMedium_ID,omitempty"`
	ServiceChannel           string `json:"ServiceChannel,omitempty"`
	ServiceKind              string `json:"ServiceKind,omitempty"`
	ServiceName              string `json:"ServiceName,omitempty"`
	ServiceProvider          string `json:"ServiceProvider,omitempty"`
	ServiceType              string `json:"ServiceType,omitempty"`
	Status                   string `json:"Status,omitempty"`
	StreamCount              string `json:"StreamCount,omitempty"`
	StreamKind               string `json:"StreamKind,omitempty"`
	StreamKindID             string `json:"StreamKindID,omitempty"`
	StreamKindPos            string `json:"StreamKindPos,omitempty"`
	StreamOrder              string `json:"StreamOrder,omitempty"`
	TimeZones                string `json:"TimeZones,omitempty"`
	Title                    string `json:"Title,omitempty"`
	UniqueID                 string `json:"UniqueID,omitempty"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

/*
 * Reference: https://github.com/MediaArea/MediaInfoLib/blob/master/Source/Resource/Text/Stream/Other.csv
 */
type MediaInfoTrackOther struct {
	TrackType string `json:"@type,omitempty"`

	AlternateGroup                    string `json:"AlternateGroup,omitempty"`
	BitRate                           string `json:"BitRate,omitempty"`
	BitRateEncoded                    string `json:"BitRate_Encoded,omitempty"`
	BitRateMaximum                    string `json:"BitRate_Maximum,omitempty"`
	BitRateMinimum                    string `json:"BitRate_Minimum,omitempty"`
	BitRateMode                       string `json:"BitRate_Mode,omitempty"`
	BitRateNominal                    string `json:"BitRate_Nominal,omitempty"`
	CodecID                           string `json:"CodecID,omitempty"`
	CodecIDDescription                string `json:"CodecID_Description,omitempty"`
	Count                             string `json:"Count,omitempty"`
	Default                           string `json:"Default,omitempty"`
	Delay                             string `json:"Delay,omitempty"`
	DelayDropFrame                    string `json:"Delay_DropFrame,omitempty"`
	DelayOriginal                     string `json:"Delay_Original,omitempty"`
	DelayOriginalDropFrame            string `json:"Delay_Original_DropFrame,omitempty"`
	DelayOriginalSettings             string `json:"Delay_Original_Settings,omitempty"`
	DelayOriginalSource               string `json:"Delay_Original_Source,omitempty"`
	DelaySettings                     string `json:"Delay_Settings,omitempty"`
	DelaySource                       string `json:"Delay_Source,omitempty"`
	Disabled                          string `json:"Disabled,omitempty"`
	Duration                          string `json:"Duration,omitempty"`
	DurationEnd                       string `json:"Duration_End,omitempty"`
	DurationStart                     string `json:"Duration_Start,omitempty"`
	FirstPacketOrder                  string `json:"FirstPacketOrder,omitempty"`
	Forced                            string `json:"Forced,omitempty"`
	Format                            string `json:"Format,omitempty"`
	FormatAdditionalFeatures          string `json:"Format_AdditionalFeatures,omitempty"`
	FormatCommercial                  string `json:"Format_Commercial,omitempty"`
	FormatCommercialIfAny             string `json:"Format_Commercial_IfAny,omitempty"`
	FormatCompression                 string `json:"Format_Compression,omitempty"`
	FormatProfile                     string `json:"Format_Profile,omitempty"`
	FormatSettings                    string `json:"Format_Settings,omitempty"`
	FormatVersion                     string `json:"Format_Version,omitempty"`
	FrameCount                        string `json:"FrameCount,omitempty"`
	FrameRate                         string `json:"FrameRate,omitempty"`
	FrameRateDen                      string `json:"FrameRate_Den,omitempty"`
	FrameRateNum                      string `json:"FrameRate_Num,omitempty"`
	ID                                string `json:"ID,omitempty"`
	Inform                            string `json:"Inform,omitempty"`
	Language                          string `json:"Language,omitempty"`
	LanguageMore                      string `json:"Language_More,omitempty"`
	MenuID                            string `json:"MenuID,omitempty"`
	MuxingMode                        string `json:"MuxingMode,omitempty"`
	OriginalSourceMediumID            string `json:"OriginalSourceMedium_ID,omitempty"`
	ServiceKind                       string `json:"ServiceKind,omitempty"`
	SourceDuration                    string `json:"Source_Duration,omitempty"`
	SourceDurationFirstFrame          string `json:"Source_Duration_FirstFrame,omitempty"`
	SourceDurationLastFrame           string `json:"Source_Duration_LastFrame,omitempty"`
	SourceFrameCount                  string `json:"Source_FrameCount,omitempty"`
	SourceStreamSize                  string `json:"Source_StreamSize,omitempty"`
	SourceStreamSizeEncoded           string `json:"Source_StreamSize_Encoded,omitempty"`
	SourceStreamSizeEncodedProportion string `json:"Source_StreamSize_Encoded_Proportion,omitempty"`
	SourceStreamSizeProportion        string `json:"Source_StreamSize_Proportion,omitempty"`
	Status                            string `json:"Status,omitempty"`
	StreamCount                       string `json:"StreamCount,omitempty"`
	StreamKind                        string `json:"StreamKind,omitempty"`
	StreamKindID                      string `json:"StreamKindID,omitempty"`
	StreamKindPos                     string `json:"StreamKindPos,omitempty"`
	StreamOrder                       string `json:"StreamOrder,omitempty"`
	StreamSize                        string `json:"StreamSize,omitempty"`
	StreamSizeDemuxed                 string `json:"StreamSize_Demuxed,omitempty"`
	StreamSizeEncoded                 string `json:"StreamSize_Encoded,omitempty"`
	StreamSizeEncodedProportion       string `json:"StreamSize_Encoded_Proportion,omitempty"`
	StreamSizeProportion              string `json:"StreamSize_Proportion,omitempty"`
	TimeCodeDropFrame                 string `json:"TimeCode_DropFrame,omitempty"`
	TimeCodeFirstFrame                string `json:"TimeCode_FirstFrame,omitempty"`
	TimeCodeLastFrame                 string `json:"TimeCode_LastFrame,omitempty"`
	TimeCodeSettings                  string `json:"TimeCode_Settings,omitempty"`
	TimeCodeSource                    string `json:"TimeCode_Source,omitempty"`
	TimeCodeStripped                  string `json:"TimeCode_Stripped,omitempty"`
	TimeStampFirstFrame               string `json:"TimeStamp_FirstFrame,omitempty"`
	Title                             string `json:"Title,omitempty"`
	Type                              string `json:"Type,omitempty"`
	UniqueID                          string `json:"UniqueID,omitempty"`
	VideoDelay                        string `json:"Video_Delay,omitempty"`
	Video0Delay                       string `json:"Video0_Delay,omitempty"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

/*
 * Reference: https://github.com/MediaArea/MediaInfoLib/blob/master/Source/Resource/Text/Stream/Text.csv
 */
type MediaInfoTrackText struct {
	TrackType string `json:"@type,omitempty"`

	AlternateGroup                    string `json:"AlternateGroup,omitempty"`
	BitDepth                          string `json:"BitDepth,omitempty"`
	BitRate                           string `json:"BitRate,omitempty"`
	BitRateEncoded                    string `json:"BitRate_Encoded,omitempty"`
	BitRateMaximum                    string `json:"BitRate_Maximum,omitempty"`
	BitRateMinimum                    string `json:"BitRate_Minimum,omitempty"`
	BitRateMode                       string `json:"BitRate_Mode,omitempty"`
	BitRateNominal                    string `json:"BitRate_Nominal,omitempty"`
	ChromaSubsampling                 string `json:"ChromaSubsampling,omitempty"`
	Codec                             string `json:"Codec,omitempty"`
	CodecID                           string `json:"CodecID,omitempty"`
	CodecIDDescription                string `json:"CodecID_Description,omitempty"`
	ColorSpace                        string `json:"ColorSpace,omitempty"`
	CompressionMode                   string `json:"Compression_Mode,omitempty"`
	CompressionRatio                  string `json:"Compression_Ratio,omitempty"`
	Count                             string `json:"Count,omitempty"`
	Default                           string `json:"Default,omitempty"`
	Delay                             string `json:"Delay,omitempty"`
	DelayDropFrame                    string `json:"Delay_DropFrame,omitempty"`
	DelayOriginal                     string `json:"Delay_Original,omitempty"`
	DelayOriginalDropFrame            string `json:"Delay_Original_DropFrame,omitempty"`
	DelayOriginalSettings             string `json:"Delay_Original_Settings,omitempty"`
	DelayOriginalSource               string `json:"Delay_Original_Source,omitempty"`
	DelaySettings                     string `json:"Delay_Settings,omitempty"`
	DelaySource                       string `json:"Delay_Source,omitempty"`
	Disabled                          string `json:"Disabled,omitempty"`
	DisplayAspectRatio                string `json:"DisplayAspectRatio,omitempty"`
	DisplayAspectRatioOriginal        string `json:"DisplayAspectRatio_Original,omitempty"`
	Duration                          string `json:"Duration,omitempty"`
	DurationBase                      string `json:"Duration_Base,omitempty"`
	DurationEnd                       string `json:"Duration_End,omitempty"`
	DurationEndCommand                string `json:"Duration_End_Command,omitempty"`
	DurationFirstFrame                string `json:"Duration_FirstFrame,omitempty"`
	DurationLastFrame                 string `json:"Duration_LastFrame,omitempty"`
	DurationStart                     string `json:"Duration_Start,omitempty"`
	DurationStartCommand              string `json:"Duration_Start_Command,omitempty"`
	DurationStart2End                 string `json:"Duration_Start2End,omitempty"`
	ElementCount                      string `json:"ElementCount,omitempty"`
	EncodedApplication                string `json:"Encoded_Application,omitempty"`
	EncodedApplicationCompanyName     string `json:"Encoded_Application_CompanyName,omitempty"`
	EncodedApplicationName            string `json:"Encoded_Application_Name,omitempty"`
	EncodedApplicationUrl             string `json:"Encoded_Application_Url,omitempty"`
	EncodedApplicationVersion         string `json:"Encoded_Application_Version,omitempty"`
	EncodedDate                       string `json:"Encoded_Date,omitempty"`
	EncodedLibrary                    string `json:"Encoded_Library,omitempty"`
	EncodedLibraryCompanyName         string `json:"Encoded_Library_CompanyName,omitempty"`
	EncodedLibraryDate                string `json:"Encoded_Library_Date,omitempty"`
	EncodedLibraryName                string `json:"Encoded_Library_Name,omitempty"`
	EncodedLibrarySettings            string `json:"Encoded_Library_Settings,omitempty"`
	EncodedLibraryVersion             string `json:"Encoded_Library_Version,omitempty"`
	EncodedOperatingSystem            string `json:"Encoded_OperatingSystem,omitempty"`
	Encryption                        string `json:"Encryption,omitempty"`
	EventsMinDuration                 string `json:"Events_MinDuration,omitempty"`
	EventsPaintOn                     string `json:"Events_PaintOn,omitempty"`
	EventsPopOn                       string `json:"Events_PopOn,omitempty"`
	EventsRollUp                      string `json:"Events_RollUp,omitempty"`
	EventsTotal                       string `json:"Events_Total,omitempty"`
	FirstDisplayDelayFrames           string `json:"FirstDisplay_Delay_Frames,omitempty"`
	FirstDisplayType                  string `json:"FirstDisplay_Type,omitempty"`
	FirstPacketOrder                  string `json:"FirstPacketOrder,omitempty"`
	Forced                            string `json:"Forced,omitempty"`
	Format                            string `json:"Format,omitempty"`
	FormatAdditionalFeatures          string `json:"Format_AdditionalFeatures,omitempty"`
	FormatCommercial                  string `json:"Format_Commercial,omitempty"`
	FormatCommercialIfAny             string `json:"Format_Commercial_IfAny,omitempty"`
	FormatCompression                 string `json:"Format_Compression,omitempty"`
	FormatProfile                     string `json:"Format_Profile,omitempty"`
	FormatSettings                    string `json:"Format_Settings,omitempty"`
	FormatSettingsWrapping            string `json:"Format_Settings_Wrapping,omitempty"`
	FormatVersion                     string `json:"Format_Version,omitempty"`
	FrameCount                        string `json:"FrameCount,omitempty"`
	FrameRate                         string `json:"FrameRate,omitempty"`
	FrameRateDen                      string `json:"FrameRate_Den,omitempty"`
	FrameRateMaximum                  string `json:"FrameRate_Maximum,omitempty"`
	FrameRateMinimum                  string `json:"FrameRate_Minimum,omitempty"`
	FrameRateMode                     string `json:"FrameRate_Mode,omitempty"`
	FrameRateModeOriginal             string `json:"FrameRate_Mode_Original,omitempty"`
	FrameRateNominal                  string `json:"FrameRate_Nominal,omitempty"`
	FrameRateNum                      string `json:"FrameRate_Num,omitempty"`
	FrameRateOriginal                 string `json:"FrameRate_Original,omitempty"`
	FrameRateOriginalDen              string `json:"FrameRate_Original_Den,omitempty"`
	FrameRateOriginalNum              string `json:"FrameRate_Original_Num,omitempty"`
	Height                            string `json:"Height,omitempty"`
	ID                                string `json:"ID,omitempty"`
	Inform                            string `json:"Inform,omitempty"`
	InternetMediaType                 string `json:"InternetMediaType,omitempty"`
	Language                          string `json:"Language,omitempty"`
	LanguageMore                      string `json:"Language_More,omitempty"`
	LinesCount                        string `json:"Lines_Count,omitempty"`
	LinesMaxCountPerEvent             string `json:"Lines_MaxCountPerEvent,omitempty"`
	MenuID                            string `json:"MenuID,omitempty"`
	MuxingMode                        string `json:"MuxingMode,omitempty"`
	MuxingModeMoreInfo                string `json:"MuxingMode_MoreInfo,omitempty"`
	OriginalSourceMediumID            string `json:"OriginalSourceMedium_ID,omitempty"`
	Resolution                        string `json:"Resolution,omitempty"`
	ServiceKind                       string `json:"ServiceKind,omitempty"`
	SourceDuration                    string `json:"Source_Duration,omitempty"`
	SourceDurationFirstFrame          string `json:"Source_Duration_FirstFrame,omitempty"`
	SourceDurationLastFrame           string `json:"Source_Duration_LastFrame,omitempty"`
	SourceFrameCount                  string `json:"Source_FrameCount,omitempty"`
	SourceStreamSize                  string `json:"Source_StreamSize,omitempty"`
	SourceStreamSizeEncoded           string `json:"Source_StreamSize_Encoded,omitempty"`
	SourceStreamSizeEncodedProportion string `json:"Source_StreamSize_Encoded_Proportion,omitempty"`
	SourceStreamSizeProportion        string `json:"Source_StreamSize_Proportion,omitempty"`
	Status                            string `json:"Status,omitempty"`
	StreamCount                       string `json:"StreamCount,omitempty"`
	StreamKind                        string `json:"StreamKind,omitempty"`
	StreamKindID                      string `json:"StreamKindID,omitempty"`
	StreamKindPos                     string `json:"StreamKindPos,omitempty"`
	StreamOrder                       string `json:"StreamOrder,omitempty"`
	StreamSize                        string `json:"StreamSize,omitempty"`
	StreamSizeDemuxed                 string `json:"StreamSize_Demuxed,omitempty"`
	StreamSizeEncoded                 string `json:"StreamSize_Encoded,omitempty"`
	StreamSizeEncodedProportion       string `json:"StreamSize_Encoded_Proportion,omitempty"`
	StreamSizeProportion              string `json:"StreamSize_Proportion,omitempty"`
	Summary                           string `json:"Summary,omitempty"`
	TaggedDate                        string `json:"Tagged_Date,omitempty"`
	TimeCodeDropFrame                 string `json:"TimeCode_DropFrame,omitempty"`
	TimeCodeFirstFrame                string `json:"TimeCode_FirstFrame,omitempty"`
	TimeCodeLastFrame                 string `json:"TimeCode_LastFrame,omitempty"`
	TimeCodeMaxFrameNumber            string `json:"TimeCode_MaxFrameNumber,omitempty"`
	TimeCodeMaxFrameNumberTheory      string `json:"TimeCode_MaxFrameNumber_Theory,omitempty"`
	TimeCodeSettings                  string `json:"TimeCode_Settings,omitempty"`
	TimeCodeSource                    string `json:"TimeCode_Source,omitempty"`
	Title                             string `json:"Title,omitempty"`
	UniqueID                          string `json:"UniqueID,omitempty"`
	VideoDelay                        string `json:"Video_Delay,omitempty"`
	Video0Delay                       string `json:"Video0_Delay,omitempty"`
	Width                             string `json:"Width,omitempty"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

/*
 * Reference: https://github.com/MediaArea/MediaInfoLib/blob/master/Source/Resource/Text/Stream/Video.csv
 */
type MediaInfoTrackVideo struct {
	TrackType string `json:"@type,omitempty"`

	ActiveDisplayAspectRatio                     string `json:"Active_DisplayAspectRatio,omitempty"`
	ActiveHeight                                 string `json:"Active_Height,omitempty"`
	ActiveWidth                                  string `json:"Active_Width,omitempty"`
	ActiveFormatDescription                      string `json:"ActiveFormatDescription,omitempty"`
	ActiveFormatDescriptionMuxingMode            string `json:"ActiveFormatDescription_MuxingMode,omitempty"`
	Alignment                                    string `json:"Alignment,omitempty"`
	AlternateGroup                               string `json:"AlternateGroup,omitempty"`
	BitDepth                                     string `json:"BitDepth,omitempty"`
	BitRate                                      string `json:"BitRate,omitempty"`
	BitRateEncoded                               string `json:"BitRate_Encoded,omitempty"`
	BitRateMaximum                               string `json:"BitRate_Maximum,omitempty"`
	BitRateMinimum                               string `json:"BitRate_Minimum,omitempty"`
	BitRateMode                                  string `json:"BitRate_Mode,omitempty"`
	BitRateNominal                               string `json:"BitRate_Nominal,omitempty"`
	Bits                                         string `json:"Bits,omitempty"`
	BufferSize                                   string `json:"BufferSize,omitempty"`
	ChromaSubsampling                            string `json:"ChromaSubsampling,omitempty"`
	ChromaSubsamplingPosition                    string `json:"ChromaSubsampling_Position,omitempty"`
	Codec                                        string `json:"Codec,omitempty"`
	CodecDescription                             string `json:"Codec_Description,omitempty"`
	CodecProfile                                 string `json:"Codec_Profile,omitempty"`
	CodecSettings                                string `json:"Codec_Settings,omitempty"`
	CodecSettingsBVOP                            string `json:"Codec_Settings_BVOP,omitempty"`
	CodecSettingsCABAC                           string `json:"Codec_Settings_CABAC,omitempty"`
	CodecSettingsGMC                             string `json:"Codec_Settings_GMC,omitempty"`
	CodecSettingsMatrix                          string `json:"Codec_Settings_Matrix,omitempty"`
	CodecSettingsMatrixData                      string `json:"Codec_Settings_Matrix_Data,omitempty"`
	CodecSettingsPacketBitStream                 string `json:"Codec_Settings_PacketBitStream,omitempty"`
	CodecSettingsQPel                            string `json:"Codec_Settings_QPel,omitempty"`
	CodecSettingsRefFrames                       string `json:"Codec_Settings_RefFrames,omitempty"`
	CodecID                                      string `json:"CodecID,omitempty"`
	CodecIDDescription                           string `json:"CodecID_Description,omitempty"`
	Colorimetry                                  string `json:"Colorimetry,omitempty"`
	ColorSpace                                   string `json:"ColorSpace,omitempty"`
	ColourDescriptionPresent                     string `json:"colour_description_present,omitempty"`
	ColourDescriptionPresentOriginal             string `json:"colour_description_present_Original,omitempty"`
	ColourDescriptionPresentOriginalSource       string `json:"colour_description_present_Original_Source,omitempty"`
	ColourDescriptionPresentSource               string `json:"colour_description_present_Source,omitempty"`
	ColourPrimaries                              string `json:"colour_primaries,omitempty"`
	ColourPrimariesOriginal                      string `json:"colour_primaries_Original,omitempty"`
	ColourPrimariesOriginalSource                string `json:"colour_primaries_Original_Source,omitempty"`
	ColourPrimariesSource                        string `json:"colour_primaries_Source,omitempty"`
	ColourRange                                  string `json:"colour_range,omitempty"`
	ColourRangeOriginal                          string `json:"colour_range_Original,omitempty"`
	ColourRangeOriginalSource                    string `json:"colour_range_Original_Source,omitempty"`
	ColourRangeSource                            string `json:"colour_range_Source,omitempty"`
	CompressionMode                              string `json:"Compression_Mode,omitempty"`
	CompressionRatio                             string `json:"Compression_Ratio,omitempty"`
	Count                                        string `json:"Count,omitempty"`
	Default                                      string `json:"Default,omitempty"`
	Delay                                        string `json:"Delay,omitempty"`
	DelayDropFrame                               string `json:"Delay_DropFrame,omitempty"`
	DelayOriginal                                string `json:"Delay_Original,omitempty"`
	DelayOriginalDropFrame                       string `json:"Delay_Original_DropFrame,omitempty"`
	DelayOriginalSettings                        string `json:"Delay_Original_Settings,omitempty"`
	DelayOriginalSource                          string `json:"Delay_Original_Source,omitempty"`
	DelaySettings                                string `json:"Delay_Settings,omitempty"`
	DelaySource                                  string `json:"Delay_Source,omitempty"`
	Disabled                                     string `json:"Disabled,omitempty"`
	DisplayAspectRatio                           string `json:"DisplayAspectRatio,omitempty"`
	DisplayAspectRatioCleanAperture              string `json:"DisplayAspectRatio_CleanAperture,omitempty"`
	DisplayAspectRatioOriginal                   string `json:"DisplayAspectRatio_Original,omitempty"`
	Duration                                     string `json:"Duration,omitempty"`
	DurationFirstFrame                           string `json:"Duration_FirstFrame,omitempty"`
	DurationLastFrame                            string `json:"Duration_LastFrame,omitempty"`
	EncodedApplication                           string `json:"Encoded_Application,omitempty"`
	EncodedApplicationCompanyName                string `json:"Encoded_Application_CompanyName,omitempty"`
	EncodedApplicationName                       string `json:"Encoded_Application_Name,omitempty"`
	EncodedApplicationUrl                        string `json:"Encoded_Application_Url,omitempty"`
	EncodedApplicationVersion                    string `json:"Encoded_Application_Version,omitempty"`
	EncodedDate                                  string `json:"Encoded_Date,omitempty"`
	EncodedLibrary                               string `json:"Encoded_Library,omitempty"`
	EncodedLibraryCompanyName                    string `json:"Encoded_Library_CompanyName,omitempty"`
	EncodedLibraryDate                           string `json:"Encoded_Library_Date,omitempty"`
	EncodedLibraryName                           string `json:"Encoded_Library_Name,omitempty"`
	EncodedLibrarySettings                       string `json:"Encoded_Library_Settings,omitempty"`
	EncodedLibraryVersion                        string `json:"Encoded_Library_Version,omitempty"`
	EncodedOperatingSystem                       string `json:"Encoded_OperatingSystem,omitempty"`
	Encryption                                   string `json:"Encryption,omitempty"`
	FirstPacketOrder                             string `json:"FirstPacketOrder,omitempty"`
	Forced                                       string `json:"Forced,omitempty"`
	Format                                       string `json:"Format,omitempty"`
	FormatAdditionalFeatures                     string `json:"Format_AdditionalFeatures,omitempty"`
	FormatCommercial                             string `json:"Format_Commercial,omitempty"`
	FormatCommercialIfAny                        string `json:"Format_Commercial_IfAny,omitempty"`
	FormatCompression                            string `json:"Format_Compression,omitempty"`
	FormatLevel                                  string `json:"Format_Level,omitempty"`
	FormatProfile                                string `json:"Format_Profile,omitempty"`
	FormatSettings                               string `json:"Format_Settings,omitempty"`
	FormatSettingsBVOP                           string `json:"Format_Settings_BVOP,omitempty"`
	FormatSettingsCABAC                          string `json:"Format_Settings_CABAC,omitempty"`
	FormatSettingsEndianness                     string `json:"Format_Settings_Endianness,omitempty"`
	FormatSettingsFrameMode                      string `json:"Format_Settings_FrameMode,omitempty"`
	FormatSettingsGMC                            string `json:"Format_Settings_GMC,omitempty"`
	FormatSettingsGOP                            string `json:"Format_Settings_GOP,omitempty"`
	FormatSettingsMatrix                         string `json:"Format_Settings_Matrix,omitempty"`
	FormatSettingsMatrixData                     string `json:"Format_Settings_Matrix_Data,omitempty"`
	FormatSettingsPacking                        string `json:"Format_Settings_Packing,omitempty"`
	FormatSettingsPictureStructure               string `json:"Format_Settings_PictureStructure,omitempty"`
	FormatSettingsPulldown                       string `json:"Format_Settings_Pulldown,omitempty"`
	FormatSettingsQPel                           string `json:"Format_Settings_QPel,omitempty"`
	FormatSettingsRefFrames                      string `json:"Format_Settings_RefFrames,omitempty"`
	FormatSettingsSliceCount                     string `json:"Format_Settings_SliceCount,omitempty"`
	FormatSettingsWrapping                       string `json:"Format_Settings_Wrapping,omitempty"`
	FormatTier                                   string `json:"Format_Tier,omitempty"`
	FormatVersion                                string `json:"Format_Version,omitempty"`
	FrameCount                                   string `json:"FrameCount,omitempty"`
	FrameRate                                    string `json:"FrameRate,omitempty"`
	FrameRateDen                                 string `json:"FrameRate_Den,omitempty"`
	FrameRateMaximum                             string `json:"FrameRate_Maximum,omitempty"`
	FrameRateMinimum                             string `json:"FrameRate_Minimum,omitempty"`
	FrameRateMode                                string `json:"FrameRate_Mode,omitempty"`
	FrameRateModeOriginal                        string `json:"FrameRate_Mode_Original,omitempty"`
	FrameRateNominal                             string `json:"FrameRate_Nominal,omitempty"`
	FrameRateNum                                 string `json:"FrameRate_Num,omitempty"`
	FrameRateOriginal                            string `json:"FrameRate_Original,omitempty"`
	FrameRateOriginalDen                         string `json:"FrameRate_Original_Den,omitempty"`
	FrameRateOriginalNum                         string `json:"FrameRate_Original_Num,omitempty"`
	FrameRateReal                                string `json:"FrameRate_Real,omitempty"`
	GopOpenClosed                                string `json:"Gop_OpenClosed,omitempty"`
	GopOpenClosedFirstFrame                      string `json:"Gop_OpenClosed_FirstFrame,omitempty"`
	HDRFormat                                    string `json:"HDR_Format,omitempty"`
	HDRFormatCommercial                          string `json:"HDR_Format_Commercial,omitempty"`
	HDRFormatCompatibility                       string `json:"HDR_Format_Compatibility,omitempty"`
	HDRFormatCompression                         string `json:"HDR_Format_Compression,omitempty"`
	HDRFormatLevel                               string `json:"HDR_Format_Level,omitempty"`
	HDRFormatProfile                             string `json:"HDR_Format_Profile,omitempty"`
	HDRFormatSettings                            string `json:"HDR_Format_Settings,omitempty"`
	HDRFormatVersion                             string `json:"HDR_Format_Version,omitempty"`
	Height                                       string `json:"Height,omitempty"`
	HeightCleanAperture                          string `json:"Height_CleanAperture,omitempty"`
	HeightOffset                                 string `json:"Height_Offset,omitempty"`
	HeightOriginal                               string `json:"Height_Original,omitempty"`
	ID                                           string `json:"ID,omitempty"`
	Inform                                       string `json:"Inform,omitempty"`
	Interlacement                                string `json:"Interlacement,omitempty"`
	InternetMediaType                            string `json:"InternetMediaType,omitempty"`
	Language                                     string `json:"Language,omitempty"`
	LanguageMore                                 string `json:"Language_More,omitempty"`
	MasteringDisplayColorPrimaries               string `json:"MasteringDisplay_ColorPrimaries,omitempty"`
	MasteringDisplayColorPrimariesOriginal       string `json:"MasteringDisplay_ColorPrimaries_Original,omitempty"`
	MasteringDisplayColorPrimariesOriginalSource string `json:"MasteringDisplay_ColorPrimaries_Original_Source,omitempty"`
	MasteringDisplayColorPrimariesSource         string `json:"MasteringDisplay_ColorPrimaries_Source,omitempty"`
	MasteringDisplayLuminance                    string `json:"MasteringDisplay_Luminance,omitempty"`
	MasteringDisplayLuminanceOriginal            string `json:"MasteringDisplay_Luminance_Original,omitempty"`
	MasteringDisplayLuminanceOriginalSource      string `json:"MasteringDisplay_Luminance_Original_Source,omitempty"`
	MasteringDisplayLuminanceSource              string `json:"MasteringDisplay_Luminance_Source,omitempty"`
	MatrixCoefficients                           string `json:"matrix_coefficients,omitempty"`
	MatrixCoefficientsOriginal                   string `json:"matrix_coefficients_Original,omitempty"`
	MatrixCoefficientsOriginalSource             string `json:"matrix_coefficients_Original_Source,omitempty"`
	MatrixCoefficientsSource                     string `json:"matrix_coefficients_Source,omitempty"`
	MaxCLL                                       string `json:"MaxCLL,omitempty"`
	MaxCLLOriginal                               string `json:"MaxCLL_Original,omitempty"`
	MaxCLLOriginalSource                         string `json:"MaxCLL_Original_Source,omitempty"`
	MaxCLLSource                                 string `json:"MaxCLL_Source,omitempty"`
	MaxFALL                                      string `json:"MaxFALL,omitempty"`
	MaxFALLOriginal                              string `json:"MaxFALL_Original,omitempty"`
	MaxFALLOriginalSource                        string `json:"MaxFALL_Original_Source,omitempty"`
	MaxFALLSource                                string `json:"MaxFALL_Source,omitempty"`
	MenuID                                       string `json:"MenuID,omitempty"`
	MultiViewBaseProfile                         string `json:"MultiView_BaseProfile,omitempty"`
	MultiViewCount                               string `json:"MultiView_Count,omitempty"`
	MultiViewLayout                              string `json:"MultiView_Layout,omitempty"`
	MuxingMode                                   string `json:"MuxingMode,omitempty"`
	OriginalSourceMediumID                       string `json:"OriginalSourceMedium_ID,omitempty"`
	PixelAspectRatio                             string `json:"PixelAspectRatio,omitempty"`
	PixelAspectRatioCleanAperture                string `json:"PixelAspectRatio_CleanAperture,omitempty"`
	PixelAspectRatioOriginal                     string `json:"PixelAspectRatio_Original,omitempty"`
	Resolution                                   string `json:"Resolution,omitempty"`
	Rotation                                     string `json:"Rotation,omitempty"`
	SampledHeight                                string `json:"Sampled_Height,omitempty"`
	SampledWidth                                 string `json:"Sampled_Width,omitempty"`
	ScanOrder                                    string `json:"ScanOrder,omitempty"`
	ScanOrderOriginal                            string `json:"ScanOrder_Original,omitempty"`
	ScanOrderStored                              string `json:"ScanOrder_Stored,omitempty"`
	ScanOrderStoredDisplayedInverted             string `json:"ScanOrder_StoredDisplayedInverted,omitempty"`
	ScanType                                     string `json:"ScanType,omitempty"`
	ScanTypeOriginal                             string `json:"ScanType_Original,omitempty"`
	ScanTypeStoreMethod                          string `json:"ScanType_StoreMethod,omitempty"`
	ScanTypeStoreMethodFieldsPerBlock            string `json:"ScanType_StoreMethod_FieldsPerBlock,omitempty"`
	ServiceKind                                  string `json:"ServiceKind,omitempty"`
	SourceDuration                               string `json:"Source_Duration,omitempty"`
	SourceDurationFirstFrame                     string `json:"Source_Duration_FirstFrame,omitempty"`
	SourceDurationLastFrame                      string `json:"Source_Duration_LastFrame,omitempty"`
	SourceFrameCount                             string `json:"Source_FrameCount,omitempty"`
	SourceStreamSize                             string `json:"Source_StreamSize,omitempty"`
	SourceStreamSizeEncoded                      string `json:"Source_StreamSize_Encoded,omitempty"`
	SourceStreamSizeEncodedProportion            string `json:"Source_StreamSize_Encoded_Proportion,omitempty"`
	SourceStreamSizeProportion                   string `json:"Source_StreamSize_Proportion,omitempty"`
	Standard                                     string `json:"Standard,omitempty"`
	Status                                       string `json:"Status,omitempty"`
	StoredHeight                                 string `json:"Stored_Height,omitempty"`
	StoredWidth                                  string `json:"Stored_Width,omitempty"`
	StreamCount                                  string `json:"StreamCount,omitempty"`
	StreamKind                                   string `json:"StreamKind,omitempty"`
	StreamKindID                                 string `json:"StreamKindID,omitempty"`
	StreamKindPos                                string `json:"StreamKindPos,omitempty"`
	StreamOrder                                  string `json:"StreamOrder,omitempty"`
	StreamSize                                   string `json:"StreamSize,omitempty"`
	StreamSizeDemuxed                            string `json:"StreamSize_Demuxed,omitempty"`
	StreamSizeEncoded                            string `json:"StreamSize_Encoded,omitempty"`
	StreamSizeEncodedProportion                  string `json:"StreamSize_Encoded_Proportion,omitempty"`
	StreamSizeProportion                         string `json:"StreamSize_Proportion,omitempty"`
	TaggedDate                                   string `json:"Tagged_Date,omitempty"`
	TimeCodeDropFrame                            string `json:"TimeCode_DropFrame,omitempty"`
	TimeCodeFirstFrame                           string `json:"TimeCode_FirstFrame,omitempty"`
	TimeCodeLastFrame                            string `json:"TimeCode_LastFrame,omitempty"`
	TimeCodeSettings                             string `json:"TimeCode_Settings,omitempty"`
	TimeCodeSource                               string `json:"TimeCode_Source,omitempty"`
	TimeStampFirstFrame                          string `json:"TimeStamp_FirstFrame,omitempty"`
	Title                                        string `json:"Title,omitempty"`
	TransferCharacteristics                      string `json:"transfer_characteristics,omitempty"`
	TransferCharacteristicsOriginal              string `json:"transfer_characteristics_Original,omitempty"`
	TransferCharacteristicsOriginalSource        string `json:"transfer_characteristics_Original_Source,omitempty"`
	TransferCharacteristicsSource                string `json:"transfer_characteristics_Source,omitempty"`
	UniqueID                                     string `json:"UniqueID,omitempty"`
	Width                                        string `json:"Width,omitempty"`
	WidthCleanAperture                           string `json:"Width_CleanAperture,omitempty"`
	WidthOffset                                  string `json:"Width_Offset,omitempty"`
	WidthOriginal                                string `json:"Width_Original,omitempty"`

	Extra map[string]interface{} `json:"extra,omitempty"`
}

func DecodeMediaInfoJson(jsonString string) (*MediaInfoReport, error) {
	jsonData := []byte(jsonString)
	type mediaInfoReport struct {
		CreatingLibrary *MediaInfoLibrary `json:"creatingLibrary,omitempty"`
		Media           struct {
			Ref    string            `json:"@ref"`
			Tracks []*MediaInfoTrack `json:"track,omitempty"`
		} `json:"media"`
	}
	tmpReport := &mediaInfoReport{}
	if err := json.Unmarshal(jsonData, tmpReport); err != nil {
		return nil, err
	}

	report := &MediaInfoReport{
		CreatingLibrary: tmpReport.CreatingLibrary,
		Media: &MediaInfoMedia{
			Ref:           tmpReport.Media.Ref,
			AudioTracks:   []*MediaInfoTrackAudio{},
			GeneralTracks: []*MediaInfoTrackGeneral{},
			GenericTracks: []*MediaInfoTrackGeneric{},
			ImageTracks:   []*MediaInfoTrackImage{},
			MenuTracks:    []*MediaInfoTrackMenu{},
			OtherTracks:   []*MediaInfoTrackOther{},
			TextTracks:    []*MediaInfoTrackText{},
			VideoTracks:   []*MediaInfoTrackVideo{},
		},
	}
	for _, t := range tmpReport.Media.Tracks {
		if t.General != nil {
			report.Media.GeneralTracks = append(report.Media.GeneralTracks, t.General)
		} else if t.Video != nil {
			report.Media.VideoTracks = append(report.Media.VideoTracks, t.Video)
		} else if t.Audio != nil {
			report.Media.AudioTracks = append(report.Media.AudioTracks, t.Audio)
		} else if t.Text != nil {
			report.Media.TextTracks = append(report.Media.TextTracks, t.Text)
		} else if t.Image != nil {
			report.Media.ImageTracks = append(report.Media.ImageTracks, t.Image)
		} else if t.Menu != nil {
			report.Media.MenuTracks = append(report.Media.MenuTracks, t.Menu)
		} else if t.Other != nil {
			report.Media.OtherTracks = append(report.Media.OtherTracks, t.Other)
		} else if t.Generic != nil {
			report.Media.GenericTracks = append(report.Media.GenericTracks, t.Generic)
		}
	}
	return report, nil
}
