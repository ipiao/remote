package remote

import (
	"encoding/json"
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/garyburd/redigo/redis"
)

// Ipstore is store for ip
type Ipstore interface {
	Save(ip, info string) error
	GetHost() (string, error)
	Size() int
}

// RedisIPStore redis存储
type RedisIPStore struct {
	conn redis.Conn
	key  string
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
	return &RedisIPStore{
		conn: conn,
		key:  k,
	}
}

// NewXiciRedisStore 默认key
func NewXiciRedisStore(host, pwd string, annoymous bool) *RedisIPStore {
	key := "ippool_xici"
	if annoymous {
		key = "ippool_xici_annoymous"
	}
	return NewRedisIPStore(host, pwd, key)
}

// Save save a info
func (s *RedisIPStore) Save(host, info string) error {
	_, err := s.conn.Do("SADD", s.key, host)
	_, err = s.conn.Do("HSET", s.key+"_info", host, info)
	return err
}

// Size get size
func (s *RedisIPStore) Size() int {
	res, err := s.conn.Do("SCARD", s.key)
	size, _ := redis.Int(res, err)
	return size
}

// GetHost get a host
func (s *RedisIPStore) GetHost() (string, error) {
	ip, err := redis.String(s.conn.Do("SRANDMEMBER", s.key))
	res, err := redis.String(s.conn.Do("HGET", s.key+"_info", ip))
	if err != nil {
		return "", err
	}

	res = strings.TrimLeft(res, "[")
	res = strings.TrimRight(res, "]")

	array := strings.Split(res, ",")

	for i := 0; i < len(array); i++ {
		array[i] = strings.Trim(array[i], "\"")
	}
	host := strings.ToLower(array[4]) + "://" + array[0] + ":" + array[1]
	log.Println("取出：", res)
	return host, nil
}

var (
	xici          = "http://www.xicidaili.com/wn/" // 透明代理
	xiciAnonymous = "http://www.xicidaili.com/nn/" // 匿名代理
)

// InitXiCiIppool 西刺动态ip池
// page 的值域请具体参考网站
// proxy 有时候爬取ip也需要代理
func InitXiCiIppool(page int, anonymous bool, store Ipstore, proxy ...string) error {

	baseURL := xici
	if anonymous {
		baseURL = xiciAnonymous
	}

	for i := 1; i <= page; i++ {

		count := 0
	GETRESP:
		count++
		request, _ := http.NewRequest("GET", baseURL+strconv.Itoa(i), nil)
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		request.Header.Set("Connection", "keep-alive")
		request.Header.Set("User-Agent", getAgent())

		client := &http.Client{
			Timeout: time.Second * 30,
		}
		if len(proxy) > 0 {
			proxyURL, err := url.Parse(proxy[0])
			if err != nil {
				return err
			}
			client.Transport = &http.Transport{
				Proxy: http.ProxyURL(proxyURL),
			}
		}
		response, err := client.Do(request)
		if err != nil || response.StatusCode != 200 {
			log.Println("遇到错误：", count, response.Status, err)
			newProxy, _ := store.GetHost()
			if newProxy != "" && count <= 3 {
				proxy = []string{newProxy}
				goto GETRESP
			}
		}

		dom, err := goquery.NewDocumentFromResponse(response)
		if err != nil {
			return err
		}
		dom.Find("#ip_list tbody tr").Each(func(i int, context *goquery.Selection) {
			ipInfo := make(map[string][]string)
			//地址
			ip := context.Find("td").Eq(1).Text()
			if ip == "" {
				return
			}
			//端口
			port := context.Find("td").Eq(2).Text()
			//地址
			address := context.Find("td").Eq(3).Find("a").Text()
			//匿名
			anonymous := context.Find("td").Eq(4).Text()
			//协议
			protocol := context.Find("td").Eq(5).Text()
			//存活时间
			survivalTime := context.Find("td").Eq(8).Text()
			//验证时间
			checkTime := context.Find("td").Eq(9).Text()
			ipInfo[ip] = append(ipInfo[ip], ip, port, address, anonymous, protocol, survivalTime, checkTime)
			hBody, _ := json.Marshal(ipInfo[ip])

			//存入
			store.Save(ip+":"+port, string(hBody))
			log.Println("存入:", string(hBody))
		})
	}
	return nil
}

/**
* 随机返回一个User-Agent
 */
func getAgent() string {
	agent := [...]string{
		"Mozilla/5.0 (Windows NT 6.1; Win64; x64; rv:50.0) Gecko/20100101 Firefox/50.0",
		"Opera/9.80 (Macintosh; Intel Mac OS X 10.6.8; U; en) Presto/2.8.131 Version/11.11",
		"Opera/9.80 (Windows NT 6.1; U; en) Presto/2.8.131 Version/11.11",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; 360SE)",
		"Mozilla/5.0 (Windows NT 6.1; rv:2.0.1) Gecko/20100101 Firefox/4.0.1",
		"Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; The World)",
		"User-Agent,Mozilla/5.0 (Macintosh; U; Intel Mac OS X 10_6_8; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
		"User-Agent, Mozilla/4.0 (compatible; MSIE 7.0; Windows NT 5.1; Maxthon 2.0)",
		"User-Agent,Mozilla/5.0 (Windows; U; Windows NT 6.1; en-us) AppleWebKit/534.50 (KHTML, like Gecko) Version/5.1 Safari/534.50",
	}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	len := len(agent)
	return agent[r.Intn(len)]
}
