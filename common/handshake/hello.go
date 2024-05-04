package handshake

import (
    "bufio"
    "encoding/json"
    "github.com/itviewer/opensocks/base"
    "io"
    "strconv"
    "time"

    "github.com/itviewer/opensocks/codec"
    "github.com/itviewer/opensocks/common/cipher"
    "github.com/itviewer/opensocks/common/enum"
)

type HelloRequest struct {
    Host      string
    Port      string
    Key       string
    Network   string
    Timestamp string
    Random    string
}

func (r *HelloRequest) MarshalBinary() ([]byte, error) {
    return json.Marshal(r)
}

func (r *HelloRequest) UnmarshalBinary(data []byte) error {
    return json.Unmarshal(data, &r)
}

func HelloToTarget(stream io.ReadWriteCloser, network string, host string, port string, key string, obfs bool) bool {
    req := &HelloRequest{}
    req.Network = network
    req.Host = host
    req.Port = port
    req.Key = key
    req.Timestamp = strconv.FormatInt(time.Now().Unix(), 10)
    req.Random = cipher.Random()
    data, err := req.MarshalBinary()
    if err != nil {
        base.Error("failed to encode request", err)
        return false
    }
    if obfs {
        data = cipher.XOR(data)
    }
    encode, err := codec.Encode(data)
    if err != nil {
        base.Error(err)
        return false
    }
    _, err = stream.Write(encode)
    if err != nil {
        base.Error(err)
        return false
    }
    return true
}

func ReadHelloRequest(reader *bufio.Reader) (bool, HelloRequest) {
    var req HelloRequest
    b, _, err := codec.Decode(reader)
    if err != nil {
        return false, req
    }
    if base.Cfg.Obfs {
        b = cipher.XOR(b)
    }
    if req.UnmarshalBinary(b) != nil {
        base.Debug("failed to decode request %v", err)
        return false, req
    }
    reqTime, _ := strconv.ParseInt(req.Timestamp, 10, 64)
    if time.Now().Unix()-reqTime > int64(enum.Timeout) {
        base.Debug("timestamp expired %v", reqTime)
        return false, req
    }
    if base.Cfg.Key != req.Key {
        base.Debug("error key %s", req.Key)
        return false, req
    }
    return true, req
}
