package main

import (
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
)

type ResponseHandler func(string)

type IRCCResponse struct {
	Success bool   `json: "success,string"`
	Message string `json: "message,string"`
	Session string `json: "session,omitempty"`
}

type StreamJSON struct {
	Type string `json: "type,string"`
	URL  string `json: "url,string,omitempty"`
}

type Config struct {
	SessionKey string
	Debug      bool
}

var config Config

func getSession(email, password string) (session string, err error) {
	resp, err := http.PostForm("https://www.irccloud.com/chat/login",
		url.Values{"email": {email}, "password": {password}})
	if err != nil {
		panic(err)
	}
	respBody, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		panic(err.Error())
	}
	var data IRCCResponse
	json.Unmarshal(respBody, &data)

	err = nil
	if data.Success != true {
		err = errors.New("Login Failed, Message: " + data.Message)
	}

	return data.Session, err
}

func irccConfig(sessionKey string) (*websocket.Config, error) {
	config, err := websocket.NewConfig("wss://www.irccloud.com:443",
		"https://www.irccloud.com")
	if err != nil {
		goto Error
	}
	config.Header.Add("Cookie", "session="+sessionKey)
	config.Header.Add("User-Agent", "ninja")

	return config, nil
Error:
	return config, err
}

func connectServerWS(key string, handler ResponseHandler) {
	config, err := irccConfig(key)
	if err != nil {
		panic(err)
	}

	conn, err := websocket.DialConfig(config)
	if err != nil {
		panic(err.Error())
	}
	var msg string
	for {
		err := websocket.Message.Receive(conn, &msg)
		if err != nil {
			panic(err)
		} else {
			handler(msg)
		}
	}
}

func responseHandler(line string) {
	var msg StreamJSON
	json.Unmarshal([]byte(line), &msg)
	if msg.Type == "oob_include" {
		go visitURL(msg.URL)
	}
	if config.Debug {
		fmt.Println(line)
	}
}

func visitURL(url string) {
	client := http.Client{}

	req, err := http.NewRequest("GET", "https://www.irccloud.com"+url, nil)
	if err != nil {
		panic(err)
	}
	req.Header.Add("Cookie", "session="+config.SessionKey)
	req.Header.Add("Accept-Encoding", "gzip")

	_, err = client.Do(req)
	if err != nil {
		panic(err)
	}
}

func main() {
	email := flag.String("email", "", "IRCCloud email")
	password := flag.String("password", "", "IRCCloud password")
	debug := flag.Bool("debug", false, "Print the responses")

	flag.Parse()

	if *email == "" || *password == "" {
		flag.PrintDefaults()
		return
	}

	key, err := getSession(*email, *password)
	if err != nil {
		panic(err)
	}

	config = Config{key, *debug}

	connectServerWS(key, responseHandler)
}
