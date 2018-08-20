package main

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"math/rand"
	"mime/multipart"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/ipiao/remote"
)

var (
	usedPhoneKey    = "used_phone"
	usedNickNameKey = "used_nick_name"
	nickNameKey     = "nick_name"
	commentKey      = "comment"
	logoKey         = "avatar"
	nickNamePage    = 19
	commentPage     = 21
	logoPage        = 13
)

func storeResource(key, val string) error {
	err := redisClient.SAdd(key, val).Err()
	if err != nil {
		log.Println("Error-storeUsedResource:", err)
	}
	return err
}

func getAllResource(key string) []string {
	res, err := redisClient.SMembers(key).Result()
	if err != nil {
		log.Println("Error-storeUsedResource:", err)
	}
	return res
}

func cardResource(key string) int {
	res, _ := redisClient.SCard(key).Result()
	return int(res)
}

func popResource(key string) (string, error) {
	return redisClient.SPop(key).Result()
}

func isUsedResource(key, val string) bool {
	res, _ := redisClient.SIsMember(key, val).Result()
	return res
}

// 获取随机手机号
func getPhone() string {
	prefixs := []string{"139", "138", "137", "136", "135", "134", "159", "158", "157", "150",
		"151", "152", "188", "187", "182", "183", "184", "178", "130", "131", "132", "156",
		"155", "186", "185", "176", "133", "153", "189", "180", "181", "177"}

	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	len := len(prefixs)
	str := prefixs[r.Intn(len)]

	for i := 0; i < 8; i++ {
		str += strconv.Itoa(rand.Intn(10))
	}

	if isUsedResource(usedPhoneKey, str) {
		return getPhone()
	}
	storeResource(usedPhoneKey, str)
	return str
}

func getNickName() string {
	name, err := popResource(nickNameKey)
	if err != nil || name == "" {
		initNickNameStore(nickNamePage)
		return getNickName()
	}
	if isUsedResource(usedNickNameKey, name) {
		return getNickName()
	}
	storeResource(usedNickNameKey, name)
	return name
}

func initNickNameStore(page int) error {

	url := fmt.Sprintf("http://www.oicq88.com/gaoxiao/%d.htm", page)
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{
		Timeout: time.Second * 30,
	}
	response, err := client.Do(request)
	if err != nil {
		log.Println("遇到错误：", err)
		return err
	}
	if response.StatusCode != 200 {
		err = errors.New(response.Status)
		log.Println("遇到错误：", err)
		return err
	}

	dom, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return err
	}

	dom.Find("body > div.main > div.box > div.boxleft > div.listfix > ul >li >p").Each(func(i int, context *goquery.Selection) {
		//地址
		name := context.Text()
		log.Println(name)
		storeResource(nickNameKey, name)
	})

	return nil
}

func initCommentStore(page int) error {

	url := "https://www.juzimi.com/todayhot?page=" + strconv.Itoa(page)
	// func main() {
	// 	initCommentStore(0)
	// }

	// func initCommentStore(page int) error {

	// 	url := "https://www.juzimi.com/article/%E7%BA%A2%E6%A5%BC%E6%A2%A6?page=" + strconv.Itoa(commentPage)
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{
		Timeout: time.Second * 30,
	}
	response, err := client.Do(request)
	if err != nil {
		log.Println("遇到错误：", err)
		return err
	}
	if response.StatusCode != 200 {
		err = errors.New(response.Status)
		log.Println("遇到错误：", err)
		return err
	}

	dom, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return err
	}

	dom.Find("#block-views-xqarticletermspage-block_1 > div > div > div > div > div.view-content> div > div > div.views-field-phpcode-1 > a").Each(func(i int, context *goquery.Selection) {
		comment := context.Text()
		log.Println(comment)
		storeResource(commentKey, comment)
	})

	return nil
}

func getComment2() string {
	r := remote.NewRemote(nsjHost)
	req := map[string]interface{}{
		"detailId":    did,
		"sort":        " asc",
		"currentPage": rand.Intn(350),
		"limit":       1,
		"isPaging":    true,
		"orderBy":     "create_time",
	}
	var ret = struct {
		CommentsList []struct {
			CommentText string `json:"commentText"`
		} `json:"commentsList"`
	}{}
	err := r.Post("/v1/bbs/queryCommentsList", createRequest(req), &ret)
	if err != nil {
		log.Println("queryCommentsList error", err)
	}
	comment := "保持队形，送上第一 +1"
	if len(ret.CommentsList) > 0 {
		comment = ret.CommentsList[0].CommentText
	}

	// log.Println(ret)
	return comment
}

func getComment() string {
	// comment, err := popResource(commentKey)
	// if err != nil || comment == "" {
	// 	commentPage++
	// 	// initCommentStore2(commentPage)
	// 	return getComment()
	// }

	// storeResource(commentKey, comment)
	// return comment
	return getComment2()

	// comments := []string{"不必太张扬 是花自然香", "美美美", "你最美", "第一是你的", "余生还长 善待自己 继续善良"}

	// r := rand.New(rand.NewSource(time.Now().UnixNano()))
	// len := len(comments)
	// comment := comments[r.Intn(len)]
	// if isUsedResource(commentKey, comment) {
	// 	return getComment()
	// }
	// storeResource(commentKey, comment)
	// return comment
}

// func main() {
// 	// initCommentStore2(2)
// 	log.Println(getComment())
// }

func resetLogos() {

	// req = map[string]interface{}{
	// 	"uid":       info.Uid,
	// 	"sex":       strconv.Itoa(rand.Int() % 2),
	// 	"nickName":  nickName,
	// 	"avatarUrl": avatarUrl,
	// }
	// request3, err := r.CovertRequest("POST", "/v1/userAccount/updateUserInfoById", createRequest(req))
	// if err != nil {
	// 	log.Println(err)
	// }
	// request3.Header.Set("authkey", info.AuthKey)
	// bs, err = r.CallRequest(request3)
	// if err != nil {
	// 	log.Println(ret)
	// }
	// err = remote.DeJSON(bs, &ret)
	// if err != nil {
	// 	log.Println(err)
	// 	dochan <- info
	// 	return
	// }
}

func initLogo(page int) error {
	url := "http://online.sccnn.com/html/cion/index-" + strconv.Itoa(page) + ".htm"
	request, _ := http.NewRequest("GET", url, nil)
	request.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")

	client := &http.Client{
		Timeout: time.Second * 30,
	}
	response, err := client.Do(request)
	if err != nil {
		log.Println("遇到错误：", err)
		return err
	}
	if response.StatusCode != 200 {
		err = errors.New(response.Status)
		log.Println("遇到错误：", err)
		return err
	}

	dom, err := goquery.NewDocumentFromResponse(response)
	if err != nil {
		return err
	}

	// log.Println(dom)
	//<img border="1" style="border:1px solid #c0c0c0;" src="http://online.sccnn.com/img2/5890/hd160507-01.png" onload="chgsize(this)" title="点击查看原图">
	// http://online.sccnn.com/html/cion/png/20160527164235.htm
	dom.Find("body > table > tbody > tr > td > table > tbody > tr > td > table > tbody > tr > td > table > tbody > tr > td > table > tbody > tr > td > a > img").Each(func(i int, context *goquery.Selection) {
		logo, _ := context.Attr("src")
		fileName := filepath.Base(logo)
		pic := "http://online.sccnn.com/img2/5890/" + filepath.Base(logo)
		log.Println(pic)
		// storeResource(commentKey, comment)

		resp, err := client.Get(pic)
		if err != nil {
			log.Println("Get Pic Error ", err)
			return
		}
		defer resp.Body.Close()
		// body, _ := ioutil.ReadAll(resp.Body)

		// out, _ := os.Create("pics/" + logo)
		// io.Copy(out, bytes.NewReader(body))
		// picpath := "pics/" + filepath.Base(logo)
		// err = ioutil.WriteFile(picpath, body, 0777)
		// if err != nil {
		// 	log.Println("WriteFile Error ", err)
		// 	return
		// }

		url, _ := updatePic(resp.Body, fileName)
		log.Println(url)

		if len(url) != 0 {
			storeResource(logoKey, url)
		}
	})

	return nil
}

func updatePic(fd io.Reader, fileName string) (string, error) {
	urll := "https://m.niaosiji.com/v1/fileupload/imageFileUpload"

	buf := new(bytes.Buffer) // caveat IMO dont use this for large files, \
	// create a tmpfile and assemble your multipart from there (not tested)
	w := multipart.NewWriter(buf)

	// Create file field
	fw, err := w.CreateFormFile("file", fileName) // w.CreateFormFile("file", name) //这里的file很重要，必须和服务器端的FormFile一致
	if err != nil {
		fmt.Println("c")
		return "", err
	}
	// fd, err := os.Open(name)
	// if err != nil {
	// 	fmt.Println("d")
	// 	return err
	// }
	// defer fd.Close()
	// Write file field from file to upload
	_, err = io.Copy(fw, fd)
	if err != nil {
		fmt.Println("e")
		return "", err
	}
	// Important if you do not close the multipart writer you will not have a
	// terminating boundry
	w.Close()
	req, err := http.NewRequest("POST", urll, buf)
	if err != nil {
		fmt.Println("f")
		return "", err
	}
	req.Header.Set("Content-Type", w.FormDataContentType())
	var client http.Client
	res, err := client.Do(req)
	if err != nil {
		fmt.Println("g")
		return "", err
	}
	defer res.Body.Close()
	// io.Copy(os.Stderr, res.Body) // Replace this with Status.Code check
	fmt.Println("h")
	b, _ := ioutil.ReadAll(res.Body)

	ret := map[string]string{}
	remote.DeJSON(b, &ret)
	return ret["url"], err
}

func getAvatarUrl() string {
	avatar, err := popResource(logoKey)
	if err != nil || avatar == "" {
		logoPage++
		initLogo(logoPage)
		return getAvatarUrl()
	}
	// if isUsedResource(logoKey, comment) {
	// 	return getComment()
	// }
	// storeResource(commentKey, comment)
	return avatar
}
