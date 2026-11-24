package flacmeta

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"sort"
)

func (c *FLACFile) Marshal() ([]byte, error) {
	var res bytes.Buffer
	if err := writeMetadata(&res, c.Meta); err != nil {
		return nil, err
	}
	if err := writeBytes(&res, c.Frames); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (c *FLACFile) MarshalMeta() ([]byte, error) {
	var res bytes.Buffer
	if err := writeMetadata(&res, c.Meta); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

const maxMetadataBlockSize = 1<<24 - 1

func metadataHeader(c *MetaDataBlock, final bool) ([4]byte, error) {
	if c == nil {
		return [4]byte{}, fmt.Errorf("flacmeta: nil metadata block")
	}
	if c.Type < 0 || c.Type >= Invalid {
		return [4]byte{}, fmt.Errorf("flacmeta: invalid metadata block type %d", c.Type)
	}
	n := len(c.Data)
	if n > maxMetadataBlockSize {
		return [4]byte{}, fmt.Errorf("flacmeta: metadata block exceeds 24-bit length: %d bytes", n)
	}
	h := [4]byte{byte(c.Type), byte(n >> 16), byte(n >> 8), byte(n)}
	if final {
		h[0] |= 0x80
	}
	return h, nil
}

func writeBytes(w io.Writer, p []byte) error {
	if len(p) == 0 {
		return nil
	}
	n, err := w.Write(p)
	if err == nil && n != len(p) {
		return io.ErrShortWrite
	}
	return err
}

func writeMetadata(w io.Writer, blocks []*MetaDataBlock) error {
	if len(blocks) == 0 {
		return ErrorNoStreamInfo
	}
	for i, b := range blocks {
		if _, err := metadataHeader(b, i == len(blocks)-1); err != nil {
			return err
		}
	}
	if err := writeBytes(w, []byte("fLaC")); err != nil {
		return err
	}
	for i, b := range blocks {
		h, _ := metadataHeader(b, i == len(blocks)-1)
		if err := writeBytes(w, h[:]); err != nil {
			return err
		}
		if err := writeBytes(w, b.Data); err != nil {
			return err
		}
	}
	return nil
}

func (c *MetaDataBlock) Marshal(isfinal bool) ([]byte, error) {
	header, err := metadataHeader(c, isfinal)
	if err != nil {
		return nil, err
	}
	return append(header[:], c.Data...), nil
}

func (v *VorbisCommentBlock) Marshal() (MetaDataBlock, error) {
	keys := make([]string, 0, len(v.Comments))
	for key := range v.Comments {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	entries := make([]string, 0, len(keys))
	for _, key := range keys {
		entries = append(entries, key+"="+v.Comments[key])
	}
	return marshalVorbisEntries(v.Vendor, entries)
}

func marshalVorbisEntries(vendor string, entries []string) (MetaDataBlock, error) {
	size := uint64(8) + uint64(len(vendor))
	for _, entry := range entries {
		size += 4 + uint64(len(entry))
		if size > maxMetadataBlockSize {
			break
		}
	}
	if size > maxMetadataBlockSize {
		return MetaDataBlock{}, fmt.Errorf("flacmeta: Vorbis comment exceeds 24-bit block length")
	}
	data := make([]byte, int(size))
	off := 0
	put32 := func(n uint32) { binary.LittleEndian.PutUint32(data[off:off+4], n); off += 4 }
	putField := func(s string) { put32(uint32(len(s))); off += copy(data[off:], s) }
	putField(vendor)
	put32(uint32(len(entries)))
	for _, entry := range entries {
		putField(entry)
	}
	return MetaDataBlock{Type: VorbisComment, Data: data}, nil
}

func (c *MetadataBlockPicture) Marshal() (MetaDataBlock, error) {
	if uint64(32)+uint64(len(c.MIME))+uint64(len(c.Description))+uint64(len(c.ImageData)) > maxMetadataBlockSize {
		return MetaDataBlock{}, fmt.Errorf("flacmeta: picture exceeds 24-bit block length")
	}
	res := new(bytes.Buffer)
	enc, err := encodeUint32(uint32(c.PictureType))
	if err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(enc); err != nil {
		return MetaDataBlock{}, err
	}
	enc, err = encodeUint32(uint32(len(c.MIME)))
	if err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(enc); err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write([]byte(c.MIME)); err != nil {
		return MetaDataBlock{}, err
	}
	enc, err = encodeUint32(uint32(len(c.Description)))
	if err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(enc); err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write([]byte(c.Description)); err != nil {
		return MetaDataBlock{}, err
	}
	enc, err = encodeUint32(c.Width)
	if err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(enc); err != nil {
		return MetaDataBlock{}, err
	}
	enc, err = encodeUint32(c.Height)
	if err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(enc); err != nil {
		return MetaDataBlock{}, err
	}
	enc, err = encodeUint32(c.ColorDepth)
	if err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(enc); err != nil {
		return MetaDataBlock{}, err
	}
	enc, err = encodeUint32(c.IndexedColorCount)
	if err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(enc); err != nil {
		return MetaDataBlock{}, err
	}
	enc, err = encodeUint32(uint32(len(c.ImageData)))
	if err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(enc); err != nil {
		return MetaDataBlock{}, err
	}
	if _, err = res.Write(c.ImageData); err != nil {
		return MetaDataBlock{}, err
	}
	return MetaDataBlock{
		Type: Picture,
		Data: res.Bytes(),
	}, nil
}
