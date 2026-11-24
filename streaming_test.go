package flacmeta

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"errors"
	"image"
	"image/jpeg"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"
)

type generatedFLAC struct {
	header                 []byte
	total, pos, frameReads int64
	maxFrameRead           int
	frameError             error
}

func (r *generatedFLAC) Read(p []byte) (int, error) {
	if r.pos >= r.total {
		return 0, io.EOF
	}
	if r.pos < int64(len(r.header)) {
		n := copy(p, r.header[r.pos:])
		r.pos += int64(n)
		return n, nil
	}
	if r.frameError != nil && r.pos >= int64(len(r.header))+32*1024 {
		return 0, r.frameError
	}
	off := r.pos - int64(len(r.header))
	if len(p) > r.maxFrameRead {
		r.maxFrameRead = len(p)
	}
	n := len(p)
	if int64(n) > r.total-r.pos {
		n = int(r.total - r.pos)
	}
	for i := 0; i < n; i++ {
		switch off + int64(i) {
		case 0:
			p[i] = 0xff
		case 1:
			p[i] = 0xf8
		default:
			p[i] = byte((off + int64(i)) % 239)
		}
	}
	r.pos += int64(n)
	r.frameReads += int64(n)
	return n, nil
}
func (r *generatedFLAC) Seek(offset int64, whence int) (int64, error) {
	switch whence {
	case io.SeekCurrent:
		offset += r.pos
	case io.SeekEnd:
		offset += r.total
	case io.SeekStart:
	default:
		return 0, errors.New("invalid seek")
	}
	if offset < 0 {
		return 0, errors.New("invalid position")
	}
	r.pos = offset
	return offset, nil
}

// A reader's optional fast path must not substitute an unbounded implementation.
func (*generatedFLAC) WriteTo(io.Writer) (int64, error) { panic("unexpected WriteTo") }

type incrementalOutput struct {
	t            *testing.T
	source       *generatedFLAC
	maxLookahead int64
	pos          int64
	headerEnd    int
	header       []byte
	frameHash    io.Writer
}

func (w *incrementalOutput) Write(p []byte) (int, error) {
	if w.source.frameReads-w.pos > w.maxLookahead {
		w.t.Fatal("source is being buffered before writing")
	}
	if w.pos < int64(w.headerEnd) {
		n := w.headerEnd - int(w.pos)
		if n > len(p) {
			n = len(p)
		}
		w.header = append(w.header, p[:n]...)
		if _, err := w.frameHash.Write(p[n:]); err != nil {
			return 0, err
		}
	} else if _, err := w.frameHash.Write(p); err != nil {
		return 0, err
	}
	w.pos += int64(len(p))
	return len(p), nil
}
func (*incrementalOutput) ReadFrom(io.Reader) (int64, error) { panic("unexpected ReadFrom") }
func minimalFLAC(t *testing.T, n int64) *generatedFLAC {
	t.Helper()
	f := FLACFile{Meta: []*MetaDataBlock{{Type: StreamInfo, Data: make([]byte, 34)}}}
	header, e := f.MarshalMeta()
	if e != nil {
		t.Fatal(e)
	}
	return &generatedFLAC{header: header, total: int64(len(header)) + n}
}
func TestSaveStreamsFramesWithBoundedMemory(t *testing.T) {
	const frameLen = 64 * 1024 * 1024
	r := minimalFLAC(t, frameLen)
	tag, e := ReadFLAC(r)
	if e != nil {
		t.Fatal(e)
	}
	tag.SetTitle("streamed")
	cb, e := tag.commentBlock()
	if e != nil {
		t.Fatal(e)
	}
	hdr, e := (&FLACFile{Meta: append(tag.flacFile.Meta, &cb)}).MarshalMeta()
	if e != nil {
		t.Fatal(e)
	}
	h := sha256.New()
	output := &incrementalOutput{t: t, source: r, maxLookahead: 64 * 1024, headerEnd: len(hdr), frameHash: h}
	var before, after runtime.MemStats
	runtime.ReadMemStats(&before)
	if e = tag.Save(output); e != nil {
		t.Fatal(e)
	}
	runtime.ReadMemStats(&after)
	allocated := after.TotalAlloc - before.TotalAlloc
	t.Logf("streamed %d audio bytes using %d bytes allocated; maximum read %d", frameLen, allocated, r.maxFrameRead)
	if allocated > 4*1024*1024 {
		t.Fatalf("Save allocated %d bytes for %d audio bytes", allocated, frameLen)
	}
	if !bytes.Equal(hdr, output.header) || output.pos != int64(len(hdr))+frameLen || r.maxFrameRead > 32*1024 {
		t.Fatalf("wrong output/read size: %d %d", output.pos, r.maxFrameRead)
	}
	sourceHash := sha256.New()
	if _, e = r.Seek(int64(len(r.header)), io.SeekStart); e != nil {
		t.Fatal(e)
	}
	if _, e = io.Copy(sourceHash, struct{ io.Reader }{r}); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(sourceHash.Sum(nil), h.Sum(nil)) {
		t.Fatal("audio bytes changed")
	}
}

type errWriter struct {
	err    error
	remain int
}

func (w *errWriter) Write(p []byte) (int, error) {
	if w.remain < len(p) {
		return 0, w.err
	}
	w.remain -= len(p)
	return len(p), nil
}

type shortWriter struct{}

func (shortWriter) Write(p []byte) (int, error) { return len(p) - 1, nil }
func TestSaveErrorsAndRepeatedOutput(t *testing.T) {
	data, e := os.ReadFile("testdata/testdata-nonEmpty.flac")
	if e != nil {
		t.Fatal(e)
	}
	reader := bytes.NewReader(data)
	tag, e := ReadFLAC(reader)
	if e != nil {
		t.Fatal(e)
	}
	before, e := ParseBytes(bytes.NewReader(data))
	if e != nil {
		t.Fatal(e)
	}
	dstErr := errors.New("injected output failure")
	if e = tag.Save(&errWriter{err: dstErr, remain: 0}); !errors.Is(e, dstErr) {
		t.Fatal("missing writer failure", e)
	}
	if e = tag.Save(shortWriter{}); !errors.Is(e, io.ErrShortWrite) {
		t.Fatal("missing short-write failure", e)
	}
	tag.SetTitle("after failure")
	var first, second bytes.Buffer
	if e = tag.Save(&first); e != nil {
		t.Fatal(e)
	}
	if _, e = reader.Seek(0, io.SeekStart); e != nil {
		t.Fatal(e)
	}
	if e = tag.Save(&second); e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("different repeated output")
	}
	output, e := ParseBytes(bytes.NewReader(first.Bytes()))
	if e != nil {
		t.Fatal(e)
	}
	if !bytes.Equal(before.Frames, output.Frames) {
		t.Fatal("audio changed")
	}
	tag2, e := ReadFLAC(bytes.NewReader(first.Bytes()))
	if e != nil || tag2.GetTitle() != "after failure" {
		t.Fatalf("bad rewritten title: %v, %v", tag2, e)
	}
	var unset FLACTag
	if e = unset.Save(io.Discard); e == nil {
		t.Fatal("zero value Save should fail")
	}
	r := minimalFLAC(t, 256*1024)
	broken, e := ReadFLAC(r)
	if e != nil {
		t.Fatal(e)
	}
	srcErr := errors.New("injected input failure")
	r.frameError = srcErr
	if e = broken.Save(io.Discard); !errors.Is(e, srcErr) {
		t.Fatal("missing source failure", e)
	}
	r.total = int64(len(r.header))
	var truncatedOutput bytes.Buffer
	if e = broken.Save(&truncatedOutput); !errors.Is(e, io.EOF) || truncatedOutput.Len() != 0 {
		t.Fatalf("truncated source produced %d bytes or wrong error: %v", truncatedOutput.Len(), e)
	}

}

func TestSavePreservesUnmanagedMetadata(t *testing.T) {
	var picBuf bytes.Buffer
	if e := jpeg.Encode(&picBuf, image.NewRGBA(image.Rect(0, 0, 3, 2)), nil); e != nil {
		t.Fatal(e)
	}
	front, e := (&MetadataBlockPicture{PictureType: PictureTypeFrontCover, MIME: "image/jpeg", ImageData: picBuf.Bytes()}).Marshal()
	if e != nil {
		t.Fatal(e)
	}
	back, e := (&MetadataBlockPicture{PictureType: PictureTypeBackCover, MIME: "image/jpeg", ImageData: picBuf.Bytes()}).Marshal()
	if e != nil {
		t.Fatal(e)
	}
	comment, e := marshalVorbisEntries("custom vendor", []string{"title=old", "artist=one", "artist=two", "X-CUSTOM=alpha=beta", "X-CUSTOM=gamma"})
	if e != nil {
		t.Fatal(e)
	}
	raw, e := (&FLACFile{Meta: []*MetaDataBlock{{Type: StreamInfo, Data: make([]byte, 34)}, &comment, &front, &back, {Type: Application, Data: []byte("ABCDopaque")}}, Frames: []byte{0xff, 0xf8, 1, 2, 3}}).Marshal()
	if e != nil {
		t.Fatal(e)
	}
	tag, e := ReadFLAC(bytes.NewReader(raw))
	if e != nil {
		t.Fatal(e)
	}
	tag.SetTitle("new")
	var out bytes.Buffer
	if e = tag.Save(&out); e != nil {
		t.Fatal(e)
	}
	f, e := ParseBytes(bytes.NewReader(out.Bytes()))
	if e != nil {
		t.Fatal(e)
	}
	if len(f.Meta) != 5 || !bytes.Equal(f.Meta[2].Data, front.Data) || !bytes.Equal(f.Meta[3].Data, back.Data) || string(f.Meta[4].Data) != "ABCDopaque" {
		t.Fatal("unmodified pictures or application block lost")
	}
	vendor, entries, e := parseVorbisEntries(*f.Meta[1])
	if e != nil || vendor != "custom vendor" || strings.Join(entries, ",") != "artist=one,artist=two,X-CUSTOM=alpha=beta,X-CUSTOM=gamma,TITLE=new" {
		t.Fatalf("lost metadata: %q %q %v", vendor, entries, e)
	}
	tag.SetCoverArt(nil)
	out.Reset()
	if e = tag.Save(&out); e != nil {
		t.Fatal(e)
	}
	f, e = ParseBytes(bytes.NewReader(out.Bytes()))
	if e != nil || len(f.Meta) != 4 || !bytes.Equal(f.Meta[2].Data, back.Data) {
		t.Fatal("removal should preserve back cover", e)
	}
}

func TestTruncationAndMetadataLengths(t *testing.T) {
	r := minimalFLAC(t, 0)
	for _, data := range [][]byte{r.header, append(append([]byte(nil), r.header...), 0xff)} {
		if _, e := ParseBytes(bytes.NewReader(data)); !errors.Is(e, ErrorNoSyncCode) {
			t.Fatalf("unexpected truncated-frame result: %v", e)
		}
	}
	huge := make([]byte, 4)
	binary.LittleEndian.PutUint32(huge, 0xffffffff)
	if _, e := ParseVorbisCommentFromMetaDataBlock(MetaDataBlock{Type: VorbisComment, Data: huge}); e == nil {
		t.Fatal("accepted corrupt Vorbis length")
	}
	pic := make([]byte, 8)
	binary.BigEndian.PutUint32(pic[4:], 0xffffffff)
	if _, e := ParsePicFromMetaDataBlock(MetaDataBlock{Type: Picture, Data: pic}); !errors.Is(e, io.ErrUnexpectedEOF) {
		t.Fatalf("incorrect picture error: %v", e)
	}
	b := MetaDataBlock{Type: Application, Data: make([]byte, maxMetadataBlockSize+1)}
	if _, e := b.Marshal(false); e == nil {
		t.Fatal("23/24-bit truncation went unnoticed")
	}
	tag, e := ReadFLAC(minimalFLAC(t, 2))
	if e != nil {
		t.Fatal(e)
	}
	tag.SetTitle(strings.Repeat("x", maxMetadataBlockSize))
	var out bytes.Buffer
	if e = tag.Save(&out); e == nil || out.Len() != 0 {
		t.Fatalf("invalid metadata wrote %d bytes; %v", out.Len(), e)
	}
}

func TestSaveRejectsSameFileWithoutWriting(t *testing.T) {
	data, err := os.ReadFile("testdata/testdata-nonEmpty.flac")
	if err != nil {
		t.Fatal(err)
	}
	path := t.TempDir() + "/source.flac"
	if err = os.WriteFile(path, data, 0600); err != nil {
		t.Fatal(err)
	}
	in, err := os.Open(path)
	if err != nil {
		t.Fatal(err)
	}
	defer in.Close()
	out, err := os.OpenFile(path, os.O_WRONLY, 0)
	if err != nil {
		t.Fatal(err)
	}
	defer out.Close()
	tag, err := ReadFLAC(in)
	if err != nil {
		t.Fatal(err)
	}
	if err = tag.Save(out); err == nil {
		t.Fatal("Save must reject a destination alias")
	}
	got, err := os.ReadFile(path)
	if err != nil || !bytes.Equal(data, got) {
		t.Fatal("source was changed", err)
	}
}
