package socks

import (
	"bytes"
	"io"
	"log"
	"net"
	"strconv"
	"sync"

	"github.com/net-byte/opensocks/codec"
	"github.com/net-byte/opensocks/common/pool"
	"github.com/net-byte/opensocks/common/util"
	"github.com/net-byte/opensocks/config"
	"github.com/net-byte/opensocks/counter"
	"github.com/net-byte/opensocks/proxy"
	"github.com/xtaci/smux"
)

type UDPWorker struct {
	UDPConn   *net.UDPConn
	Config    config.Config
	headerMap sync.Map
	streamMap sync.Map
	Session   *smux.Session
	Lock      sync.Mutex
}

// Start the UDP server
func (u *UDPWorker) Start() *net.UDPConn {
	udpAddr, _ := net.ResolveUDPAddr("udp", u.Config.LocalAddr)
	var err error
	u.UDPConn, err = net.ListenUDP("udp", udpAddr)
	if err != nil {
		log.Panicf("[udp] failed to listen udp %v", err)
	}
	go u.toServer()
	log.Printf("opensocks [udp] client started on %v", u.Config.LocalAddr)
	return u.UDPConn
}

// toServer handle the udp packet from client
func (u *UDPWorker) toServer() {
	defer u.UDPConn.Close()
	buf := pool.BytePool.Get()
	defer pool.BytePool.Put(buf)
	for {
		n, cliAddr, err := u.UDPConn.ReadFromUDP(buf)
		if err != nil {
			break
		}
		b := buf[:n]
		dstAddr, header, data := parseUDPData(b)
		if dstAddr == nil || header == nil || data == nil {
			continue
		}
		key := cliAddr.String()
		var stream io.ReadWriteCloser
		if value, ok := u.streamMap.Load(key); !ok {
			u.Lock.Lock()
			if u.Session == nil {
				var err error
				xconn := proxy.NewProxy(u.Config).ConnectServer()
				if xconn == nil {
					u.Lock.Unlock()
					continue
				}
				u.Session, err = newMuxSession(xconn)
				if err != nil || u.Session == nil {
					log.Println(err)
					u.Lock.Unlock()
					continue
				}
			}
			u.Lock.Unlock()
			stream, err = u.Session.Open()
			if err != nil {
				u.Session = nil
				util.PrintLog(u.Config.Verbose, "failed to open session:%v", err)
				continue
			}
			ok := proxy.ClientHandshake(stream, "udp", dstAddr.IP.String(), strconv.Itoa(dstAddr.Port), u.Config.Key, u.Config.Obfs)
			if !ok {
				u.Session = nil
				log.Println("[udp] failed to handshake")
				continue
			}
			u.streamMap.Store(key, stream)
			u.headerMap.Store(key, header)
			go u.toClient(stream, cliAddr)
		} else {
			stream = value.(io.ReadWriteCloser)
		}
		data = codec.EncodeData(data, u.Config)
		stream.Write(data)
		counter.IncrWrittenBytes(n)
	}
}

// toClient handle the udp packet from server
func (u *UDPWorker) toClient(stream io.ReadWriteCloser, cliAddr *net.UDPAddr) {
	key := cliAddr.String()
	buffer := pool.BytePool.Get()
	defer pool.BytePool.Put(buffer)
	defer stream.Close()
	for {
		n, err := stream.Read(buffer)
		if err != nil {
			break
		}
		if header, ok := u.headerMap.Load(key); ok {
			b := buffer[:n]
			b, err = codec.DecodeData(b, u.Config)
			if err != nil {
				util.PrintLog(u.Config.Verbose, "failed to decode:%v", err)
				break
			}
			var data bytes.Buffer
			data.Write(header.([]byte))
			data.Write(b)
			_, err = u.UDPConn.WriteToUDP(data.Bytes(), cliAddr)
			if err != nil {
				break
			}
			counter.IncrReadBytes(n)
		}
	}
	u.headerMap.Delete(key)
	u.streamMap.Delete(key)
}
