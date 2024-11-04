package flacmeta

import (
	"bytes"
	"io"
)

func (c *FLACFile) GetStreamInfo() (*StreamInfoBlock, error) {
	if c.Meta[0].Type != StreamInfo {
		return nil, ErrorNoStreamInfo
	}
	streamInfo := bytes.NewReader(c.Meta[0].Data)
	res := StreamInfoBlock{}

	if buf, err := readUint16(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.BlockSizeMin = int(buf)
	}

	if buf, err := readUint16(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.BlockSizeMax = int(buf)
	}

	buf := bytes.NewBuffer([]byte{0})
	buf24 := make([]byte, 3)
	if _, err := io.ReadFull(streamInfo, buf24); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}
	buf.Write(buf24)
	if buf, err := readUint32(buf); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.FrameSizeMin = int(buf)
	}
	buf.Reset()
	buf.WriteByte(0)
	if _, err := io.ReadFull(streamInfo, buf24); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}
	buf.Write(buf24)
	if buf, err := readUint32(buf); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.FrameSizeMax = int(buf)
	}

	buf.Reset()
	buf.WriteByte(0)
	smpl := make([]byte, 3)
	if _, err := io.ReadFull(streamInfo, smpl); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}
	buf.Write(smpl)
	if smplrate, err := readUint32(buf); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.SampleRate = int(smplrate >> 4)
	}
	if _, err := streamInfo.Seek(-1, io.SeekCurrent); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}

	if channel, err := readUint8(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.ChannelCount = int(channel<<4>>5) + 1
	}
	buf.Reset()
	if _, err := streamInfo.Seek(-1, io.SeekCurrent); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}

	if bitdepth, err := readUint16(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		res.BitDepth = int(bitdepth<<7>>11) + 1
	}
	if _, err := streamInfo.Seek(-1, io.SeekCurrent); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}

	var smplcount int64
	if count, err := readUint32(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		smplcount += int64(count<<4>>4) << 8
	}
	if count, err := readUint8(streamInfo); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	} else {
		smplcount += int64(count)
	}
	res.SampleCount = smplcount

	res.AudioMD5 = make([]byte, 16)
	if _, err := io.ReadFull(streamInfo, res.AudioMD5); err != nil {
		return nil, ErrorStreamInfoEarlyEOF
	}

	return &res, nil

}
