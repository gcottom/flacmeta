package flacmeta

import (
	"bytes"
	"encoding/binary"
	"io"
)

func encodeUint32(n uint32) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.BigEndian, n); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func encodeComment(comment string) ([]byte, error) {
	out, err := encodeUint32L(uint32(len(comment)))
	if err != nil {
		return nil, err
	}
	return append(out, []byte(comment)...), nil
}

func encodeUint32L(n uint32) ([]byte, error) {
	buf := new(bytes.Buffer)
	if err := binary.Write(buf, binary.LittleEndian, n); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func readUint32L(r io.Reader) (res uint32, err error) {
	err = binary.Read(r, binary.LittleEndian, &res)
	return
}

func readUint8(r io.Reader) (res uint8, err error) {
	err = binary.Read(r, binary.BigEndian, &res)
	return
}

func readUint16(r io.Reader) (res uint16, err error) {
	err = binary.Read(r, binary.BigEndian, &res)
	return
}

func readUint32(r io.Reader) (res uint32, err error) {
	err = binary.Read(r, binary.BigEndian, &res)
	return
}

func readBytesWith32bitSize(r io.Reader) (res []byte, err error) {
	var size uint32
	size, err = readUint32(r)
	if err != nil {
		return
	}
	bufall := new(bytes.Buffer)
	for size > 0 {
		var nn int
		buf := make([]byte, size)
		nn, err = r.Read(buf)
		if err != nil {
			return
		}
		bufall.Write(buf)
		size -= uint32(nn)
	}
	res = bufall.Bytes()
	return
}

func readFLACStream(f io.ReadSeeker) ([]byte, error) {
	result, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if result[0] != 0xFF || result[1]>>2 != 0x3E {
		return nil, ErrorNoSyncCode
	}
	return result, nil
}

func checkFLACStream(f io.ReadSeeker) error {
	res := make([]byte, 2)
	_, err := f.Read(res)
	if err != nil {
		return err
	}
	if res[0] != 0xFF || res[1]>>2 != 0x3E {
		return ErrorNoSyncCode
	}
	_, err = f.Seek(-2, io.SeekCurrent)
	return nil
}

func parseMetadataBlock(f io.ReadSeeker) (block *MetaDataBlock, isfinal bool, err error) {
	block = new(MetaDataBlock)
	header := make([]byte, 4)
	_, err = io.ReadFull(f, header)
	if err != nil {
		return
	}
	isfinal = header[0]>>7 != 0
	block.Type = BlockType(header[0] << 1 >> 1)
	var length uint32
	err = binary.Read(bytes.NewBuffer(header), binary.BigEndian, &length)
	if err != nil {
		return
	}
	length = length << 8 >> 8

	buf := make([]byte, length)
	_, err = io.ReadFull(f, buf)
	if err != nil {
		return
	}
	block.Data = buf

	return
}

func readMetadataBlocks(f io.ReadSeeker) (blocks []*MetaDataBlock, err error) {
	finishMetaData := false
	for !finishMetaData {
		var block *MetaDataBlock
		block, finishMetaData, err = parseMetadataBlock(f)
		if err != nil {
			return
		}
		blocks = append(blocks, block)
	}
	return
}

func readFLACHead(f io.ReadSeeker) error {
	buffer := make([]byte, 4)
	_, err := io.ReadFull(f, buffer)
	if err != nil {
		return err
	}
	if string(buffer) != "fLaC" {
		return ErrorNoFLACHeader
	}
	return nil
}
