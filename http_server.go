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

type PostMessageReq struct {
	Token   string `json:"token"`
	Message string `json:"msg"`
	Tp      string `json:"type"`
}

type GetMessageReq struct {
	Token             string `json:"token"`
	Limit             int    `json:"limit"`
	LastSeenMessageID int    `json:"last_message_id"`
}

type BeeRegisterReq struct {
	BeeName           string `json:"bee_name"`
	InstanceID        string `json:"instance_id"`
	HeartbeatDuration int    `json:"heartbeat_duration"`
}

type HttpServer struct {
	sm *SessionMgr
}

func parseReq[T PostMessageReq | GetMessageReq | BeeRegisterReq](c *gin.Context) (*T, error) {
	req := new(T)
	jsonData, err := ioutil.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	err = json.Unmarshal(jsonData, &req)
	if err != nil {
		return nil, err
	}
	return req, nil
}

func NewHttpServer(sm *SessionMgr) *HttpServer {
	return &HttpServer{
		sm: sm,
	}
}

func (s *HttpServer) Serve() {
	router := gin.Default()
	router.POST("/post", func(c *gin.Context) {
		req, err := parseReq[PostMessageReq](c)
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
		req, err := parseReq[GetMessageReq](c)
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
		msgs := sess.GetMessages(100, req.LastSeenMessageID)
		if len(msgs) > 0 {
			c.JSON(200, msgs)
		}
	})

	router.POST("/bees/register", func(c *gin.Context) {
		req, err := parseReq[BeeRegisterReq](c)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		bee := NewBee(s.sm.hive, req.BeeName, req.InstanceID, req.HeartbeatDuration)
		err = s.sm.hive.AddBee(bee)
		if err != nil {
			c.AbortWithError(500, err)
			return
		}
		c.JSON(200, gin.H{
			"msg": "ok",
		})
	})

	router.GET("/bees/list", func(c *gin.Context) {
		c.JSON(200, s.sm.hive.AllBees())
	})

	router.GET("/bee/instance/:instance", func(c *gin.Context) {
		instance := c.Param("instance")
		bee := s.sm.hive.Bee(instance)
		if bee == nil {
			c.AbortWithError(404, errors.New("no such bee"))
			return
		}
		c.JSON(200, bee)

	})

	router.GET("/bee/name/:name", func(c *gin.Context) {
		name := c.Param("name")
		bees := s.sm.hive.BeesByName(name)
		if bees == nil {
			c.AbortWithError(404, errors.New("no such bee"))
			return
		}
		c.JSON(200, bees)
	})

	router.GET("/bee/heartbeat/:instance", func(c *gin.Context) {
		instance := c.Param("instance")
		bee := s.sm.hive.Bee(instance)
		if bee == nil {
			c.AbortWithError(404, errors.New("no such bee"))
			return
		}
		bee.UpdateHeartbeat()
		c.JSON(200, gin.H{
			"msg": "ok",
		})
	})

	router.Run(*httpServerAddr)
}
