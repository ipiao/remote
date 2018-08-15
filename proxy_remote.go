package remote

import (
	"errors"
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

// Host return host
func (info *ProxyInfo) Host() string {
	if info == nil {
		return ""
	}
	return info.Protocol + "://" + info.IP + ":" + info.Port
}

// Ipstore is store for ip
type Ipstore interface {
	Save(info *ProxyInfo, opts ...Option) error
	Get() (*ProxyInfo, error)
	Size() int
	Del(*ProxyInfo) error
	Clear() error
	NotBad(*ProxyInfo) bool
	AddBad(*ProxyInfo) error
}

// RedisIPStore redis存储
type RedisIPStore struct {
	conn    redis.Conn
	key     string
	infokey string
	badkey  string
}

// Option to check ProxyInfo
type Option func(*ProxyInfo) bool

// IsBad 是否不可用
func (s *RedisIPStore) NotBad(info *ProxyInfo) bool {
	res, _ := redis.Int(s.conn.Do("SISMEMBER", s.badkey, info.Host()))
	return res == 0
}

// AddBad 添加不可用
func (s *RedisIPStore) AddBad(info *ProxyInfo) error {
	_, err := s.conn.Do("SADD", s.badkey, info.Host())
	return err
}

// Save save a info
func (s *RedisIPStore) Save(info *ProxyInfo, opts ...Option) error {

	if len(opts) == 0 {
		opts = []Option{s.NotBad}
	}

	for _, fn := range opts {
		if !fn(info) {
			s.AddBad(info)
			return errors.New("bad proxy")
		}
	}

	info.Protocol = strings.ToLower(info.Protocol)
	infoBytes, err := EnJSON(info)
	if err != nil {
		return err
	}
	host := info.Host()
	infoStr := string(infoBytes)

	_, err = s.conn.Do("SADD", s.key, host)
	if err != nil {
		return err
	}
	_, err = s.conn.Do("HSET", s.infokey, host, infoStr)
	if err != nil {
		return err
	}
	log.Println("存入:", infoStr)
	return nil
}

// Size get size
func (s *RedisIPStore) Size() int {
	res, err := s.conn.Do("SCARD", s.key)
	size, _ := redis.Int(res, err)
	return size
}

// Get get a host
func (s *RedisIPStore) Get() (*ProxyInfo, error) {
	host, err := redis.String(s.conn.Do("SRANDMEMBER", s.key))
	if err != nil {
		return nil, err
	}
	res, err := redis.Bytes(s.conn.Do("HGET", s.infokey, host))
	if err != nil {
		return nil, err
	}
	info := new(ProxyInfo)
	err = DeJSON(res, info)
	log.Printf("取出：%s,Host: %s\n", string(res), host)
	return info, err
}

// Clear rest the store
func (s *RedisIPStore) Clear() error {
	_, err := s.conn.Do("DEL", s.key)
	if err != nil {
		return err
	}
	_, err = s.conn.Do("DEL", s.infokey)
	return err
}

// ClearBad rest the bad
func (s *RedisIPStore) ClearBad() error {
	_, err := s.conn.Do("DEL", s.badkey)
	return err
}

// Del one
func (s *RedisIPStore) Del(info *ProxyInfo) error {
	host := info.Host()
	_, err := s.conn.Do("SREM", s.key, host)
	if err != nil {
		return err
	}
	_, err = s.conn.Do("HDEL", s.infokey, host)
	return err
}

// DelBad into bad
func (s *RedisIPStore) DelBad(info *ProxyInfo) error {
	err := s.Del(info)
	if err == nil {
		err = s.AddBad(info)
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
		conn:    conn,
		key:     k,
		infokey: k + "_info",
		badkey:  k + "_bad",
	}
	return store
}

// ProxyRemote  is ProxyRemote
type ProxyRemote struct {
	*Remote
	proxyInfo *ProxyInfo
}

// Proxy return proxy
func (r *ProxyRemote) Proxy() string {
	if r.proxyInfo == nil {
		return ""
	}
	return r.proxyInfo.Host()
}

// ProxyInfo return proxyInfo
func (r *ProxyRemote) ProxyInfo() *ProxyInfo {
	return r.proxyInfo
}

// NewProxyRemote 代理
func NewProxyRemote(host string, proxy *ProxyInfo) *ProxyRemote {
	return NewProxyRemoteTimeout(host, proxy, defaultTimeOut)
}

// NewProxyRemoteTimeout 代理
func NewProxyRemoteTimeout(host string, proxy *ProxyInfo, timeout time.Duration) *ProxyRemote {

	remote := NewRemoteTimeout(host, timeout)

	if proxy != nil {
		proxyURL, err := url.Parse(proxy.Host())
		if err != nil {
			panic(err)
		}
		if tr, ok := remote.cli.Transport.(*http.Transport); ok {
			tr.Proxy = http.ProxyURL(proxyURL)
		}
	}
	return &ProxyRemote{Remote: remote, proxyInfo: proxy}
}

// ProxyRemoteStore 代理库
type ProxyRemoteStore struct {
	host    string
	store   Ipstore
	remotes chan *ProxyRemote
	size    int
	timeout time.Duration
}

// NewProxyRemoteStore creat
func NewProxyRemoteStore(host string, size int, store Ipstore) (*ProxyRemoteStore, error) {
	return NewProxyRemoteStoreTimeout(host, size, store, defaultTimeOut)
}

// NewProxyRemoteStoreTimeout creat
func NewProxyRemoteStoreTimeout(host string, size int, store Ipstore, timeout time.Duration) (*ProxyRemoteStore, error) {

	if size != 0 && store.Size() > 0 {
		return nil, errors.New("store size is not enough to initize")
	}

	rs := &ProxyRemoteStore{
		host:    host,
		store:   store,
		size:    size,
		timeout: timeout,
	}

	if size > 0 {
		rs.remotes = make(chan *ProxyRemote, size*2)
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
func (rs *ProxyRemoteStore) Get() (*ProxyRemote, error) {
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
func (rs *ProxyRemoteStore) New() (*ProxyRemote, error) {
	info, err := rs.store.Get()
	remote := NewProxyRemoteTimeout(rs.host, info, rs.timeout)
	return remote, err
}

// New create one
func (rs *ProxyRemoteStore) NewPrxoyRemote(info *ProxyInfo) *ProxyRemote {
	return NewProxyRemoteTimeout(rs.host, info, rs.timeout)
}
