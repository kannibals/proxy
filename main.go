package main

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"

	"github.com/armon/go-socks5"
	"github.com/elazarl/goproxy"
)

func main() {
	user := os.Getenv("PROXY_USER")
	pass := os.Getenv("PROXY_PASSWORD")
	httpPort := os.Getenv("HTTP_PORT")
	if httpPort == "" {
		httpPort = "14050"
	}
	socksPort := os.Getenv("SOCKS_PORT")
	if socksPort == "" {
		socksPort = "14051"
	}

	// --- SOCKS5 ---
	creds := socks5.StaticCredentials{user: pass}
	auth := socks5.UserPassAuthenticator{Credentials: creds}
	conf := &socks5.Config{AuthMethods: []socks5.Authenticator{auth}}
	socksServer, err := socks5.New(conf)
	if err != nil {
		log.Fatal(err)
	}
	go func() {
		log.Println("SOCKS5 running on", socksPort)
		if err := socksServer.ListenAndServe("tcp", ":"+socksPort); err != nil {
			log.Fatal(err)
		}
	}()

	// --- HTTP proxy ---
	proxy := goproxy.NewProxyHttpServer()
	proxy.Verbose = true
	proxy.OnRequest().DoFunc(func(r *http.Request, ctx *goproxy.ProxyCtx) (*http.Request, *http.Response) {
		authHeader := r.Header.Get("Proxy-Authorization")
		if !checkAuth(authHeader, user, pass) {
			return r, goproxy.NewResponse(r, goproxy.ContentTypeText, http.StatusProxyAuthRequired, "Proxy authentication required")
		}
		return r, nil
	})

	log.Println("HTTP proxy running on", httpPort)
	log.Fatal(http.ListenAndServe(":"+httpPort, proxy))
}

func checkAuth(header, user, pass string) bool {
	const prefix = "Basic "
	if !strings.HasPrefix(header, prefix) {
		return false
	}
	payload, err := base64.StdEncoding.DecodeString(header[len(prefix):])
	if err != nil {
		return false
	}
	parts := strings.SplitN(string(payload), ":", 2)
	if len(parts) != 2 {
		return false
	}
	return parts[0] == user && parts[1] == pass
}