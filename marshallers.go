package flacmeta

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

func (c *FLACFile) Marshal() ([]byte, error) {
	res := new(bytes.Buffer)
	if _, err := res.Write([]byte("fLaC")); err != nil {
		return nil, err
	}
	for i, meta := range c.Meta {
		last := i == len(c.Meta)-1
		data, err := meta.Marshal(last)
		if err != nil {
			return nil, err
		}
		if _, err := res.Write(data); err != nil {
			return nil, err
		}
	}
	if _, err := res.Write(c.Frames); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (c *FLACFile) MarshalMeta() ([]byte, error) {
	res := new(bytes.Buffer)
	if _, err := res.Write([]byte("fLaC")); err != nil {
		return nil, err
	}
	for i, meta := range c.Meta {
		last := i == len(c.Meta)-1
		data, err := meta.Marshal(last)
		if err != nil {
			return nil, err
		}
		if _, err := res.Write(data); err != nil {
			return nil, err
		}
	}
	return res.Bytes(), nil
}

func (c *MetaDataBlock) Marshal(isfinal bool) ([]byte, error) {
	res := new(bytes.Buffer)
	if isfinal {
		if err := res.WriteByte(byte(c.Type + 1<<7)); err != nil {
			return nil, err
		}
	} else {
		if err := res.WriteByte(byte(c.Type)); err != nil {
			return nil, err
		}
	}
	size, err := encodeUint32(uint32(len(c.Data)))
	if err != nil {
		return nil, err
	}
	if _, err := res.Write(size[len(size)-3:]); err != nil {
		return nil, err
	}
	if _, err := res.Write(c.Data); err != nil {
		return nil, err
	}
	return res.Bytes(), nil
}

func (v *VorbisCommentBlock) Marshal() (MetaDataBlock, error) {
	vendor, err := encodeComment(v.Vendor)
	if err != nil {
		return MetaDataBlock{}, err
	}
	buf := make([]byte, 4)
	binary.LittleEndian.PutUint32(buf, uint32(len(v.Comments)))
	commentPack := append(vendor, buf...)
	for key, value := range v.Comments {
		comment := fmt.Sprintf("%s=%s", key, value)
		enc, err := encodeComment(comment)
		if err != nil {
			return MetaDataBlock{}, err
		}
		commentPack = append(commentPack, enc...)
	}
	return MetaDataBlock{
		Type: VorbisComment,
		Data: commentPack,
	}, nil
}

func (c *MetadataBlockPicture) Marshal() (MetaDataBlock, error) {
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
