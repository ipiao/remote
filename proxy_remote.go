package remote

import (
	"fmt"
	"log"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/garyburd/redigo/redis"
)

// ProxyInfo 记录代理信息
type ProxyInfo struct {
	IP           string
	Port         string
	Address      string
	Anonymous    string
	Protocol     string
	SurvivalTime string
	CheckTime    string
}

// Ipstore is store for ip
type Ipstore interface {
	Save(*ProxyInfo) error
	Get() (string, error)
	Size() int
	Clear() error
}

// RedisIPStore redis存储
type RedisIPStore struct {
	conn redis.Conn
	key  string
}

// Save save a info
func (s *RedisIPStore) Save(info *ProxyInfo) error {
	info.Protocol = strings.ToLower(info.Protocol)
	infoStr, err := EnJSON(info)
	if err != nil {
		return err
	}
	_, err = s.conn.Do("SADD", s.key, infoStr)
	if err != nil {
		return err
	}
	log.Println("存入", infoStr)
	return nil
}

// Size get size
func (s *RedisIPStore) Size() int {
	res, err := s.conn.Do("SCARD", s.key)
	size, _ := redis.Int(res, err)
	return size
}

// Get get a host
func (s *RedisIPStore) Get() (string, error) {
	res, err := redis.Bytes(s.conn.Do("SRANDMEMBER", s.key))
	if err != nil {
		return "", err
	}

	info := new(ProxyInfo)
	err = DeJSON(res, info)
	host := info.Protocol + "://" + info.IP + ":" + info.Port
	log.Printf("取出：%s,Host: %s\n", string(res), host)
	return host, err
}

// Clear rest the store
func (s *RedisIPStore) Clear() error {
	_, err := s.conn.Do("DEL", s.key)
	if err != nil {
		return err
	}
	return err
}

// NewRedisIPStore create
func NewRedisIPStore(host, pwd string, key ...string) *RedisIPStore {
	k := "ippool"
	if len(key) > 0 {
		k = key[0]
	}
	conn, err := redis.Dial("tcp", host, redis.DialPassword(pwd))
	if err != nil {
		panic(err)
	}
	store := &RedisIPStore{
		conn: conn,
		key:  k,
	}
	return store
}

// NewProxyRemote 代理
func NewProxyRemote(host, proxy string) *Remote {
	return NewProxyRemoteTimeout(host, proxy, defaultTimeOut)
}

// NewProxyRemoteTimeout 代理
func NewProxyRemoteTimeout(host, proxy string, timeout time.Duration) *Remote {
	proxyURL, err := url.Parse(proxy)
	if err != nil {
		panic(err)
	}
	remote := NewRemoteTimeout(host, timeout)

	if tr, ok := remote.cli.Transport.(*http.Transport); ok {
		tr.Proxy = http.ProxyURL(proxyURL)
	}
	return remote
}

// ProxyRemoteStore 代理库
type ProxyRemoteStore struct {
	host    string
	store   Ipstore
	remotes chan *Remote
	size    int
	timeout time.Duration
}

// NewProxyRemoteStore creat
func NewProxyRemoteStore(host string, size int, store Ipstore) (*ProxyRemoteStore, error) {
	return NewProxyRemoteStoreTimeout(host, size, store, defaultTimeOut)
}

// NewProxyRemoteStoreTimeout creat
func NewProxyRemoteStoreTimeout(host string, size int, store Ipstore, timeout time.Duration) (*ProxyRemoteStore, error) {
	ssize := store.Size()
	if ssize == 0 || ssize < size {
		return nil, fmt.Errorf("store size is %d,and request size is %d", ssize, size)
	}

	rs := &ProxyRemoteStore{
		host:    host,
		store:   store,
		size:    size,
		timeout: timeout,
	}

	if size > 0 {
		rs.remotes = make(chan *Remote, size*2)
	}

	for i := 0; i < size; i++ {
		r, err := rs.New()
		if err != nil {
			return nil, err
		}
		rs.remotes <- r
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
	proxy, err := rs.store.Get()
	if err != nil {
		return nil, err
	}
	remote := NewProxyRemoteTimeout(rs.host, proxy, rs.timeout)
	return remote, nil
}
