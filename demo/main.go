package main

import (
	"log"
	"net/http"
	"time"

	"github.com/ipiao/remote"
)

func ipHander(w http.ResponseWriter, r *http.Request) {
	log.Println("111")
	log.Println("remoteAddr:", r.RemoteAddr)
	log.Println("referer:", r.Referer())
	log.Println("header:", r.Header)
}

func client() {
	rs, err := remote.NewProxyRemoteStore("https://127.0.0.1:1234", 0, remote.NewXiciRedisStore("118.25.7.38:6379", "", false))
	if err != nil {
		panic(err)
	}
	for {
		r, err := rs.New()
		if err != nil {
			log.Println(err)
		}
		_, err = r.Call("GET", "/ip", nil)
		if err != nil {
			log.Println(err)
		}
		time.Sleep(time.Second * 1)
	}
}

func serve() {
	http.HandleFunc("/ip", ipHander)

	log.Fatal(http.ListenAndServe(":8999", nil))
}

func main() {
	serve()
	// client()
}
