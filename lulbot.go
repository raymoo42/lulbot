// Copyright 2016 LINE Corporation
//
// LINE Corporation licenses this file to you under the Apache License,
// version 2.0 (the "License"); you may not use this file except in compliance
// with the License. You may obtain a copy of the License at:
//
//   http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS, WITHOUT
// WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied. See the
// License for the specific language governing permissions and limitations
// under the License.

package main

import (
	"log"
	"net/http"
	"strings"
	"flag"
	"os"

	"github.com/line/line-bot-sdk-go/linebot"
	"github.com/BurntSushi/toml"
)

type Config struct{
	Secret string
	Token string
}

func main() {
	// Read from flag
	confFilePtr := flag.String("conf","/etc/lulbot/config.toml","TOML config file")
	flag.Parse()
	if _, err := os.Stat(*confFilePtr); err == nil {
		log.Printf("Using \"%s\" as configuration file", *confFilePtr)

	} else {
		if !os.IsNotExist(err) {
			log.Fatal(err)
		} else {
			log.Fatal(err)
		}
	}
	var conf Config
	if _, err := toml.DecodeFile(*confFilePtr, &conf); err != nil {
		log.Fatal(err)
		return
	}

	bot, err := linebot.New(
		conf.Secret,
		conf.Token,
	)
	if err != nil {
		log.Fatal(err)
	}
	// Setup HTTP Server for receiving requests from LINE platform
	http.HandleFunc("/", func(w http.ResponseWriter, req *http.Request) {
		events, err := bot.ParseRequest(req)
		if err != nil {
			if err == linebot.ErrInvalidSignature {
				w.WriteHeader(400)
			} else {
				w.WriteHeader(500)
			}
			return
		}
		for _, event := range events {
			if event.Type == linebot.EventTypeMessage {
				switch message := event.Message.(type) {
				case *linebot.TextMessage:
					islul, reply := checkForLul(message.Text)
					if (islul) {

						if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
							log.Print(err)
						}
					}
				}
			}
		}
	})
	log.Print("Trying to Start Server on Port 3000")
	if err := http.ListenAndServe(":3000", nil); err != nil {
		log.Fatal(err)
	}

}

func checkForLul(msg string) (bool, string) {
	lowercase := strings.ToLower(msg)
	hasLul := strings.Contains(lowercase, "lul")
	hasLuu := strings.Contains(lowercase, "luu")

	if (hasLul) {
		return hasLul, "lul"
	} else if (hasLuu) {
		return hasLuu, "luu"
	} else {
		return false, ""
	}
}
