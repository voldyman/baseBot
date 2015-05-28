package main

import (
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"

	"github.com/howeyc/gopass"
	"golang.org/x/net/websocket"
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

func getAuthToken() (data IRCCResponse, err error) {
	resp, err := http.Post("https://www.irccloud.com/chat/auth-formtoken", "", nil)
	if err != nil {
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	json.Unmarshal(respBody, &data)

	return
}

func getSession(email, password string) (session string, err error) {
	data, err := getAuthToken()
	if err != nil {
		return
	}

	body := url.Values{}

	body.Add("email", email)
	body.Add("password", password)
	body.Add("token", data.Token)

	headers := make(http.Header)
	headers.Add("x-auth-formtoken", data.Token)
	headers.Add("User-Agent", "ninja")
	headers.Add("Content-Type", "x-www-form-urlencoded")

	resp, err := post("https://www.irccloud.com/chat/login", headers, body)
	if err != nil {
		return
	}

	respBody, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return
	}
	json.Unmarshal(respBody, &data)

	err = nil
	if data.Success != true {
		err = errors.New("Login Failed, Message: " + data.Message)
	}

	return data.Session, err
}

func irccConfig(sessionKey string) (config *websocket.Config, err error) {
	config, err = websocket.NewConfig("wss://www.irccloud.com:443",
		"https://www.irccloud.com")
	if err != nil {
		return
	}

	config.Header.Add("Cookie", "session="+sessionKey)
	config.Header.Add("User-Agent", "ninja")
	return
}

func connectServerWS(key string, handler ResponseHandler) (err error) {
	config, err := irccConfig(key)
	if err != nil {
		return
	}

	conn, err := websocket.DialConfig(config)
	if err != nil {
		return
	}
	var msg string
	for {
		err = websocket.Message.Receive(conn, &msg)
		if err != nil {
			return
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

func start(username, password string, debug bool, tries int) {
	key, err := getSession(username, password)
	if err != nil {
		goto ManageError
	}

	config = Config{key, debug}

	err = connectServerWS(key, responseHandler)
	if err != nil {
		goto ManageError
	}

	return
ManageError:
	if tries > 0 {
		fmt.Println("Error:", err.Error())
		fmt.Println("Attempt number", tries)
		start(username, password, debug, tries-1)
	} else {
		fmt.Println("Gave up reattempting")
		fmt.Println("Last error", err.Error())
	}
}

func main() {
	email := flag.String("email", "", "IRCCloud email")
	password := flag.String("password", "", "IRCCloud password")
	debug := flag.Bool("debug", false, "Print the responses")
	attempts := flag.Int("attempts", 5, "Number of attempts to make when disconnected from server")

	flag.Parse()

	if *email == "" {
		flag.PrintDefaults()
		return
	}

	if *password == "" {
		fmt.Print("Password:")
		*password = string(gopass.GetPasswd())
	}

	// try reconnecting 5 or specified times
	start(*email, *password, *debug, *attempts)
}
