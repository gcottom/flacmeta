package flacmeta

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"reflect"
	"strconv"
	"strings"

	"github.com/aler9/writerseeker"
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

	flacFile *FLACFile
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

func ReadFLAC(r io.ReadSeeker) (*FLACTag, error) {
	ffile, err := ParseMetaGetStreamReader(r)
	if err != nil {
		return nil, err
	}
	commentBlockIdx := -1
	pictureBlockIdx := -1

	for i, v := range ffile.Meta {
		if v.Type == VorbisComment {
			commentBlockIdx = i
		}
		if v.Type == Picture {
			pictureBlockIdx = i
		}
	}
	tag := new(FLACTag)
	if commentBlockIdx != -1 {
		vorbisComment, err := ParseVorbisCommentFromMetaDataBlock(*ffile.Meta[commentBlockIdx])
		if err != nil {
			return nil, err
		}
		for k, v := range vorbisComment.Comments {
			fieldName := strings.ToUpper(k)
			fieldName, ok := tagFieldMapping[fieldName]
			if !ok {
				continue
			}
			switch fieldName {
			case structFieldBPM, structFieldDiscNumber, structFieldDiscTotal, structFieldTrackNumber, structFieldTrackTotal:
				val, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				if val != 0 {
					reflect.ValueOf(tag).Elem().FieldByName(fieldName).SetInt(int64(val))
				}
			default:
				if v != "" {
					reflect.ValueOf(tag).Elem().FieldByName(fieldName).SetString(v)
				}
			}

		}
	}
	if pictureBlockIdx != -1 {
		picture, err := ParsePicFromMetaDataBlock(*ffile.Meta[pictureBlockIdx])
		if err != nil {
			return nil, err
		}
		img, _, err := image.Decode(bytes.NewReader(picture.ImageData))
		if err != nil {
			return nil, err
		}
		tag.CoverArt = &img
	}
	tag.flacFile = ffile
	return tag, nil
}

func (f *FLACTag) Save(w io.Writer) error {
	blocks := make([]*MetaDataBlock, 0)
	comment := NewVorbisCommentBlock()
	for k, v := range tagFieldMapping {
		field := reflect.ValueOf(f).Elem().FieldByName(v)
		if field.Kind() == reflect.Int {
			comment.Comments[k] = strconv.Itoa(int(field.Int()))
		} else {
			comment.Comments[k] = field.String()
		}
	}
	cb, err := comment.Marshal()
	if err != nil {
		return err
	}
	blocks = append(blocks, &cb)

	if f.CoverArt != nil {
		pic := new(MetadataBlockPicture)
		buf := new(bytes.Buffer)
		err := jpeg.Encode(buf, *f.CoverArt, nil)
		if err != nil {
			return err
		}
		pic.ImageData = buf.Bytes()
		pic.MIME = "image/jpeg"
		pic.PictureType = PictureTypeFrontCover
		pic.Description = "Cover Art"
		pb, err := pic.Marshal()
		if err != nil {
			return err
		}
		blocks = append(blocks, &pb)
	}

	// remove existing vorbis comment and picture blocks
	for {
		found := -1
	inner:
		for idx, v := range f.flacFile.Meta {
			if v.Type == Picture || v.Type == VorbisComment {
				found = idx
				break inner
			}
		}
		if found == -1 {
			break
		} else {
			f.flacFile.Meta = append(f.flacFile.Meta[:found], f.flacFile.Meta[found+1:]...)
		}
	}

	f.flacFile.Meta = append(f.flacFile.Meta, blocks...)
	out, err := f.flacFile.MarshalMeta()
	if err != nil {
		return err
	}
	temp := new(writerseeker.WriterSeeker)
	if _, err = temp.Write(out); err != nil {
		return err
	}
	if _, err = io.Copy(temp, f.flacFile.StreamReader); err != nil {
		return err
	}
	if _, err = temp.Seek(0, io.SeekStart); err != nil {
		return err
	}
	if _, err = io.Copy(w, temp.BytesReader()); err != nil {
		return err
	}
	return nil
}
