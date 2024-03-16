package httpx

import (
    "golang.org/x/net/proxy"
    "io"
    "net/http"
)

type HttpProxyHandler struct {
    Dialer proxy.Dialer
}

// 一旦建立 HTTP CONNECT 隧道(Hijack)，将直接使用 conn，不会再使用本方法
func (h *HttpProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
    // dump, err := httputil.DumpRequest(r, true)
    // if err != nil {
    //     http.Error(w, err.Error(), http.StatusInternalServerError)
    //     return
    // }
    // log.Print(string(dump))

    hijack, ok := w.(http.Hijacker)
    if !ok {
        http.Error(w, "webserver doesn't support hijacking", http.StatusInternalServerError)
        return
    }

    port := r.URL.Port()
    if port == "" {
        port = "80"
    }
    // 使用 golang 标准代理库同请求的 HOST建立 Socket 连接，透明转发流量，无论客户端应用是否使用隧道模式
    proxyConn, err := h.Dialer.Dial("tcp", r.URL.Hostname()+":"+port)
    if err != nil {
        http.Error(w, err.Error(), http.StatusBadGateway)
        return
    }
    defer proxyConn.Close()

    // 无论是否 CONNECT 请求，都直接在 tcp 级别转发，不分析流量
    clientConn, _, err := hijack.Hijack()
    if err != nil {
        http.Error(w, err.Error(), http.StatusInternalServerError)
        return
    }
    defer clientConn.Close()

    // CONNECT 隧道请求，一般为 https 流量
    if r.Method == http.MethodConnect {
        clientConn.Write([]byte("HTTP/1.1 200 Connection Established\r\n\r\n"))
    } else {
        // After a call to Hijack, the original Request.Body must not be used.
        // 非 CONNECT 请求（一般为 http 流量），直接转发首次请求，后续请求使用 clientConn 获得
        // 测试地址 http://www.banshujiang.cn/
        r.Write(proxyConn)
    }

    // 按照 http 协议，直到服务端或者浏览器等客户端主动中断请求从而退出函数
    // 转发客户端的 后续 请求给目标服务器
    go io.Copy(proxyConn, clientConn)
    // 转发目标服务器的响应给客户端
    io.Copy(clientConn, proxyConn)
}
