package packet

import (
	"io"
//	"log"
	"encoding/binary"
)

type BinaryReader struct {
	reader io.Reader
	offset int64
}

func NewBinaryReader(reader io.Reader) (*BinaryReader, error) {
	binaryReader := new(BinaryReader)
	binaryReader.reader = reader
	binaryReader.offset = 0
	return binaryReader, nil
}

func (self *BinaryReader) Offset() int64 {
	return self.offset
}

func (self *BinaryReader) ReadByte() (uint8, error) {
	var i uint8
	err := binary.Read(self.reader, binary.LittleEndian, &i)
	self.offset += 1
	return i, err
}

func (self *BinaryReader) ReadUINT16() (uint16, error) {
	var i uint16
	err := binary.Read(self.reader, binary.LittleEndian, &i)
	self.offset += 2
	return i, err
}

func (self *BinaryReader) ReadString(size int) ([]byte, error) {
	var i byte
	var cache []byte
	for j := 0; j < size; j++ {
		err := binary.Read(self.reader, binary.LittleEndian, &i)
		if err != nil {
			return nil, err
		}
		cache = append(cache, i)
		self.offset += 1
	}
//	var result string = string(cache)
//	log.Printf("ReadString(%d) = %v = %v", size, cache, result)
	return cache, nil
}

func (self *BinaryReader) ReadUntil(ch byte) ([]byte, error) {
	var i byte
	var cache []byte
	for {
		err := binary.Read(self.reader, binary.LittleEndian, &i)
		if err != nil {
			return nil, err
		}
		cache = append(cache, i)
		self.offset += 1
		if i == ch {
			break
		}
	}
//	var result string = string(cache)
//	log.Printf("ReadUntil(%c) = %v = %v", ch, cache, result)
	return cache, nil
}

