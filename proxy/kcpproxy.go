package proxy

import (
	"crypto/sha1"
	"log"
	"net"

	"github.com/net-byte/opensocks/common/enum"
	"github.com/net-byte/opensocks/config"
	"github.com/xtaci/kcp-go/v5"
	"golang.org/x/crypto/pbkdf2"
)

type KCPProxy struct {
	config   config.Config
	listener *kcp.Listener
}

func (p *KCPProxy) ConnectServer() net.Conn {
	key := pbkdf2.Key([]byte(p.config.Key), []byte("opensocks@2022"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	c, err := kcp.DialWithOptions(p.config.ServerAddr, block, 10, 3)
	if err != nil {
		log.Printf("[client] failed to dial kcp server %s %v", p.config.ServerAddr, err)
		return nil
	}
	c.SetWindowSize(enum.SndWnd, enum.RcvWnd)
	log.Printf("[client] kcp server connected %s", p.config.ServerAddr)
	return c
}

func (p *KCPProxy) StartServer() {
	key := pbkdf2.Key([]byte(p.config.Key), []byte("opensocks@2022"), 1024, 32, sha1.New)
	block, _ := kcp.NewAESBlockCrypt(key)
	var err error
	if p.listener, err = kcp.ListenWithOptions(p.config.ServerAddr, block, 10, 3); err == nil {
		log.Printf("opensocks kcp server started on %s", p.config.ServerAddr)
		for {
			conn, err := p.listener.AcceptKCP()
			if err != nil {
				break
			}
			conn.SetWindowSize(enum.SndWnd, enum.RcvWnd)
			go p.Handler(conn)
		}
	}
}

func (p *KCPProxy) StopServer() {
	if err := p.listener.Close(); err != nil {
		log.Printf("failed to shutdown kcp server: %v", err)
	}
}

func (p *KCPProxy) Handler(conn net.Conn) {
	MuxHandler(conn, p.config)
}
