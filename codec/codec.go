package codec

import (
	"bufio"
	"bytes"
	"encoding/binary"

	"github.com/golang/snappy"
	"github.com/net-byte/opensocks/common/cipher"
	"github.com/net-byte/opensocks/config"
)

// Encode encodes a byte array into a byte array
func Encode(data []byte) ([]byte, error) {
	length := int32(len(data))
	pkg := new(bytes.Buffer)
	err := binary.Write(pkg, binary.LittleEndian, length)
	if err != nil {
		return nil, err
	}
	err = binary.Write(pkg, binary.LittleEndian, data)
	if err != nil {
		return nil, err
	}
	return pkg.Bytes(), nil
}

// Decode decodes a byte array into a byte array
func Decode(reader *bufio.Reader) ([]byte, int32, error) {
	len, _ := reader.Peek(4)
	blen := bytes.NewBuffer(len)
	var dlen int32
	err := binary.Read(blen, binary.LittleEndian, &dlen)
	if err != nil {
		return nil, 0, err
	}
	if int32(reader.Buffered()) < dlen+4 {
		return nil, 0, err
	}
	pack := make([]byte, 4+dlen)
	_, err = reader.Read(pack)
	if err != nil {
		return nil, 0, err
	}
	return pack[4:], dlen, nil
}

func EncodeData(b []byte, config config.Config) []byte {
	if config.Obfs {
		b = cipher.XOR(b)
	}
	if config.Compress {
		b = snappy.Encode(nil, b)
	}
	return b
}

func DecodeData(b []byte, config config.Config) ([]byte, error) {
	var err error
	if config.Compress {
		b, err = snappy.Decode(nil, b)
	}
	if config.Obfs {
		b = cipher.XOR(b)
	}
	return b, err
}
