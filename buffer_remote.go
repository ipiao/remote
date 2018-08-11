package remote

import (
	"fmt"
	"time"

	"github.com/ipiao/metools/cache"
)

type dataBuffer interface {
	GetData(k1, k2 string) interface{}
	PutData(k1, k2 string, obj interface{}, expired ...time.Duration)
}

// BufferRemote 带缓存的调用
type BufferRemote struct {
	*Remote
	dataBuffers dataBuffer
}

// NewBufferRemote new
func NewBufferRemote(host string) *BufferRemote {
	remote := &BufferRemote{
		Remote: NewRemote(host),
	}
	remote.dataBuffers = cache.NewDataBuffer(500)
	return remote
}

// SetDataBuffer reset dataBuffer
func (r *BufferRemote) SetDataBuffer(d dataBuffer) {
	r.dataBuffers = d
}

// BufferGet get data from buffer first
func (r *BufferRemote) BufferGet(url string, req interface{}, ret interface{}) error {
	mk := getMethodKey("GET", url)
	uk := getURLKey(req)
	if r := r.dataBuffers.GetData(mk, uk); r != nil {
		ret = r
		return nil
	}
	res := r.Get(url, req, ret)
	if res == nil {
		r.dataBuffers.PutData(mk, uk, req)
	}
	return res
}

func getURLKey(req interface{}) string {
	s, _ := GobEncode(req)
	return s
}

func getMethodKey(method, url string) string {
	return fmt.Sprintf("%s-%s", method, url)
}
