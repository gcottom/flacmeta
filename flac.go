package flacmeta

import (
	"bytes"
	"image"
	"image/jpeg"
	"io"
	"reflect"
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

func (f *FLACTag) GetComposer() string {
	return f.Composer
}

func (f *FLACTag) GetCopyright() string {
	return f.Copyright
}

func (f *FLACTag) GetCoverArt() *image.Image {
	return f.CoverArt
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

func (f *FLACTag) GetTitle() string {
	return f.Title
}

func (f *FLACTag) GetTrackNumber() int {
	return f.TrackNumber
}

func (f *FLACTag) GetTrackTotal() int {
	return f.TrackTotal
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

func (f *FLACTag) SetComposer(composer string) {
	f.Composer = composer
}

func (f *FLACTag) SetCopyright(copyright string) {
	f.Copyright = copyright
}

func (f *FLACTag) SetCoverArt(coverArt *image.Image) {
	f.CoverArt = coverArt
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

func (f *FLACTag) SetTitle(title string) {
	f.Title = title
}

func (f *FLACTag) SetTrackNumber(trackNumber int) {
	f.TrackNumber = trackNumber
}

func (f *FLACTag) SetTrackTotal(trackTotal int) {
	f.TrackTotal = trackTotal
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
			switch fieldName {
			case FIELD_BPM, FIELD_DISCNUMBER, FIELD_DISCTOTAL, FIELD_TRACKNUMBER, FIELD_TRACKTOTAL:
				val, err := strconv.Atoi(v)
				if err != nil {
					continue
				}
				reflect.ValueOf(tag).Elem().FieldByName(tagFieldMapping[strings.ToUpper(k)]).SetInt(int64(val))
			default:
				reflect.ValueOf(tag).Elem().FieldByName(tagFieldMapping[strings.ToUpper(k)]).SetString(v)
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
	comment := new(VorbisCommentBlock)
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
	if _, err = w.Write(out); err != nil {
		return err
	}
	if _, err = io.Copy(w, f.flacFile.StreamReader); err != nil {
		return err
	}
	return nil
}
