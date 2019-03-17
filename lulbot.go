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
	"cloud.google.com/go/datastore"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"github.com/gorilla/mux"
	"github.com/line/line-bot-sdk-go/linebot"
	"log"
	"net/http"
	"os"
	"strings"
)

type Command struct {
	Action  string `json:"action"`
	Message string `json:"message"`
}

var (
	bot    *linebot.Client
	ctx    context.Context
	db     *sql.DB
	client *datastore.Client
)

func main() {
	// INIT ctx for cloud datastore
	ctx = context.Background()
	projectID := "just-monika-234604"
	ds, err := datastore.NewClient(ctx, projectID)
	if err != nil {
		log.Fatalf("Failed to create Datastore Client: %v", err)
	}
	client = ds

	// Read LINE vars from Environment
	if client, err := linebot.New(
		os.Getenv("LINE_CHANNEL_SECRET"),
		os.Getenv("LINE_CHANNEL_TOKEN"),
	); err != nil {
		log.Fatal(err)
	} else {
		bot = client
	}

	r := mux.NewRouter()

	r.HandleFunc("/callback", LineCallbackHandler)
	r.HandleFunc("/api/command/", CommandHandler).Methods("GET")
	r.HandleFunc("/api/command/", CreateCommandHandler).Methods("POST")
	r.HandleFunc("/api/command/{action}", GetCommandHandler).Methods("GET")
	r.HandleFunc("/api/command/{action}", DeleteCommandHandler).Methods("DELETE")

	r.PathPrefix("/").Handler(http.FileServer(http.Dir("web")))

	log.Print("Trying to Start Server on Port 3000")
	if err := http.ListenAndServe(":3000", r); err != nil {
		log.Fatal(err)
	}

}

/**
	GET ALL
 */
func CommandHandler(writer http.ResponseWriter, request *http.Request) {
	q := datastore.NewQuery("command")
	var commands []Command
	_, err := client.GetAll(ctx, q, &commands)
	if err != nil {
		log.Printf("Error getting all: %v", err)
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "Error fetching commands: %v", err)
	}

	writer.Header().Set("Content-Type", "application/json")
	var json = json.NewEncoder(writer)
	err = json.Encode(commands)
	if err != nil {
		writer.WriteHeader(500)
		log.Printf("Error encoding commands: %v", err)
	}
}

func GetCommandHandler(writer http.ResponseWriter, request *http.Request) {
	vars := mux.Vars(request)
	cmd := vars["action"]
	log.Printf("GET COMMAND for action %v", vars["action"])

	key := datastore.NameKey("command", cmd, nil)

	var command Command
	err := client.Get(ctx, key, &command)
	if err != nil {
		writer.WriteHeader(500)
		fmt.Fprintf(writer, "Error fetching command: %v", err)
	}

	var json = json.NewEncoder(writer)
	writer.Header().Set("Content-Type", "application/json")

	err = json.Encode(command)
	if err != nil {
		writer.WriteHeader(500)
		log.Printf("Error encoding commands: %v", err)
	}
}

func DeleteCommandHandler(writer http.ResponseWriter, request *http.Request) {

}

func CreateCommandHandler(writer http.ResponseWriter, request *http.Request) {
	writer.WriteHeader(400)
}

/**

 */
func LineCallbackHandler(writer http.ResponseWriter, request *http.Request) {
	events, err := bot.ParseRequest(request)
	if err != nil {
		log.Println(request.Method, request.Header, request.GetBody, err)

		if err == linebot.ErrInvalidSignature {
			writer.WriteHeader(400)
		} else {
			writer.WriteHeader(500)
		}
		return
	}
	for _, event := range events {
		if event.Type == linebot.EventTypeMessage {
			switch message := event.Message.(type) {
			case *linebot.TextMessage:
				if isCmd, cmd := checkForCmd(message.Text); isCmd {
					log.Printf("Received Monika command : %s", cmd)
					msg, err := getMessage(cmd)
					if err == nil {
						if msg != "" {
							if err2, _ := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err2 != nil {
								log.Print(err2)
								return
							}
						}
					} else {
						if cmd == "help" {
							cmds := getAllCommands()
							head := "Here are the available commands: \n"
							msg := head + strings.Join(cmds, "\n")
							if err2, _ := bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(msg)).Do(); err2 != nil {
								log.Print(err2)
								return
							}
						}
					}
				} else if islul, reply := checkForLul(message.Text); islul {
					if _, err = bot.ReplyMessage(event.ReplyToken, linebot.NewTextMessage(reply)).Do(); err != nil {
						log.Print(err)
						return
					}
				}
			}
		}
	}
}

/**
	Internal helper fucntions
 */
func getMessage(s string) (string, interface{}) {
	key := datastore.NameKey("command", s, nil)

	var command Command
	err := client.Get(ctx, key, &command)
	if err != nil {
		log.Println("Error:", err)
		return "", err
	}
	log.Printf("Found Message %v for command %v", command.Message, s)
	return command.Message, nil
}

func getAllCommands() []string {
	q := datastore.NewQuery("command")
	var commands []Command
	_, err := client.GetAll(ctx, q, &commands)
	if err != nil {
		log.Printf("Error getting all: %v", err)
	}

	result := make([]string, len(commands))
	for i, c := range commands {
		result[i] = c.Action
	}
	return result
}

func checkForLul(msg string) (bool, string) {
	lowercase := strings.ToLower(msg)
	hasLul := strings.Contains(lowercase, "lul")
	hasLol := strings.Contains(lowercase, "lol")
	hasLuu := strings.Contains(lowercase, "luu")

	if hasLul {
		return hasLul, "lul"
	} else if hasLuu {
		return hasLuu, "luu"
	} else if hasLol {
		return hasLol, "lol"
	}
	return false, ""
}

func checkForCmd(msg string) (bool, string) {
	lowercase := strings.ToLower(msg)
	if strings.HasPrefix(lowercase, "just") {
		sl := strings.Split(lowercase, " ")
		if len(sl) > 1 {
			return true, sl[1]
		} else {
			return true, "help"
		}
	}
	return false, ""
}
