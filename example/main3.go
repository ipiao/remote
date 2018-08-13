package main

import (
	"log"
	"time"

	"github.com/ipiao/remote"
)

var (
	pt      = remote.XiciProxyTypeWN
	timeout = time.Second * 10
	nsjHost = "https://nsj-m.yy0578.com"

	ipStore         = remote.NewXiciRedisStore("118.25.7.38:6379", "", pt)
	accessableStore = remote.NewRedisIPStore("118.25.7.38:6379", "", "accessable_pool")

	proxyRemoteStore, _           = remote.NewProxyRemoteStoreTimeout(nsjHost, 0, ipStore, timeout)
	accessableProxyRemoteStore, _ = remote.NewProxyRemoteStoreTimeout(nsjHost, 0, accessableStore, timeout)
)

func makeNsjOpts(r *remote.ProxyRemote, store *remote.RedisIPStore) []remote.Option {
	return []remote.Option{store.NotBad, func(*remote.ProxyInfo) bool {
		ret := make(map[string]interface{})
		err := r.Post("/v1/bbs/queryDetails", map[string]interface{}{"detailId": "654"}, &ret)
		if err != nil {
			log.Println(err)
		}
		return err == nil
	}}
}

func initIPStore(pages []int) {
	err := remote.InitXiCiIppool(pages, pt, ipStore, "")
	if err != nil {
		log.Println("Error-InitXiCiIppool:", err)
		time.Sleep(time.Second * 10)
		initIPStore(pages)
	}
}

func initAccessablePool(size int, page int) {

	nowsize := accessableStore.Size()
	if nowsize < size {
		i := 0
		for i < size-nowsize {
			info, err := ipStore.Get()
			if err != nil {
				log.Println("Error-ipStore.Get:", err)
				if ipStore.Size() == 0 {
					page++
					initIPStore([]int{page})
					initAccessablePool(size, page)
				}
				continue
			}

			r, err := proxyRemoteStore.New()
			if err != nil {
				log.Println("Error-proxyRemoteStore.New:", err)
				continue
			}
			opts := makeNsjOpts(r, accessableStore)
			err = accessableStore.Save(info, opts...)
			if err == nil {
				i++
			}
			ipStore.DelBad(info)
		}
	}

	log.Println("initAccessablePool success:", size)

	go func() {
		select {
		case <-time.After(time.Minute * 10):
			if ipStore.Size() > size*2 {
				go func() {
					defer recover()
					initIPStore([]int{page/50 + 1})
				}()
				page++
			}
			initAccessablePool(size, page)
		}
	}()

}

func createRequest(r map[string]interface{}) map[string]interface{} {
	r["deviceId"] = "A0644CSC-5FFB-431B-DFD1-323C2F34537D"
	r["appVersion"] = "1.0.2"
	r["deviceName"] = "iPhone10.2(iOS11.4.1)"
	r["deviceType"] = "2"
	return r
}

type DetailResult struct {
	Success   bool `json:"success"`
	DetailsPO struct {
		PostsList []struct {
			Id       int `json:"id"`
			Praises  int `json:"praises"`
			Comments int `json:"comments"`
		} `json:"postsList"`
	} `json:"detailsPO"`
}

type Detail struct {
	Success bool `json:"success"`
	Detail  struct {
		Id       int `json:"id"`
		Praises  int `json:"praises"`
		Comments int `json:"comments"`
	} `json:"detail"`
}

func getMaxPraize(did int) (max, second, self int, err error) {
	r := remote.NewRemote(nsjHost) //accessableProxyRemoteStore.New()
	// if err != nil {
	// 	log.Println(err)
	// 	return
	// }
	req := map[string]interface{}{
		"labelPO": map[string]interface{}{
			"id":        80,
			"labelName": "#鸟斯基#",
		},
		"labelActsCount": 1,
		"currentPage":    1,
		"limit":          10,
		"isPaging":       true,
	}
	ret := new(DetailResult)
	err = r.Post("/v1/bbs/queryPostsRecommendDetails", createRequest(req), ret)
	if err != nil {
		log.Println("queryPostsRecommendDetails:", err)
		// accessableStore.DelBad(r.ProxyInfo())
		return
	}

	hasSelf := false
	prises := []int{}
	for _, p := range ret.DetailsPO.PostsList {
		prises = append(prises, p.Praises)
		if p.Praises >= max {
			second = max
			max = p.Praises
		} else {
			if p.Praises > second {
				second = p.Praises
			}
		}
		if p.Id == did {
			self = p.Praises
			hasSelf = true
		}
	}

	if !hasSelf {
		nret := new(Detail)
		err = r.Post("/v1/bbs/queryDetails", createRequest(map[string]interface{}{"detailId": "654"}), &nret)
		self = nret.Detail.Praises
	}
	log.Println(prises)
	log.Println("RETURN-getMaxPraize:", max, second, self, err)
	return
}

func main() {
	var err error
	// err = ipStore.Clear()
	// log.Println(err)
	// err = ipStore.ClearBad()
	// log.Println(err)
	// err = accessableStore.Clear()
	// log.Println(err)
	// err = accessableStore.ClearBad()
	// log.Println(err)
	// initIPStore([]int{1})
	initAccessablePool(3, 2)

	var max, second, self int

GetMaxPraize:
	max, second, self, err = getMaxPraize(654)
	if err != nil {
		time.Sleep(time.Minute * 10)
		log.Println("10 seconds later getMaxPraize")
		goto GetMaxPraize
	}

	if self == max {
		log.Print("当前已经是第一了，比第二名多:", self-second)
	} else {
		log.Print("当前已经不是第一了，比第一名少:", max-self)
	}

}
