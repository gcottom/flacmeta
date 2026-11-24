package flacmeta

import (
	"encoding/binary"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/gcottom/audiometa-core"
)

// OpenMetadata reads the shared metadata model without decoding artwork or
// buffering compressed frames. Leave r open and unchanged until all saves finish.
func OpenMetadata(r io.ReadSeeker) (*core.Editor, error) {
	if r == nil {
		return nil, errors.New("flacmeta: nil source")
	}
	file, err := ParseMetaGetStreamReader(r)
	if err != nil {
		return nil, err
	}
	firstFrame, err := r.Seek(0, io.SeekCurrent)
	if err != nil {
		return nil, err
	}
	if len(file.Meta) == 0 || file.Meta[0].Type != StreamInfo || len(file.Meta[0].Data) != 34 {
		return nil, ErrorNoStreamInfo
	}

	vendor := "AudioMeta"
	var comments []string
	var commentHasTrailer bool
	var pictures, fallback []core.Value
	commentIndex := -1
	pictureIndexes := make(map[int]bool)
	for i, block := range file.Meta {
		switch block.Type {
		case VorbisComment:
			if commentIndex >= 0 {
				return nil, fmt.Errorf("%w: duplicate FLAC comment block", core.ErrUnsupported)
			}
			vendor, comments, err = parseVorbisEntries(*block)
			if err != nil {
				return nil, err
			}
			commentIndex = i
			size := 8 + len(vendor)
			for _, entry := range comments {
				size += 4 + len(entry)
			}
			commentHasTrailer = size != len(block.Data)
		case Picture:
			if p, err := core.ParsePictureBlock(block.Data); err == nil {
				value := core.PictureValue(p)
				if core.Validate(core.Artwork, value) == nil {
					pictures = append(pictures, value)
					pictureIndexes[i] = true
				}
			}
		}
	}
	for _, entry := range comments {
		k, value, ok := strings.Cut(entry, "=")
		if ok && strings.EqualFold(k, core.CommentFallbackName(core.Artwork)) {
			if v, err := core.Decode(core.Artwork, value); err == nil {
				fallback = append(fallback, v)
			}
		}
	}
	fields, custom := core.ReadComments(core.FLAC, comments)
	if len(pictures) > 0 {
		fields[core.Artwork] = pictures
	} else if len(fallback) > 0 {
		fields[core.Artwork] = fallback
	}
	info := binary.BigEndian.Uint64(file.Meta[0].Data[10:18])
	rate, samples := int64(info>>44), int64(info&((uint64(1)<<36)-1))
	if rate > 0 && samples > 0 {
		fields[core.AudioDuration] = []core.Value{core.IntValue(samples * 1000 / rate)}
	}

	return core.NewEditor(fields, custom, func(w io.Writer, changes core.Changes) error {
		var replacements []*MetaDataBlock
		art, editArt := changes.Fields[core.Artwork]
		for _, value := range art {
			p, ok := value.Picture()
			if !ok {
				return core.ErrInvalidValue
			}
			if uint64(len(p.Data))+uint64(len(p.MIME))+uint64(len(p.Description))+32 > maxMetadataBlockSize {
				return fmt.Errorf("%w: FLAC picture exceeds 24-bit block size", core.ErrUnsupported)
			}
			raw, err := core.MarshalPictureBlock(p)
			if err != nil {
				return err
			}
			replacements = append(replacements, &MetaDataBlock{Type: Picture, Data: raw})
		}
		var comment *MetaDataBlock
		if len(changes.Custom) > 0 || len(changes.Fields) > 0 || changes.ClearUnknown {
			result, err := core.EditComments(core.FLAC, comments, changes)
			if err != nil {
				return err
			}
			if editArt && len(art) == 0 {
				kept := result[:0]
				for _, entry := range result {
					k, value, ok := strings.Cut(entry, "=")
					if ok && strings.EqualFold(k, core.CommentFallbackName(core.Artwork)) {
						if _, err := core.Decode(core.Artwork, value); err == nil {
							continue
						}
					}
					kept = append(kept, entry)
				}
				result = kept
			}
			if !equalComments(comments, result) {
				if commentHasTrailer {
					return fmt.Errorf("%w: trailing bytes in FLAC comment block", core.ErrUnsupported)
				}
				commentBlock, err := marshalVorbisEntries(vendor, result)
				if err != nil {
					return err
				}
				comment = &commentBlock
			}
		}

		blocks := make([]*MetaDataBlock, 0, len(file.Meta)+len(replacements)+1)
		for i, block := range file.Meta {
			switch {
			case i == commentIndex && comment != nil:
				blocks = append(blocks, comment)
				comment = nil
			case block.Type == Picture && pictureIndexes[i] && editArt:
				blocks = append(blocks, replacements...)
				replacements = nil
			case block.Type == Picture && changes.ClearUnknown && !pictureIndexes[i]:
				// This opaque block is known to be descriptive artwork.
			default:
				blocks = append(blocks, block)
			}
		}
		if comment != nil {
			blocks = append(blocks, comment)
		}
		blocks = append(blocks, replacements...)
		return saveFrames(r, w, firstFrame, blocks)
	}), nil
}

func equalComments(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i, entry := range a {
		if entry != b[i] {
			return false
		}
	}
	return true
}
