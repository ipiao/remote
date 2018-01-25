package remote

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/ipiao/metools/cache"
)

var defaultTimeOut = time.Second * 5

// Remote for http call
type Remote struct {
	host        string
	dataBuffers *cache.DataBuffer
	TimeOut     time.Duration
}

// NewRemote return a remote
func NewRemote(host string) *Remote {
	return &Remote{
		host:        host,
		dataBuffers: cache.NewDataBuffer(500),
		TimeOut:     defaultTimeOut,
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

	bs, err := r.call("POST", url, payload)
	if err != nil {
		return err
	}
	err = DeJSON(bs, ret)
	return err
}

// Get for get request
func (r *Remote) Get(url string, req interface{}, ret interface{}) error {
	url = url + "?" + mapToURLValues(req).Encode()
	bs, err := r.call("GET", url, nil)
	if err != nil {
		return err
	}
	err = DeJSON(bs, ret)
	return err
}

// BufferGet get data from buffer first
func (r *Remote) BufferGet(url string, req interface{}, ret interface{}) error {
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

// call for call
func (r *Remote) call(method, url string, payload io.Reader) ([]byte, error) {
	var body = []byte{}
	req, _ := http.NewRequest(method, r.host+url, payload)
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

func getURLKey(req interface{}) string {
	s, _ := GobEncode(req)
	return s
}

func getMethodKey(method, url string) string {
	return fmt.Sprintf("%s-%s", method, url)
}
