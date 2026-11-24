package flacmeta

import (
	"bytes"
	"errors"
	"fmt"
	"image"
	"image/jpeg"
	"io"
	"os"
	"reflect"
	"sort"
	"strconv"
	"strings"
)

type FLACTag struct {
	Album        string
	AlbumArtist  string
	Artist       string
	BPM          int
	Contact      string
	Composer     string
	Copyright    string
	CoverArt     *image.Image
	Date         string
	Description  string
	DiscNumber   int
	DiscTotal    int
	Encoder      string
	Genre        string
	ISRC         string
	License      string
	Location     string
	Organization string
	Performer    string
	Title        string
	TrackNumber  int
	TrackTotal   int
	Version      string

	flacFile        *FLACFile
	firstFrame      int64
	commentVendor   string
	comments        []string
	primaryPicture  int
	initialCoverArt *image.Image
	coverArtEdited  bool
}

func (f *FLACTag) GetAlbum() string {
	return f.Album
}

func (f *FLACTag) GetAlbumArtist() string {
	return f.AlbumArtist
}

func (f *FLACTag) GetArtist() string {
	return f.Artist
}

func (f *FLACTag) GetBPM() int {
	return f.BPM
}

func (f *FLACTag) GetContact() string {
	return f.Contact
}

func (f *FLACTag) GetComposer() string {
	return f.Composer
}

func (f *FLACTag) GetCopyright() string {
	return f.Copyright
}

func (f *FLACTag) GetCoverArt() *image.Image {
	return f.CoverArt
}

func (f *FLACTag) GetDate() string {
	return f.Date
}

func (f *FLACTag) GetDescription() string {
	return f.Description
}

func (f *FLACTag) GetDiscNumber() int {
	return f.DiscNumber
}

func (f *FLACTag) GetDiscTotal() int {
	return f.DiscTotal
}

func (f *FLACTag) GetEncoder() string {
	return f.Encoder
}

func (f *FLACTag) GetGenre() string {
	return f.Genre
}

func (f *FLACTag) GetISRC() string {
	return f.ISRC
}

func (f *FLACTag) GetLicense() string {
	return f.License
}

func (f *FLACTag) GetLocation() string {
	return f.Location
}

func (f *FLACTag) GetOrganization() string {
	return f.Organization
}

func (f *FLACTag) GetPerformer() string {
	return f.Performer
}

func (f *FLACTag) GetTitle() string {
	return f.Title
}

func (f *FLACTag) GetTrackNumber() int {
	return f.TrackNumber
}

func (f *FLACTag) GetTrackTotal() int {
	return f.TrackTotal
}

func (f *FLACTag) GetVersion() string {
	return f.Version
}

func (f *FLACTag) SetAlbum(album string) {
	f.Album = album
}

func (f *FLACTag) SetAlbumArtist(albumArtist string) {
	f.AlbumArtist = albumArtist
}

func (f *FLACTag) SetArtist(artist string) {
	f.Artist = artist
}

func (f *FLACTag) SetBPM(bpm int) {
	f.BPM = bpm
}

func (f *FLACTag) SetContact(contact string) {
	f.Contact = contact
}

func (f *FLACTag) SetComposer(composer string) {
	f.Composer = composer
}

func (f *FLACTag) SetCopyright(copyright string) {
	f.Copyright = copyright
}

func (f *FLACTag) SetCoverArt(coverArt *image.Image) {
	f.CoverArt = coverArt
	f.coverArtEdited = true
}

func (f *FLACTag) SetDate(date string) {
	f.Date = date
}

func (f *FLACTag) SetDescription(description string) {
	f.Description = description
}

func (f *FLACTag) SetDiscNumber(discNumber int) {
	f.DiscNumber = discNumber
}

func (f *FLACTag) SetDiscTotal(discTotal int) {
	f.DiscTotal = discTotal
}

func (f *FLACTag) SetEncoder(encoder string) {
	f.Encoder = encoder
}

func (f *FLACTag) SetGenre(genre string) {
	f.Genre = genre
}

func (f *FLACTag) SetISRC(isrc string) {
	f.ISRC = isrc
}

func (f *FLACTag) SetLicense(license string) {
	f.License = license
}

func (f *FLACTag) SetLocation(location string) {
	f.Location = location
}

func (f *FLACTag) SetOrganization(organization string) {
	f.Organization = organization
}

func (f *FLACTag) SetPerformer(performer string) {
	f.Performer = performer
}

func (f *FLACTag) SetTitle(title string) {
	f.Title = title
}

func (f *FLACTag) SetTrackNumber(trackNumber int) {
	f.TrackNumber = trackNumber
}

func (f *FLACTag) SetTrackTotal(trackTotal int) {
	f.TrackTotal = trackTotal
}

func (f *FLACTag) SetVersion(version string) {
	f.Version = version
}

// ReadFLAC reads metadata, leaving the caller-owned source open for any subsequent Save calls.
func ReadFLAC(r io.ReadSeeker) (*FLACTag, error) {
	if r == nil {
		return nil, errors.New("flacmeta: nil source")
	}
	file, err := ParseMetaGetStreamReader(r)
	if err != nil {
		return nil, err
	}
	firstFrame, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, fmt.Errorf("flacmeta: record frame offset: %w", err)
	}
	tag := &FLACTag{flacFile: file, firstFrame: firstFrame, commentVendor: NewVorbisCommentBlock().Vendor, primaryPicture: -1}
	fields := reflect.ValueOf(tag).Elem()
	frontFound := false
	var primaryPicture *MetadataBlockPicture
	for i, block := range file.Meta {
		switch block.Type {
		case VorbisComment:
			vendor, entries, err := parseVorbisEntries(*block)
			if err != nil {
				return nil, err
			}
			tag.commentVendor = vendor
			tag.comments = append(tag.comments, entries...)
			for _, entry := range entries {
				key, value, _ := strings.Cut(entry, "=")
				if name, ok := tagFieldMapping[strings.ToUpper(key)]; ok {
					field := fields.FieldByName(name)
					if field.Kind() == reflect.Int {
						if n, e := strconv.Atoi(value); e == nil {
							field.SetInt(int64(n))
						}
					} else {
						field.SetString(value)
					}
				}
			}
		case Picture:
			picture, err := ParsePicFromMetaDataBlock(*block)
			if err != nil {
				return nil, err
			}
			// A linked picture is not a decoded image, but remains in the original metadata.
			if picture.MIME == MIMEURL {
				continue
			}
			if picture.PictureType != PictureTypeFrontCover && frontFound {
				continue
			}
			tag.primaryPicture = i
			primaryPicture = picture
			frontFound = picture.PictureType == PictureTypeFrontCover
		}
	}
	if primaryPicture != nil {
		img, _, err := image.Decode(bytes.NewReader(primaryPicture.ImageData))
		if err != nil {
			return nil, err
		}
		tag.CoverArt = &img
		tag.initialCoverArt = tag.CoverArt
	}
	return tag, nil
}

func (f *FLACTag) currentComment(key string) string {
	value := reflect.ValueOf(f).Elem().FieldByName(tagFieldMapping[key])
	if value.Kind() == reflect.Int {
		if value.Int() == 0 {
			return ""
		}
		return strconv.FormatInt(value.Int(), 10)
	}
	return value.String()
}

func (f *FLACTag) commentBlock() (MetaDataBlock, error) {
	effective := make(map[string]string)
	for _, entry := range f.comments {
		key, value, _ := strings.Cut(entry, "=")
		key = strings.ToUpper(key)
		if name, ok := tagFieldMapping[key]; ok {
			if reflect.ValueOf(f).Elem().FieldByName(name).Kind() == reflect.Int {
				if n, err := strconv.Atoi(value); err != nil {
					continue
				} else if n == 0 {
					value = ""
				} else {
					value = strconv.Itoa(n)
				}
			}
			effective[key] = value
		}
	}
	entries := make([]string, 0, len(f.comments)+len(tagFieldMapping))
	for _, entry := range f.comments {
		key, _, _ := strings.Cut(entry, "=")
		key = strings.ToUpper(key)
		if _, managed := tagFieldMapping[key]; !managed || f.currentComment(key) == effective[key] {
			entries = append(entries, entry)
		}
	}
	keys := make([]string, 0, len(tagFieldMapping))
	for key := range tagFieldMapping {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	for _, key := range keys {
		if value := f.currentComment(key); value != "" && value != effective[key] {
			entries = append(entries, key+"="+value)
		}
	}
	return marshalVorbisEntries(f.commentVendor, entries)
}

// Save writes updated metadata and streams the unchanged audio frames from the
// original reader. The source must remain open and must not be the destination.
// Neither stream is closed. The destination may contain partial output on error;
// to replace a file, write to a separate temporary output and rename on success.
func (f *FLACTag) Save(w io.Writer) error {
	if f == nil || f.flacFile == nil || f.flacFile.StreamReader == nil {
		return errors.New("flacmeta: Save requires a tag from ReadFLAC")
	}
	if w == nil {
		return errors.New("flacmeta: nil destination")
	}
	source := f.flacFile.StreamReader
	if src, ok := source.(*os.File); ok {
		if dst, ok := w.(*os.File); ok {
			a, err := src.Stat()
			if err != nil {
				return err
			}
			b, err := dst.Stat()
			if err != nil {
				return err
			}
			if os.SameFile(a, b) {
				return errors.New("flacmeta: source and destination are the same file")
			}
		}
	}
	changesArt := f.coverArtEdited || f.CoverArt != f.initialCoverArt
	var replacement *MetaDataBlock
	if changesArt && f.CoverArt != nil {
		if *f.CoverArt == nil {
			return errors.New("flacmeta: nil cover image")
		}
		var buf bytes.Buffer
		if err := jpeg.Encode(&buf, *f.CoverArt, nil); err != nil {
			return err
		}
		bounds := (*f.CoverArt).Bounds()
		picture := &MetadataBlockPicture{ImageData: buf.Bytes(), MIME: "image/jpeg", PictureType: PictureTypeFrontCover, Description: "Cover Art", Width: uint32(bounds.Dx()), Height: uint32(bounds.Dy()), ColorDepth: 24}
		b, err := picture.Marshal()
		if err != nil {
			return err
		}
		replacement = &b
	}
	comment, err := f.commentBlock()
	if err != nil {
		return err
	}
	blocks := make([]*MetaDataBlock, 0, len(f.flacFile.Meta)+2)
	wroteComment := false
	for i, b := range f.flacFile.Meta {
		if b.Type == VorbisComment {
			if !wroteComment {
				blocks = append(blocks, &comment)
				wroteComment = true
			}
		} else if i == f.primaryPicture && changesArt {
			if replacement != nil {
				blocks = append(blocks, replacement)
				replacement = nil
			}
		} else {
			blocks = append(blocks, b)
		}
	}
	if !wroteComment {
		blocks = append(blocks, &comment)
	}
	if replacement != nil {
		blocks = append(blocks, replacement)
	}
	if _, err := source.Seek(f.firstFrame, io.SeekStart); err != nil {
		return fmt.Errorf("flacmeta: seek to frames: %w", err)
	}
	if err := checkFLACStream(source); err != nil {
		return fmt.Errorf("flacmeta: inspect frames: %w", err)
	}
	if err := writeMetadata(w, blocks); err != nil {
		return fmt.Errorf("flacmeta: write metadata: %w", err)
	}
	// Hide optional WriteTo/ReadFrom hooks so the frame copy is bounded by our buffer.
	if _, err := io.CopyBuffer(struct{ io.Writer }{w}, struct{ io.Reader }{source}, make([]byte, 32*1024)); err != nil {
		return fmt.Errorf("flacmeta: copy frames: %w", err)
	}
	return nil
}
