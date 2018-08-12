package main

import (
	"log"
	"time"

	"github.com/ipiao/remote"
)

// func main() {

// 	// done:=make(chan struct{})

// 	fordo()

// 	// <-done
// }

func fordo() {
	store := remote.NewXiciRedisStore("118.25.7.38:6379", "", remote.XiciProxyTypeNN)
	store.Clear()
	err := remote.InitXiCiIppool([]int{1}, remote.XiciProxyTypeNT, store)
	if err != nil {
		panic(err)
	}

	rs, err := remote.NewProxyRemoteStoreTimeout("https://nsj-m.yy0578.com", 0, store, time.Second*3)
	if err != nil {
		panic(err)
	}

	r, err := rs.New()
	if err != nil {
		panic(err)
	}

	// r := remote.NewProxyRemote("https://nsj-m.yy0578.com", "https://183.129.244.16:18118")
	ret := make(map[string]interface{})
	i := 0
	for i < 1800000 {

		log.Println(i)
		err = r.Post("/v1/bbs/queryDetails", map[string]interface{}{"detailId": "557"}, &ret)
		if err != nil {
			log.Println(err)
			r, err = rs.New()
			if err != nil {
				log.Println(err)
			}
		}
		i++
		log.Println(ret)
		time.Sleep(time.Millisecond * 5)
	}
}

func tickerdo() {
	ticker := time.NewTicker(time.Millisecond * 50)
	for {
		select {
		case <-ticker.C:
			do()
		}
	}
}

func do() {
	r := remote.NewRemote("https://nsj-m.yy0578.com")
	ret := make(map[string]interface{})
	err := r.Post("/v1/bbs/queryDetails", map[string]interface{}{"detailId": "658"}, &ret)
	if err != nil {
		log.Println(err)
	}
	log.Println(ret)
}
