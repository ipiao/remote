package remote

import (
	"fmt"
	"net"
	"net/http"
	"net/url"
	"time"
)

// NewProxyRemote 代理
func NewProxyRemote(host, proxy string) *Remote {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		panic(err)
	}
	return &Remote{
		host:    host,
		TimeOut: defaultTimeOut,
		cli: &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, defaultTimeOut)
					if err != nil {
						return nil, err
					}
					return c, nil
				},
				Proxy:           http.ProxyURL(proxyURL),
				MaxIdleConns:    10,
				IdleConnTimeout: defaultTimeOut * 2,
			},
		},
	}
}

// ProxyRemoteStore 代理库
type ProxyRemoteStore struct {
	host    string
	store   Ipstore
	remotes chan *Remote
	size    int
}

// NewProxyRemoteStore creat
func NewProxyRemoteStore(host string, size int, store Ipstore) (*ProxyRemoteStore, error) {
	ssize := store.Size()
	if ssize == 0 || ssize < size {
		return nil, fmt.Errorf("store size is %d,and request size is %d", ssize, size)
	}

	rs := &ProxyRemoteStore{
		host:    host,
		store:   store,
		remotes: make(chan *Remote, size*2),
		size:    size,
	}

	for i := 0; i < size; i++ {
		proxy, err := rs.store.GetHost()
		if err != nil {
			return nil, err
		}
		rs.remotes <- NewProxyRemote(rs.host, proxy)
	}
	return rs, nil
}

// Get one
func (rs *ProxyRemoteStore) Get() (*Remote, error) {
	if rs.size == 0 {
		return rs.New()
	}
	select {
	case r := <-rs.remotes:
		return r, nil
	case <-time.After(time.Millisecond * 5):
		return rs.New()
	}
}

// New create one
func (rs *ProxyRemoteStore) New() (*Remote, error) {
	proxy, err := rs.store.GetHost()
	if err != nil {
		return nil, err
	}
	remote := NewProxyRemote(rs.host, proxy)
	return remote, nil
}
