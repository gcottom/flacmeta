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

func readUint32L(r io.Reader) (res uint32, err error) {
	err = binary.Read(r, binary.LittleEndian, &res)
	return
}

func readUint32(r io.Reader) (res uint32, err error) {
	err = binary.Read(r, binary.BigEndian, &res)
	return
}

func readBytesWith32bitSize(r *bytes.Buffer) ([]byte, error) {
	size, err := readUint32(r)
	if err != nil {
		return nil, err
	}
	if uint64(size) > uint64(r.Len()) {
		return nil, io.ErrUnexpectedEOF
	}
	return append([]byte(nil), r.Next(int(size))...), nil
}

func readFLACStream(f io.ReadSeeker) ([]byte, error) {
	result, err := io.ReadAll(f)
	if err != nil {
		return nil, err
	}
	if len(result) < 2 || result[0] != 0xFF || result[1]>>2 != 0x3E {
		return nil, ErrorNoSyncCode
	}
	return result, nil
}

func checkFLACStream(f io.ReadSeeker) error {
	res := make([]byte, 2)
	_, err := io.ReadFull(f, res)
	if err != nil {
		return err
	}
	if res[0] != 0xFF || res[1]>>2 != 0x3E {
		return ErrorNoSyncCode
	}
	_, err = f.Seek(-2, io.SeekCurrent)
	return err
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
