package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"path"
	"time"

	"github.com/ipiao/metools/cache"
)

// MicroRPCRemote for micro-prc call
type MicroRPCRemote struct {
	host        string
	path        string
	dataBuffers *cache.DataBuffer
	service     string
	TimeOut     time.Duration
}

// MicroRPCRequest 调用microRPC的格式结构
type MicroRPCRequest struct {
	Method  string                 `json:"method"`
	Service string                 `json:"service"`
	Request map[string]interface{} `json:"request"`
}

// NewMicroRPCRequest 构建一个请求体
func (r *MicroRPCRemote) NewMicroRPCRequest(method string, request map[string]interface{}) *MicroRPCRequest {
	return NewMicroRPCRequest(r.service, method, request)
}

// NewMicroRPCRequest 构建一个请求体
func NewMicroRPCRequest(service, method string, request map[string]interface{}) *MicroRPCRequest {
	req := &MicroRPCRequest{
		Method:  method,
		Service: service,
		Request: request,
	}
	if req.Request == nil {
		req.Request = make(map[string]interface{})
	}
	return req
}

// SetParam 设置参数
func (req *MicroRPCRequest) SetParam(key string, value interface{}) {
	req.Request[key] = value
}

// NewMicroRPCRemote return a MicroRPCRemote
func NewMicroRPCRemote(host, path, service string) *MicroRPCRemote {
	return &MicroRPCRemote{
		host:        host,
		path:        path,
		service:     service,
		dataBuffers: cache.NewDataBuffer(500),
		TimeOut:     defaultTimeOut,
	}
}

// Host return the host
func (r *MicroRPCRemote) Host() string {
	return r.host
}

// URL return the url
func (r *MicroRPCRemote) URL() string {
	return path.Join(r.host, r.path)
}

// Post for post request
func (r *MicroRPCRemote) Post(req *MicroRPCRequest, ret interface{}) error {
	bs, _ := json.Marshal(req)
	payload := bytes.NewReader(bs)

	bs, err := r.call("POST", payload)
	if err != nil {
		return err
	}
	err = DeJSON(bs, ret)
	return err
}

// BufferPost get data from buffer first
func (r *MicroRPCRemote) BufferPost(req *MicroRPCRequest, ret interface{}) error {
	mk := req.getMethodKey()
	uk := req.getURLKey()
	if r := r.dataBuffers.GetData(mk, uk); r != nil {
		ret = r
		return nil
	}
	res := r.Post(req, ret)
	if res == nil {
		r.dataBuffers.PutData(mk, uk, req)
	}
	return res
}

// call for call
func (r *MicroRPCRemote) call(method string, payload io.Reader) ([]byte, error) {
	var body = []byte{}
	req, _ := http.NewRequest(method, r.URL(), payload)
	req.Header.Add("content-type", "application/json")

	client := http.Client{
		Timeout: r.TimeOut,
	}
	res, err := client.Do(req)
	if err != nil {
		return body, ErrRequestCall
	}
	defer res.Body.Close()
	if res.StatusCode == http.StatusGatewayTimeout {
		return body, ErrTimeOut
	}
	body, err = ioutil.ReadAll(res.Body)
	return body, err
}

func (req *MicroRPCRequest) getURLKey() string {
	s, _ := GobEncode(req.Request)
	return s
}

func (req *MicroRPCRequest) getMethodKey() string {
	return fmt.Sprintf("%s.%s", req.Service, req.Method)
}
