package socks5

import (
    "bytes"
    "encoding/binary"
    "log"
    "net"
    "strconv"

    "github.com/itviewer/opensocks/common/enum"
    "github.com/xtaci/smux"
)

// https://datatracker.ietf.org/doc/html/rfc1928#

func parseAddr(b []byte) (host string, port string) {
    /**
      +----+-----+-------+------+----------+----------+
      |VER | CMD |  RSV  | ATYP | DST.ADDR | DST.PORT |
      +----+-----+-------+------+----------+----------+
      | 1  |  1  | X'00' |  1   | Variable |    2     |
      +----+-----+-------+------+----------+----------+
    */
    len := len(b)
    switch b[3] {
    case enum.Ipv4Address:
        host = net.IPv4(b[4], b[5], b[6], b[7]).String()
    case enum.FqdnAddress:
        host = string(b[5 : len-2])
    case enum.Ipv6Address:
        host = net.IP(b[4:20]).String()
    default:
        return "", ""
    }
    port = strconv.Itoa(int(b[len-2])<<8 | int(b[len-1]))
    return host, port
}

func parseUDPData(b []byte) (dstAddr *net.UDPAddr, header []byte, data []byte) {
    /*
       +----+------+------+----------+----------+----------+
       |RSV | FRAG | ATYP | DST.ADDR | DST.PORT |   DATA   |
       +----+------+------+----------+----------+----------+
       |  2 |   1  |   1  | Variable |     2    | Variable |
       +----+------+------+----------+----------+----------+
    */
    if b[2] != 0x00 {
        log.Printf("[udp] not support frag %v", b[2])
        return nil, nil, nil
    }
    switch b[3] {
    case enum.Ipv4Address:
        dstAddr = &net.UDPAddr{
            IP:   net.IPv4(b[4], b[5], b[6], b[7]),
            Port: int(b[8])<<8 | int(b[9]),
        }
        header = b[0:10]
        data = b[10:]
    case enum.FqdnAddress:
        dlen := int(b[4])
        domain := string(b[5 : 5+dlen])
        ipAddr, err := net.ResolveIPAddr("ip", domain)
        if err != nil {
            log.Printf("[udp] failed to resolve dns %s:%v", domain, err)
            return nil, nil, nil
        }
        dstAddr = &net.UDPAddr{
            IP:   ipAddr.IP,
            Port: int(b[5+dlen])<<8 | int(b[6+dlen]),
        }
        header = b[0 : 7+dlen]
        data = b[7+dlen:]
    case enum.Ipv6Address:
        {
            dstAddr = &net.UDPAddr{
                IP:   net.IP(b[4:20]),
                Port: int(b[20])<<8 | int(b[21]),
            }
            header = b[0:22]
            data = b[22:]
        }
    default:
        return nil, nil, nil
    }
    return dstAddr, header, data
}

// resp is a response
func resp(conn net.Conn, rep byte) {
    /**
      +----+-----+-------+------+----------+----------+
      |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
      +----+-----+-------+------+----------+----------+
      | 1  |  1  | X'00' |  1   | Variable |    2     |
      +----+-----+-------+------+----------+----------+
    */
    conn.Write([]byte{enum.Socks5Version, rep, 0x00, 0x01, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00})
}

// respNoAuth is a no auth response
func respNoAuth(conn net.Conn) {
    /**
      +----+--------+
      |VER | METHOD |
      +----+--------+
      | 1  |   1    |
      +----+--------+
    */
    conn.Write([]byte{enum.Socks5Version, enum.NoAuth})
}

// respSuccess is a success response
func respSuccess(conn net.Conn, ip net.IP, port int) {
    /**
      +----+-----+-------+------+----------+----------+
      |VER | REP |  RSV  | ATYP | BND.ADDR | BND.PORT |
      +----+-----+-------+------+----------+----------+
      | 1  |  1  | X'00' |  1   | Variable |    2     |
      +----+-----+-------+------+----------+----------+
    */
    resp := []byte{enum.Socks5Version, enum.SuccessReply, 0x00, 0x01}
    buffer := bytes.NewBuffer(resp)
    binary.Write(buffer, binary.BigEndian, ip)
    binary.Write(buffer, binary.BigEndian, uint16(port))
    conn.Write(buffer.Bytes())
}

func newMuxSession(conn net.Conn) (*smux.Session, error) {
    // KeepAliveDisabled: false,
    smuxConfig := smux.DefaultConfig()
    smuxConfig.Version = enum.SmuxVer
    smuxConfig.MaxReceiveBuffer = enum.SmuxBuf
    smuxConfig.MaxStreamBuffer = enum.StreamBuf
    return smux.Client(conn, smuxConfig)
}
