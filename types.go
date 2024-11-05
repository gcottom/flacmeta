package flacmeta

import "io"

var tagFieldMapping = map[string]string{FIELD_ALBUM: structFieldAlbum, FIELD_ALBUMARTIST: structFieldAlbumArtist, FIELD_ARTIST: structFieldArtist, FIELD_BPM: structFieldBPM, FIELD_CONTACT: structFieldContact, FIELD_COMPOSER: structFieldComposer,
	FIELD_COPYRIGHT: structFieldCopyright, FIELD_DATE: structFieldDate, FIELD_DESCRIPTION: structFieldDescription, FIELD_DISCNUMBER: structFieldDiscNumber, FIELD_DISCTOTAL: structFieldDiscTotal, FIELD_ENCODER: structFieldEncoder, FIELD_GENRE: structFieldGenre, FIELD_ISRC: structFieldISRC, FIELD_LICENSE: structFieldLicense, FIELD_LOCATION: structFieldLocation, FIELD_ORGANIZATION: structFieldOrganization, FIELD_PERFORMER: structFieldPerformer, FIELD_TITLE: structFieldTitle,
	FIELD_TRACKNUMBER: structFieldTrackNumber, FIELD_TRACKTOTAL: structFieldTrackTotal, FIELD_VERSION: structFieldVersion}

const (
	// FIELD_TITLE Track/Work name
	FIELD_TITLE = "TITLE"
	// FIELD_VERSION The version field may be used to differentiate multiple versions of the same track title in a single collection. (e.g. remix info)
	FIELD_VERSION = "VERSION"
	// FIELD_ALBUM The collection name to which this track belongs
	FIELD_ALBUM = "ALBUM"
	// FIELD_TRACKNUMBER The track number of this piece if part of a specific larger collection or album
	FIELD_TRACKNUMBER = "TRACKNUMBER"
	// FIELD_ARTIST The artist generally considered responsible for the work. In popular music this is usually the performing band or singer. For classical music it would be the composer. For an audio book it would be the author of the original text.
	FIELD_ARTIST      = "ARTIST"
	FIELD_ALBUMARTIST = "ALBUMARTIST"
	FIELD_BPM         = "BPM"
	FIELD_COMPOSER    = "COMPOSER"
	FIELD_DISCNUMBER  = "DISCNUMBER"
	FIELD_DISCTOTAL   = "DISCTOTAL"
	FIELD_ENCODER     = "ENCODER"
	FIELD_TRACKTOTAL  = "TRACKTOTAL"
	// FIELD_PERFORMER The artist(s) who performed the work. In classical music this would be the conductor, orchestra, soloists. In an audio book it would be the actor who did the reading. In popular music this is typically the same as the ARTIST and is omitted.
	FIELD_PERFORMER = "PERFORMER"
	// FIELD_COPYRIGHT Copyright attribution, e.g., '2001 Nobody's Band' or '1999 Jack Moffitt'
	FIELD_COPYRIGHT = "COPYRIGHT"
	// FIELD_LICENSE License information, eg, 'All Rights Reserved', 'Any Use Permitted', a URL to a license such as a Creative Commons license ("www.creativecommons.org/blahblah/license.html") or the EFF Open Audio License ('distributed under the terms of the Open Audio License. see http://www.eff.org/IP/Open_licenses/eff_oal.html for details'), etc.
	FIELD_LICENSE = "LICENSE"
	// FIELD_ORGANIZATION Name of the organization producing the track (i.e. the 'record label')
	FIELD_ORGANIZATION = "ORGANIZATION"
	// FIELD_DESCRIPTION A short text description of the contents
	FIELD_DESCRIPTION = "DESCRIPTION"
	// FIELD_GENRE A short text indication of music genre
	FIELD_GENRE = "GENRE"
	// FIELD_DATE Date the track was recorded
	FIELD_DATE = "DATE"
	// FIELD_LOCATION Location where track was recorded
	FIELD_LOCATION = "LOCATION"
	// FIELD_CONTACT Contact information for the creators or distributors of the track. This could be a URL, an email address, the physical address of the producing label.
	FIELD_CONTACT = "CONTACT"
	// FIELD_ISRC ISRC number for the track; see the ISRC intro page for more information on ISRC numbers.
	FIELD_ISRC = "ISRC"
)

const (
	structFieldAlbum        = "Album"
	structFieldAlbumArtist  = "AlbumArtist"
	structFieldArtist       = "Artist"
	structFieldBPM          = "BPM"
	structFieldContact      = "Contact"
	structFieldComposer     = "Composer"
	structFieldCopyright    = "Copyright"
	structFieldDate         = "Date"
	structFieldDescription  = "Description"
	structFieldDiscNumber   = "DiscNumber"
	structFieldDiscTotal    = "DiscTotal"
	structFieldEncoder      = "Encoder"
	structFieldGenre        = "Genre"
	structFieldISRC         = "ISRC"
	structFieldLicense      = "License"
	structFieldLocation     = "Location"
	structFieldOrganization = "Organization"
	structFieldPerformer    = "Performer"
	structFieldTitle        = "Title"
	structFieldTrackNumber  = "TrackNumber"
	structFieldTrackTotal   = "TrackTotal"
	structFieldVersion      = "Version"
)
const (
	// MIMEURL is the MIME string indicating that imgData is a URL pointing to the image
	MIMEURL = "-->"
)

// BlockType representation of types of FLAC Metadata Block
type BlockType int

// BlockData data in a FLAC Metadata Block. Custom Metadata decoders and modifiers should accept/modify whole MetaDataBlock instead.
type BlockData []byte

const (
	// StreamInfo METADATA_BLOCK_STREAMINFO
	// This block has information about the whole stream, like sample rate, number of channels, total number of samples, etc. It must be present as the first metadata block in the stream. Other metadata blocks may follow, and ones that the decoder doesn't understand, it will skip.
	StreamInfo BlockType = iota
	// Padding METADATA_BLOCK_PADDING
	// This block allows for an arbitrary amount of padding. The contents of a PADDING block have no meaning. This block is useful when it is known that metadata will be edited after encoding; the user can instruct the encoder to reserve a PADDING block of sufficient size so that when metadata is added, it will simply overwrite the padding (which is relatively quick) instead of having to insert it into the right place in the existing file (which would normally require rewriting the entire file).
	Padding
	// Application METADATA_BLOCK_APPLICATION
	// This block is for use by third-party applications. The only mandatory field is a 32-bit identifier. This ID is granted upon request to an application by the FLAC maintainers. The remainder is of the block is defined by the registered application. Visit the registration page if you would like to register an ID for your application with FLAC.
	Application
	// SeekTable METADATA_BLOCK_SEEKTABLE
	// This is an optional block for storing seek points. It is possible to seek to any given sample in a FLAC stream without a seek table, but the delay can be unpredictable since the bitrate may vary widely within a stream. By adding seek points to a stream, this delay can be significantly reduced. Each seek point takes 18 bytes, so 1% resolution within a stream adds less than 2k. There can be only one SEEKTABLE in a stream, but the table can have any number of seek points. There is also a special 'placeholder' seekpoint which will be ignored by decoders but which can be used to reserve space for future seek point insertion.
	SeekTable
	// VorbisComment METADATA_BLOCK_VORBIS_COMMENT
	// This block is for storing a list of human-readable name/value pairs. Values are encoded using UTF-8. It is an implementation of the Vorbis comment specification (without the framing bit). This is the only officially supported tagging mechanism in FLAC. There may be only one VORBIS_COMMENT block in a stream. In some external documentation, Vorbis comments are called FLAC tags to lessen confusion.
	VorbisComment
	// CueSheet METADATA_BLOCK_CUESHEET
	// This block is for storing various information that can be used in a cue sheet. It supports track and index points, compatible with Red Book CD digital audio discs, as well as other CD-DA metadata such as media catalog number and track ISRCs. The CUESHEET block is especially useful for backing up CD-DA discs, but it can be used as a general purpose cueing mechanism for playback.
	CueSheet
	// Picture METADATA_BLOCK_PICTURE
	// This block is for storing pictures associated with the file, most commonly cover art from CDs. There may be more than one PICTURE block in a file. The picture format is similar to the APIC frame in ID3v2. The PICTURE block has a type, MIME type, and UTF-8 description like ID3v2, and supports external linking via URL (though this is discouraged). The differences are that there is no uniqueness constraint on the description field, and the MIME type is mandatory. The FLAC PICTURE block also includes the resolution, color depth, and palette size so that the client can search for a suitable picture without having to scan them all.
	Picture
	// Reserved Reserved Metadata Block Types
	Reserved
	// Invalid Invalid Metadata Block Type
	Invalid BlockType = 127
)

// MetaDataBlock is the struct representation of a FLAC Metadata Block
type MetaDataBlock struct {
	Type BlockType
	Data BlockData
}

type FLACFile struct {
	Meta         []*MetaDataBlock
	Frames       FrameData
	StreamReader io.ReadSeeker
}

type FrameData []byte

type PictureType uint32

const (
	PictureTypeOther PictureType = iota
	PictureTypeFileIcon
	PictureTypeOtherIcon
	PictureTypeFrontCover
	PictureTypeBackCover
	PictureTypeLeaflet
	PictureTypeMedia
	PictureTypeLeadArtist
	PictureTypeArtist
	PictureTypeConductor
	PictureTypeBand
	PictureTypeComposer
	PictureTypeLyricist
	PictureTypeRecordingLocation
	PictureTypeDuringRecording
	PictureTypeDuringPerformance
	PictureTypeScreenCapture
	PictureTypeBrightColouredFish
	PictureTypeIllustration
	PictureTypeBandArtistLogotype
	PictureTypePublisherStudioLogotype
)

// MetadataBlockPicture represents a picture metadata block
type MetadataBlockPicture struct {
	PictureType       PictureType
	MIME              string
	Description       string
	Width             uint32
	Height            uint32
	ColorDepth        uint32
	IndexedColorCount uint32
	ImageData         []byte
}
