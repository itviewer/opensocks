package handshake

import (
    "bufio"
    "encoding/json"
    "io"
    "log"
    "strconv"
    "time"

    "github.com/itviewer/opensocks/codec"
    "github.com/itviewer/opensocks/common/cipher"
    "github.com/itviewer/opensocks/common/enum"
    "github.com/itviewer/opensocks/common/util"
    "github.com/itviewer/opensocks/config"
)

type AddressingRequest struct {
    Host      string
    Port      string
    Key       string
    Network   string
    Timestamp string
    Random    string
}

func (r *AddressingRequest) MarshalBinary() ([]byte, error) {
    return json.Marshal(r)
}

func (r *AddressingRequest) UnmarshalBinary(data []byte) error {
    return json.Unmarshal(data, &r)
}

func ConnectToHost(stream io.ReadWriteCloser, network string, host string, port string, key string, obfs bool) bool {
    req := &AddressingRequest{}
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

func ReadAddressingRequest(config config.Config, reader *bufio.Reader) (bool, AddressingRequest) {
    var req AddressingRequest
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
