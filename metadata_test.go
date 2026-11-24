package flacmeta

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/gcottom/audiometa-core/v2"
	"github.com/gcottom/audiometa-core/v2/contracttest"
)

func TestCanonicalContract(t *testing.T) {
	src, err := os.ReadFile("testdata/testdata-nonEmpty.flac")
	if err != nil {
		t.Fatal(err)
	}
	// Keep original metadata and one well-formed frame signature. The
	// integration suite separately exercises full, independently decoded audio.
	r := bytes.NewReader(src)
	if _, err := ParseMetadata(r); err != nil {
		t.Fatal(err)
	}
	offset, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		t.Fatal(err)
	}
	contracttest.Run(t, OpenMetadata, src[:offset+2])
}

func TestCanonicalFLACPreservesOpaqueBlocksAndArtwork(t *testing.T) {
	raw, err := os.ReadFile("testdata/testdata-nonEmpty.flac")
	if err != nil {
		t.Fatal(err)
	}
	r := bytes.NewReader(raw)
	orig, err := ParseMetadata(r)
	if err != nil {
		t.Fatal(err)
	}
	first, _ := r.Seek(0, io.SeekCurrent)
	opaque := &MetaDataBlock{Type: Application, Data: []byte("TESTnever edit")}
	reserved := &MetaDataBlock{Type: Reserved, Data: []byte{1, 2, 3}}
	orig.Meta = append(orig.Meta, opaque, reserved)
	source, err := orig.MarshalMeta()
	if err != nil {
		t.Fatal(err)
	}
	source = append(source, raw[first:]...)
	tag := contracttest.OpenFixture(t, OpenMetadata, source)
	if v, present := tag.GetField(core.AudioDuration); !present || v.Kind() != core.KindInt {
		t.Fatal("STREAMINFO duration missing")
	}
	if err := tag.SetField(core.Title, core.TextValue("")); err != nil {
		t.Fatal(err)
	}
	out := contracttest.SaveFixture(t, tag)
	got, err := ParseMetadata(bytes.NewReader(out))
	if err != nil {
		t.Fatal(err)
	}
	var pictures, actual int
	for _, block := range orig.Meta {
		if block.Type == Picture {
			pictures++
		}
	}
	for _, block := range got.Meta {
		switch block.Type {
		case Application:
			if !bytes.Equal(block.Data, opaque.Data) {
				t.Fatal("application block changed")
			}
		case Reserved:
			if !bytes.Equal(block.Data, reserved.Data) {
				t.Fatal("reserved block changed")
			}
		case Picture:
			matched := false
			for _, initial := range orig.Meta {
				if initial.Type == Picture && bytes.Equal(block.Data, initial.Data) {
					matched = true
					break
				}
			}
			if !matched {
				t.Fatal("original picture encoding changed")
			}
			actual++
		}
	}
	if pictures != actual {
		t.Fatal("picture count changed")
	}
}

func BenchmarkCanonicalFLACSave(b *testing.B) {
	for _, tc := range []struct {
		name       string
		audio, art int64
	}{
		{"small", 16 << 10, 0},
		{"large_lossless_transport", 64 << 20, 0},
		{"artwork_payload", 16 << 10, 4 << 20},
	} {
		b.Run(tc.name, func(b *testing.B) {
			blocks := []*MetaDataBlock{{Type: StreamInfo, Data: make([]byte, 34)}}
			if tc.art > 0 {
				picture, err := core.MarshalPictureBlock(core.Picture{Role: 3, MIME: "image/jpeg", Data: make([]byte, tc.art)})
				if err != nil {
					b.Fatal(err)
				}
				blocks = append(blocks, &MetaDataBlock{Type: Picture, Data: picture})
			}
			header, err := (&FLACFile{Meta: blocks}).MarshalMeta()
			if err != nil {
				b.Fatal(err)
			}
			source := &generatedFLAC{header: header, total: int64(len(header)) + tc.audio}
			tag, err := OpenMetadata(source)
			if err != nil {
				b.Fatal(err)
			}
			if err := tag.SetField(core.Title, core.TextValue("changed")); err != nil {
				b.Fatal(err)
			}
			b.SetBytes(tc.audio + tc.art)
			b.ReportAllocs()
			b.ResetTimer()
			for i := 0; i < b.N; i++ {
				if err := tag.Save(io.Discard); err != nil {
					b.Fatal(err)
				}
			}
		})
	}
}

func FuzzVorbisBlock(f *testing.F) {
	seed, err := marshalVorbisEntries("encoder", []string{"TITLE=音", "Unknown=x", "Unknown=y"})
	if err != nil {
		f.Fatal(err)
	}
	f.Add([]byte(seed.Data))
	f.Add([]byte{})
	f.Fuzz(func(t *testing.T, data []byte) {
		if len(data) > 1<<20 {
			return
		}
		vendor, entries, err := parseVorbisEntries(MetaDataBlock{Type: VorbisComment, Data: data})
		if err != nil {
			return
		}
		rewritten, err := marshalVorbisEntries(vendor, entries)
		if err != nil {
			t.Fatal(err)
		}
		back, comments, err := parseVorbisEntries(rewritten)
		if err != nil || back != vendor || !equalComments(entries, comments) {
			t.Fatalf("Vorbis metadata projection changed: %v", err)
		}
	})
}
