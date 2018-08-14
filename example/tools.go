package main

import (
	"errors"
	"fmt"
	"log"
	"math/rand"
	"net/http"
	"strconv"
	"time"

	"github.com/PuerkitoBio/goquery"
	"github.com/garyburd/redigo/redis"
)

var (
	usedPhoneKey    = "used_phone"
	usedNickNameKey = "used_nick_name"
	nickNameKey     = "nick_name"
	commentKey      = "comment"
	nickNamePage    = 1
	commentPage     = 0
)

func storeResource(key, val string) error {
	_, err := redisClient.Do("SADD", key, val)
	if err != nil {
		log.Println("Error-storeUsedResource:", err)
	}
	return err
}

func cardResource(key string) int {
	res, _ := redis.Int(redisClient.Do("SCARD", key))
	return res
}

func popResource(key string) (string, error) {
	return redis.String(redisClient.Do("SPOP", key))
}

func isUsedResource(key, val string) bool {
	res, _ := redis.Int(redisClient.Do("SISMEMBER", key, val))
	return res != 0
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

// func main() {
// 	initCommentStore(0)
// }

func initCommentStore(page int) error {

	url := "https://www.juzimi.com/article/%E7%BA%A2%E6%A5%BC%E6%A2%A6?page=" + strconv.Itoa(commentPage)
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

func getComment() string {
	comment, err := popResource(commentKey)
	if err != nil || comment == "" {
		initCommentStore(commentPage)
		return getComment()
	}
	// if isUsedResource(commentKey, comment) {
	// 	return getComment()
	// }
	// storeResource(usedNickNameKey, name)
	return comment
}
