package main

import (
	"encoding/json"
	"fmt"
	"net/http"
)

//обработка сообщений от tcp сервера
func tcpHandle(w http.ResponseWriter, r *http.Request) {

	var data JsonActionMessage

	err := json.NewDecoder(r.Body).Decode(&data)
	if err != nil || data.Rcpt == 0 {
		w.WriteHeader(400)
		fmt.Println("data.id=0 нету json")
		return
	}

	message := JsonResponseMessage{
		Action: data.Action,
		Body:   data.Body,
		Rcpt:   data.Rcpt,
		From:   data.From,
		Params: "",
	}

	response := ResponseMessage{data.Rcpt, "", message}

	switch message.Action {
	case "sendMessage":
		sendMessageToUser(response)
	case "showMessage":
		sendMessageToUser(response)
	case "allIsViewed":
		sendMessageToUser(response)
	case "closeConnection":
		sendMessageToUser(response)
	case "downloadProgress":
		fmt.Println("DownloadProgressStarted")
		sendMessageToUser(response)
	default:
	}

	w.WriteHeader(201)
}

func startTcpServer(ip string) {

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		tcpHandle(w, r)
	})

	http.ListenAndServe(ip, nil)
}
