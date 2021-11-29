package main

import (
	"encoding/json"
	"errors"
	"flag"
	"io/ioutil"

	"github.com/c4pt0r/log"
	"github.com/gin-gonic/gin"
)

var (
	httpServerAddr = flag.String("http-endpoint", ":8089", "http server endpoint")
	advisoryAddr   = flag.String("advisory-addr", "http://127.0.0.1:8089", "advisory address, will show in usage text")
)

type Req struct {
	Token   string `json:"token"`
	Message string `json:"msg"`
	Tp      string `json:"type"`
}

type HttpServer struct {
	sm *SessionMgr
}

func NewHttpServer(sm *SessionMgr) *HttpServer {
	return &HttpServer{sm}
}

func (s *HttpServer) parseReq(c *gin.Context) (Req, error) {
	var req Req
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return Req{}, err
	}
	err = json.Unmarshal(jsonData, &req)
	if err != nil {
		return Req{}, err
	}
	return req, nil
}

func (s *HttpServer) Serve() {
	router := gin.Default()

	router.POST("/post", func(c *gin.Context) {
		req, err := s.parseReq(c)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		sess, ok := s.sm.GetSessionByToken(req.Token)
		if !ok {
			c.AbortWithError(404, errors.New("no such chat"))
			return
		}
		sess.SendMarkdown(req.Message)
	})

	router.GET("/message", func(c *gin.Context) {
		req, err := s.parseReq(c)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		sess, ok := s.sm.GetSessionByToken(req.Token)
		if !ok {
			c.AbortWithError(404, errors.New("no such chat"))
			return
		}
		log.I("Get message for session", sess)
		msgs := sess.GetMessages()
		if len(msgs) > 0 {
			c.JSON(200, msgs)
		}
	})
	router.Run(*httpServerAddr)
}
