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

var defaultTimeOut = time.Second * 30

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
func NewRemoteTimeout(host string, timeout time.Duration) *Remote {
	return &Remote{
		host: host,
		cli: &http.Client{
			Transport: &http.Transport{
				Dial: func(netw, addr string) (net.Conn, error) {
					c, err := net.DialTimeout(netw, addr, timeout) //设置建立连接超时
					if err != nil {
						return nil, err
					}
					c.SetDeadline(time.Now().Add(timeout * 2)) //设置发送接收数据超时
					return c, nil
				},
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
	bs, _ := json.Marshal(req)
	payload := bytes.NewReader(bs)

	bs, err := r.Call("POST", url, payload)
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

// Call for call
func (r *Remote) Call(method, url string, payload io.Reader) ([]byte, error) {
	var body = []byte{}
	req, _ := http.NewRequest(method, r.host+url, payload)
	req.Header.Add("content-type", "application/json")

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
				var tmp, _ = json.Marshal(v)
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
func EnJSON(v interface{}) (string, error) {
	var ret string
	bs := bytes.NewBufferString(ret)
	var encode = json.NewEncoder(bs)
	err := encode.Encode(v)
	return bs.String(), err
}
