package main

import (
	"bytes"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/http"
	"net/url"
	"path"
	"time"

	"github.com/c4pt0r/log"
)

var (
	instanceID string
)

func randomString(n int) string {
	const letterBytes = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	b := make([]byte, n)
	for i := range b {
		b[i] = letterBytes[rand.Intn(len(letterBytes))]
	}
	return string(b)
}

func init() {
	rand.Seed(time.Now().UnixNano())
	instanceID = fmt.Sprintf("bee_%s", randomString(5))
}

var (
	eveAddr = flag.String("eve", "http://localhost:8089", "EveTheBot http server address")
)

func getURL(api ...string) string {
	u, err := url.Parse(*eveAddr)
	if err != nil {
		panic(err)
	}
	u.Path = path.Join(u.Path, path.Join(api...))
	return u.String()
}

func heartbeat() error {
	url := getURL("/bee/heartbeat", instanceID)
	log.I("Sending heartbeat to", url)
	resp, err := http.Get(url)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		if resp.StatusCode == http.StatusNotFound {
			// TODO: re-register
			log.Fatal("timeout, need to re-register")
		}
	}
	return nil
}

func register() error {
	url := getURL("/bees/register")
	log.I("Sending register to", url)
	registerReq := map[string]interface{}{
		"bee_name":           "simple-bee",
		"instance_id":        instanceID,
		"heartbeat_duration": 10,
	}
	b, _ := json.Marshal(registerReq)
	log.I(string(b))
	resp, err := http.Post(url, "application/json", bytes.NewReader(b))
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		log.Fatal("register failed")
	}
	return nil
}

func main() {
	flag.Parse()

	if err := register(); err != nil {
		log.Fatal(err)
	}
	for {
		if err := heartbeat(); err != nil {
			log.Fatal(err)
		}
		time.Sleep(time.Second * 3)
	}
}
