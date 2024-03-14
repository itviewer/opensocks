# Opensocks

A simple multiplexing proxy.

# Features
* Support socks5 proxy
* Support http(s) proxy

# Usage
```
Usage of opensocks:
  -S	server mode
  -bypass
      bypass private ip
  -k string
      encryption key (default "6w9z$C&F)J@NcRfUjXn2r4u7x!A%D*G-")
  -l string
      local socks5 proxy address (default "127.0.0.1:1080")
  -obfs
      enable data obfuscation
  -compress
      enable data compression
  -p string
      protocol ws/wss/kcp/tcp (default "wss")
  -s string
      server address (default ":8081")
  -http string
        local http proxy address (default ":8008")
  -http-proxy
        enable http proxy
  -v    enable verbose output
```
# Run
## Run client
```
./opensocks-linux-amd64 -s=YOUR_DOMIAN:8081 -l=127.0.0.1:1080 -k=123456 -p kcp -obfs
```

## Run client(enable http proxy)
```
./opensocks-linux-amd64 -s=YOUR_DOMIAN:8081 -l=127.0.0.1:1080 -k=123456 -p kcp -obfs -http-proxy -http 127.0.0.1:8000
```

## Run server
```
./opensocks-linux-amd64 -S -k=123456 -obfs -p kcp
```

# Docker

## Run client
```
docker run -d --restart=always  --network=host \
--name opensocks-client netbyte/opensocks -s=YOUR_DOMIAN:8081 -l=127.0.0.1:1080 -k=123456 -p ws -obfs
```

## Run server
```
docker run  -d --restart=always --net=host \
--name opensocks-server netbyte/opensocks -S -k=123456 -obfs
```

## Reverse proxy server
add tls for opensocks ws server(8081) via nginx/caddy(443)

## Server settings
settings for kcp with good performance
```
ulimit -n 65535
vi /etc/sysctl.conf
net.core.rmem_max=26214400 // BDP - bandwidth delay product
net.core.rmem_default=26214400
net.core.wmem_max=26214400
net.core.wmem_default=26214400
net.core.netdev_max_backlog=2048 // proportional to -rcvwnd
sysctl -p /etc/sysctl.conf
```

# Cross-platform client
[opensocks-gui](https://github.com/net-byte/opensocks-gui)
<p>
<a href="https://play.google.com/store/apps/details?id=com.netbyte.opensocks"><img src="https://play.google.com/intl/en_us/badges/images/generic/en-play-badge.png" height="100"></a>
</p>

# Deploy to cloud
[opensocks-cloud](https://github.com/net-byte/opensocks-cloud)

# License
[The MIT License (MIT)](https://raw.githubusercontent.com/net-byte/opensocks/main/LICENSE)

## Credits

This repo relies on the following third-party projects:
- [websocket](https://github.com/gorilla/websocket)
- [kcp-go](https://github.com/xtaci/kcp-go)
- [smux](https://github.com/xtaci/smux)
- [bpool](https://github.com/oxtoacart/bpool)
