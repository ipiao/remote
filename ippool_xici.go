package remote

import (
	"log"
	"math/rand"
	"net/http"
	"net/url"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
)

var (
	xiciHost = "http://www.xicidaili.com/"
)

// defined
const (
	XiciProxyTypeWN XiciProxyType = "wn" // 国内https
	XiciProxyTypeNN XiciProxyType = "nn" // 国内高匿
	XiciProxyTypeNT XiciProxyType = "nt" // 国内普通
	XiciProxyTypeWT XiciProxyType = "wt" // 国内http
	XiciProxyTypeQQ XiciProxyType = "qq" // sock
)

// NewXiciRedisStore 默认key
func NewXiciRedisStore(host, pwd string, pt XiciProxyType) *RedisIPStore {
	key := "ippool_xici_" + string(pt)

	store := NewRedisIPStore(host, pwd, key)
	return store
}

// XiciProxyType proxy type
type XiciProxyType string

// InitXiCiIppool 西刺动态ip池
// page 的值域请具体参考网站
// proxy 有时候爬取ip也需要代理
func InitXiCiIppool(pages []int, pt XiciProxyType, store Ipstore, proxy string, opts ...Option) error {
	if len(pages) == 0 {
		pages = []int{1, 2}
	}

	baseURL := xiciHost + string(pt) + "/"

	for _, page := range pages {

		count := 0
	GETRESP:
		count++
		request, _ := http.NewRequest("GET", baseURL+strconv.Itoa(page), nil)
		request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
		request.Header.Set("Connection", "keep-alive")
		request.Header.Set("User-Agent", getAgent())

		client := &http.Client{
			Timeout: time.Second * 30,
		}
		if len(proxy) > 0 {
			proxyURL, err := url.Parse(proxy)
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
			newProxy, _ := store.Get()
			if newProxy == nil || newProxy.IP != "" && count <= 3 {
				proxy = newProxy.Host()
				goto GETRESP
			}
		}

		dom, err := goquery.NewDocumentFromResponse(response)
		if err != nil {
			return err
		}
		dom.Find("#ip_list tbody tr").Each(func(i int, context *goquery.Selection) {
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

			//存入
			err = store.Save(&ProxyInfo{
				IP:           ip,
				Port:         port,
				Address:      address,
				Anonymous:    anonymous,
				Protocol:     protocol,
				SurvivalTime: survivalTime,
				CheckTime:    checkTime,
			}, opts...)
			if err != nil {
				log.Println(err)
			}
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
