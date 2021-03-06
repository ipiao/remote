package remote

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"net"
	"net/http"
	"net/url"
	"reflect"
	"time"
)

var (
	defaultTimeOut   = time.Second * 30
	defaultKeepAlive = time.Minute * 5
	defaultHeader    = http.Header{
		"content-type": []string{"application/json"},
	}
)

// Remote for http call
type Remote struct {
	host string
	cli  *http.Client
}

// NewRemote return a remote
func NewRemote(host string) *Remote {
	return NewRemoteTimeout(host, defaultTimeOut)
}

// NewRemoteTimeout return a remote

// NewRemoteTimeout return a remote
func NewRemoteTimeout(host string, timeout time.Duration) *Remote {
	return &Remote{
		host: host,
		cli: &http.Client{
			Transport: &http.Transport{
				DialContext: (&net.Dialer{
					Timeout:   timeout,
					KeepAlive: defaultKeepAlive,
					DualStack: true,
				}).DialContext,
				MaxIdleConns:        50,
				IdleConnTimeout:     90 * time.Second,
				TLSHandshakeTimeout: 10 * time.Second,
			},
			Timeout: timeout,
		},
	}
}

// Host return the host
func (r *Remote) Host() string {
	return r.host
}

// Post for post request
func (r *Remote) Post(url string, req interface{}, ret interface{}) error {

	bs, err := EnJSON(req)
	if err != nil {
		return err
	}
	payload := bytes.NewReader(bs)

	bs, err = r.Call("POST", url, payload)
	if err != nil {
		return err
	}
	err = DeJSON(bs, ret)
	return err
}

// Get for get request
func (r *Remote) Get(url string, req interface{}, ret interface{}) error {

	url = url + "?" + mapToURLValues(req).Encode()
	bs, err := r.Call("GET", url, nil)

	if err != nil {
		return err
	}
	err = DeJSON(bs, ret)
	return err
}

// Call for default call
func (r *Remote) Call(method, url string, payload io.Reader) ([]byte, error) {
	req, _ := http.NewRequest(method, r.host+url, payload)
	req.Header = defaultHeader

	return r.CallRequest(req)
}

// CovertRequest 将结构提装换成request
func CovertRequest(method, url string, req interface{}) (*http.Request, error) {
	var payload io.Reader
	if method == "GET" {
		url = url + "?" + mapToURLValues(req).Encode()
	} else {
		bs, err := EnJSON(req)
		if err != nil {
			return nil, err
		}
		payload = bytes.NewReader(bs)
	}
	request, err := http.NewRequest(method, url, payload)
	if err == nil && request != nil {
		request.Header = defaultHeader
	}
	return request, err
}

// CovertRequest 将结构提装换成request
func (r *Remote) CovertRequest(method, url string, req interface{}) (*http.Request, error) {
	return CovertRequest(method, r.host+url, req)
}

// CallRequest for call http.Request
func (r *Remote) CallRequest(req *http.Request) ([]byte, error) {
	var body = []byte{}

	res, err := r.cli.Do(req)
	if err != nil {
		return body, err
	}
	defer res.Body.Close()

	if res.StatusCode != 200 {
		return body, errors.New(res.Status)
	}
	body, err = ioutil.ReadAll(res.Body)
	return body, err
}

func mapToURLValues(i interface{}) url.Values {
	switch i.(type) {
	case map[string]string:
		var m = url.Values{}
		for k, v := range i.(map[string]string) {
			m.Set(k, v)
		}
		return m
	case map[string]interface{}:
		var m = url.Values{}
		for k, v := range i.(map[string]interface{}) {
			var val = reflect.ValueOf(v)
			if val.Kind() == reflect.Struct {
				var tmp, _ = EnJSON(v)
				m.Set(k, string(tmp))
			} else {
				m.Set(k, fmt.Sprint(v))
			}

		}
		return m
	case url.Values:
		return i.(url.Values)
	}
	return url.Values{}
}

// DeJSON decode json to interface
func DeJSON(data []byte, v interface{}) error {
	var decode = json.NewDecoder(bytes.NewBuffer(data))
	decode.UseNumber()
	return decode.Decode(&v)
}

// EnJSON 解析成json
func EnJSON(v interface{}) ([]byte, error) {
	var bs []byte
	bf := bytes.NewBuffer(bs)
	var encode = json.NewEncoder(bf)
	err := encode.Encode(v)
	return bf.Bytes(), err
}
