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
	if meta.Type != VorbisComment {
		return nil, ErrorNotVorbisComment
	}

	reader := bytes.NewReader(meta.Data)
	res := new(VorbisCommentBlock)

	vendorlen, err := readUint32L(reader)
	if err != nil {
		return nil, err
	}
	vendorbytes := make([]byte, vendorlen)
	nn, err := reader.Read(vendorbytes)
	if err != nil {
		return nil, err
	}
	if nn != int(vendorlen) {
		return nil, ErrorUnexpEof
	}
	res.Vendor = string(vendorbytes)

	cmtcount, err := readUint32L(reader)
	if err != nil {
		return nil, err
	}
	res.Comments = make(map[string]string, 0)
	for i := 0; i < int(cmtcount); i++ {
		cmtlen, err := readUint32L(reader)
		if err != nil {
			return nil, err
		}
		cmtbytes := make([]byte, cmtlen)
		nn, err := reader.Read(cmtbytes)
		if err != nil {
			return nil, err
		}
		if nn != int(cmtlen) {
			return nil, ErrorUnexpEof
		}
		sp := strings.Split(string(cmtbytes), "=")
		res.Comments[sp[0]] = strings.Join(sp[1:], "=")
	}
	return res, nil
}
