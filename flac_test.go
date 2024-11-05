package flacmeta

import (
	"bytes"
	"image/jpeg"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReadFlac(t *testing.T) {
	// Open the file
	flacFile, err := os.Open("./testdata/testdata-nonEmpty.flac")
	assert.NoError(t, err)
	defer flacFile.Close()

	tag, err := ReadFLAC(flacFile)
	assert.NoError(t, err)
	assert.Equal(t, "TestTitle1", tag.GetTitle())
	assert.Equal(t, "TestArtist1", tag.GetArtist())
	assert.Equal(t, "TestAlbum1", tag.GetAlbum())
}

func TestReadWriteUpdate(t *testing.T) {
	// Open the file
	flacFile, err := os.Open("./testdata/testdata-nonEmpty.flac")
	assert.NoError(t, err)
	defer flacFile.Close()

	tag, err := ReadFLAC(flacFile)
	assert.NoError(t, err)
	assert.Equal(t, "TestTitle1", tag.GetTitle())
	assert.Equal(t, "TestArtist1", tag.GetArtist())
	assert.Equal(t, "TestAlbum1", tag.GetAlbum())
	tag.SetAlbum("TestAlbum2")
	tag.SetArtist("TestArtist2")
	tag.SetTitle("TestTitle2")
	tag.SetBPM(123)
	tag.SetComposer("TestComposer")
	tag.SetCopyright("TestCopyright")
	tag.SetDiscNumber(1)
	tag.SetDiscTotal(2)
	tag.SetEncoder("TestEncoder")
	tag.SetGenre("TestGenre")
	tag.SetTrackNumber(1)
	tag.SetTrackTotal(2)
	tag.SetAlbumArtist("TestAlbumArtist")
	tag.SetContact("TestContact")
	tag.SetDate("TestDate")
	tag.SetDescription("TestDescription")
	tag.SetISRC("TestISRC")
	tag.SetLicense("TestLicense")
	tag.SetLocation("TestLocation")
	tag.SetOrganization("TestOrganization")
	tag.SetPerformer("TestPerformer")
	tag.SetVersion("TestVersion")

	fimg, err := os.Open("./testdata/testdata-img-1.jpg")
	assert.NoError(t, err)
	defer fimg.Close()

	img, err := jpeg.Decode(fimg)
	assert.NoError(t, err)

	tag.SetCoverArt(&img)

	out := new(bytes.Buffer)
	err = tag.Save(out)
	assert.NoError(t, err)

	reader := bytes.NewReader(out.Bytes())
	tag2, err := ReadFLAC(reader)
	assert.NoError(t, err)
	assert.Equal(t, "TestTitle2", tag2.GetTitle())
	assert.Equal(t, "TestArtist2", tag2.GetArtist())
	assert.Equal(t, "TestAlbum2", tag2.GetAlbum())
	assert.Equal(t, 123, tag2.GetBPM())
	assert.Equal(t, "TestComposer", tag2.GetComposer())
	assert.Equal(t, "TestCopyright", tag2.GetCopyright())
	assert.Equal(t, 1, tag2.GetDiscNumber())
	assert.Equal(t, 2, tag2.GetDiscTotal())
	assert.Equal(t, "TestEncoder", tag2.GetEncoder())
	assert.Equal(t, "TestGenre", tag2.GetGenre())
	assert.Equal(t, 1, tag2.GetTrackNumber())
	assert.Equal(t, 2, tag2.GetTrackTotal())
	assert.Equal(t, "TestAlbumArtist", tag2.GetAlbumArtist())
	assert.Equal(t, "TestContact", tag2.GetContact())
	assert.Equal(t, "TestDate", tag2.GetDate())
	assert.Equal(t, "TestDescription", tag2.GetDescription())
	assert.Equal(t, "TestISRC", tag2.GetISRC())
	assert.Equal(t, "TestLicense", tag2.GetLicense())
	assert.Equal(t, "TestLocation", tag2.GetLocation())
	assert.Equal(t, "TestOrganization", tag2.GetOrganization())
	assert.Equal(t, "TestPerformer", tag2.GetPerformer())
	assert.Equal(t, "TestVersion", tag2.GetVersion())
	assert.NotNil(t, tag2.GetCoverArt())
}

func TestParseBytes(t *testing.T) {
	// Open the file
	flacFile, err := os.Open("./testdata/testdata-nonEmpty.flac")
	assert.NoError(t, err)
	defer flacFile.Close()

	ff, err := ParseBytes(flacFile)
	assert.NoError(t, err)
	assert.Greater(t, len(ff.Frames), 0)
	assert.Greater(t, len(ff.Meta), 0)
}
