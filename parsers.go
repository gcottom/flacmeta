package flacmeta

import (
	"bytes"
	"io"
	"strings"
)

// ParseMetadata accepts a reader to a FLAC stream and consumes only FLAC metadata
// Frames is always nil
func ParseMetadata(f io.ReadSeeker) (*FLACFile, error) {
	res := new(FLACFile)

	if err := readFLACHead(f); err != nil {
		return nil, err
	}
	meta, err := readMetadataBlocks(f)
	if err != nil {
		return nil, err
	}

	res.Meta = meta

	return res, nil
}

// ParseBytes accepts a reader to a FLAC stream and returns the final file
func ParseBytes(f io.ReadSeeker) (*FLACFile, error) {
	res, err := ParseMetadata(f)
	if err != nil {
		return nil, err
	}

	res.Frames, err = readFLACStream(f)
	if err != nil {
		return nil, err
	}

	return res, nil
}

func ParseMetaGetStreamReader(f io.ReadSeeker) (*FLACFile, error) {
	res, err := ParseMetadata(f)
	if err != nil {
		return nil, err
	}

	if err := checkFLACStream(f); err != nil {
		return nil, err
	}
	res.StreamReader = f
	return res, nil
}

func ParseVorbisCommentFromMetaDataBlock(meta MetaDataBlock) (*VorbisCommentBlock, error) {
	vendor, entries, err := parseVorbisEntries(meta)
	if err != nil {
		return nil, err
	}
	res := &VorbisCommentBlock{Vendor: vendor, Comments: make(map[string]string, len(entries))}
	for _, entry := range entries {
		key, value, _ := strings.Cut(entry, "=")
		res.Comments[key] = value
	}
	return res, nil
}

func parseVorbisEntries(meta MetaDataBlock) (string, []string, error) {
	if meta.Type != VorbisComment {
		return "", nil, ErrorNotVorbisComment
	}
	reader := bytes.NewBuffer(meta.Data)
	readField := func() (string, error) {
		n, err := readUint32L(reader)
		if err != nil {
			return "", err
		}
		if uint64(n) > uint64(reader.Len()) {
			return "", ErrorUnexpEof
		}
		return string(reader.Next(int(n))), nil
	}
	vendor, err := readField()
	if err != nil {
		return "", nil, err
	}
	count, err := readUint32L(reader)
	if err != nil {
		return "", nil, err
	}
	if uint64(count) > uint64(reader.Len()/4) {
		return "", nil, ErrorUnexpEof
	}
	entries := make([]string, 0, int(count))
	for i := uint32(0); i < count; i++ {
		item, err := readField()
		if err != nil {
			return "", nil, err
		}
		key, _, ok := strings.Cut(item, "=")
		if !ok || key == "" {
			return "", nil, ErrorMalformedComment
		}
		entries = append(entries, item)
	}
	return vendor, entries, nil
}
