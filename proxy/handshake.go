package proxy

import (
	"bufio"
	"encoding/json"
	"io"
	"log"
	"strconv"
	"time"

	"github.com/net-byte/opensocks/codec"
	"github.com/net-byte/opensocks/common/cipher"
	"github.com/net-byte/opensocks/common/enum"
	"github.com/net-byte/opensocks/common/util"
	"github.com/net-byte/opensocks/config"
)

type HandshakeRequest struct {
	Host      string
	Port      string
	Key       string
	Network   string
	Timestamp string
	Random    string
}

func (r *HandshakeRequest) MarshalBinary() ([]byte, error) {
	return json.Marshal(r)
}

func (r *HandshakeRequest) UnmarshalBinary(data []byte) error {
	return json.Unmarshal(data, &r)
}

func ClientHandshake(stream io.ReadWriteCloser, network string, host string, port string, key string, obfs bool) bool {
	req := &HandshakeRequest{}
	req.Network = network
	req.Host = host
	req.Port = port
	req.Key = key
	req.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
	req.Random = cipher.Random()
	data, err := req.MarshalBinary()
	if err != nil {
		log.Printf("[client] failed to encode request %v", err)
		return false
	}
	if obfs {
		data = cipher.XOR(data)
	}
	encode, err := codec.Encode(data)
	if err != nil {
		log.Println(err)
		return false
	}
	_, err = stream.Write(encode)
	if err != nil {
		log.Println(err)
		return false
	}
	return true
}

func ServerHandshake(config config.Config, reader *bufio.Reader) (bool, HandshakeRequest) {
	var req HandshakeRequest
	b, _, err := codec.Decode(reader)
	if err != nil {
		return false, req
	}
	if config.Obfs {
		b = cipher.XOR(b)
	}
	if req.UnmarshalBinary(b) != nil {
		util.PrintLog(config.Verbose, "[server] failed to decode request %v", err)
		return false, req
	}
	reqTime, _ := strconv.ParseInt(req.Timestamp, 10, 64)
	if time.Now().Unix()-reqTime > int64(enum.Timeout) {
		util.PrintLog(config.Verbose, "[server] timestamp expired %v", reqTime)
		return false, req
	}
	if config.Key != req.Key {
		util.PrintLog(config.Verbose, "[server] error key %s", req.Key)
		return false, req
	}
	return true, req
}
