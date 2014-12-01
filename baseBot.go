package main

import (
	"bufio"
	"code.google.com/p/go.net/websocket"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"os"
)

type ResponseHandler func(string)

type IRCCResponse struct {
	Success bool   `json: "success,string"`
	Message string `json: "message,string"`
	Session string `json: "session,omitempty"`
	Token   string `json: "token,omitempty"`
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
	resp, err := http.Post("https://www.irccloud.com/chat/auth-formtoken", "", nil)
	if err != nil {
		panic(err)
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
	var data IRCCResponse
	json.Unmarshal(respBody, &data)
	
	body := url.Values{}
	
	body.Add("email", email)
	body.Add("password", password)
	body.Add("token", data.Token)

	headers := make(http.Header)
	headers.Add("x-auth-formtoken", data.Token)
	headers.Add("User-Agent", "ninja")
	headers.Add("Content-Type", "x-www-form-urlencoded")

	resp, err = post("https://www.irccloud.com/chat/login", headers, body)
	if err != nil {
		panic(err)
	}
	

	respBody, err = ioutil.ReadAll(resp.Body)
	if err != nil {
		panic(err)
	}
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

	if *email == "" {
		flag.PrintDefaults()
		return
	}

	if *password == "" {
		fmt.Print("Password:")
		bio := bufio.NewReader(os.Stdin)
		line, _, _ := bio.ReadLine()
		*password = string(line)
	}

	key, err := getSession(*email, *password)
	if err != nil {
		panic(err)
	}

	config = Config{key, *debug}

	connectServerWS(key, responseHandler)
}
