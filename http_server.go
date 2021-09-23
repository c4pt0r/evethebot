package main

import (
	"encoding/json"
	"flag"
	"io/ioutil"
	"log"

	"github.com/gin-gonic/gin"
)

var (
	httpServerAddr = flag.String("http-endpoint", ":8089", "http server endpoint")
)

type Req struct {
	Token   string `json:"token"`
	Message string `json:"msg"`
	Tp      string `json:"type"`
}

func serveHttp() {
	router := gin.Default()

	router.POST("/post", func(c *gin.Context) {
		var req Req
		jsonData, err := ioutil.ReadAll(c.Request.Body)
		if err != nil {
			log.Println(err)
			return
		}
		err = json.Unmarshal(jsonData, &req)
		if err != nil {
			log.Println(err)
			return
		}
		// get session by chat id, if not exists create one
		sess, ok := SM().GetSessionByToken(req.Token)
		if ok {
			sess.Send(req.Message)
		}
	})
	router.Run(*httpServerAddr)
}
