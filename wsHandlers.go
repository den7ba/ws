package main

import (
	"encoding/json"
	"fmt"
	"github.com/FogCreek/mini"
	"html"
	"io/ioutil"
	"net/http"
	"os"
	"strings"
)

//delete user from global array of users
func deleteUser(id *int, channel *chan JsonResponseMessage) {
	state.Lock()

	if userStates, ok := state.users[*id]; ok {
		for i, userState := range userStates {
			if channel == userState.Channel {
				state.users[*id] = removeUserStateFromSlice(userStates, i)
			}
		}
	}

	if len(state.users[*id]) == 0 {
		delete(state.users, *id)
	}

	state.Unlock()
}

//global send
func sendMessageToUser(response ResponseMessage) {
	state.Lock()

	userStates := state.users[response.Id]

	for _, userState := range userStates {
		*userState.Channel <- response.Message //todo: существует вероятность что канала уже не будет в пуле (writer закрыт)
	}

	state.Unlock()
}

//get user id from auth center using cookies
func session2Id(session string) int {
	client := &http.Client{}

	req, err := http.NewRequest("GET", authCenter, nil)
	fatal(err)

	req.Header.Set("Cookie", cookieAuthKey+"="+session+"; XSRF=x")
	resp, err := client.Do(req)
	fatal(err)

	var data JsonResponse
	body, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &data)

	defer resp.Body.Close()

	return data.Id
}

func messageHandle(observerId int, data JsonResponseMessage) {
	//проверяем сообщение
	data.Body = strings.TrimSpace(data.Body)
	if data.Body == "" || getUserId(data.Rcpt) == 0 {
		fmt.Println("Invalid message. Iteration skipped.")
		return
	}

	//модифицируем данные
	data.Body = html.EscapeString(data.Body)
	data.Action = "addMessage"
	data.From = observerId
	data.Observer = observerId

	//записываем сообщение в базу
	messageId, err := addMessageToDb(data)
	if err != nil {
		fatal(err)
		fmt.Println("Error writing to the DB. Iteration skipped.")
		return
	}

	//модифицируем данные
	data.Id = messageId

	//возвращаем сообщение отправителю
	response := ResponseMessage{observerId, "", data}
	sendMessageToUser(response)

	//модифицируем данные
	data.Observer = data.Rcpt

	//отправляем сообщение получателю
	response = ResponseMessage{data.Rcpt, "", data}
	sendMessageToUser(response)

	//прибавляем счетчик приватных сообщений получателю
	response = ResponseMessage{data.Rcpt, "", JsonResponseMessage{
		Action: "incPrivate",
	}}

	sendMessageToUser(response)
}

func writingHandle(observerId int, data JsonResponseMessage) {
	//формируем ws-сообщение
	message := JsonResponseMessage{
		Action: "writing",
		Rcpt:   observerId,
	}

	response := ResponseMessage{data.Rcpt, "", message}

	//отправляем обсерверу
	sendMessageToUser(response)
}

//получить последние 20 сообщений
func getMessagesHandle(observerId int, data JsonResponseMessage, first int) {

	var messages = make([]JsonResponseMessage, 0)
	var rec JsonResponseMessage
	var action = ""

	if first == 0 {
		action = "messagesPack"
	} else {
		action = "AddsPack"
	}

	//получаем данные участников диалога
	observer, err := getUser(observerId)
	if err != nil {
		fatal(err)
		fmt.Println("Error reading DB. Iteration skipped.1")
		return
	}

	opponent, err := getUser(data.Rcpt)
	if err != nil {
		fatal(err)
		fmt.Println("Error reading DB. Iteration skipped.2")
		return
	}

	//читаем последние 20 сообщений из бд (или 20 с указанного места first)
	rows, err := readMessages(observerId, data.Rcpt, first, 20)
	if err != nil {
		fatal(err)
		fmt.Println("Error reading DB. Iteration skipped.3")
		return
	}

	//обработка строк
	for rows.Next() {
		if err = rows.Scan(&rec.Id, &rec.Dialog, &rec.From, &rec.Rcpt, &rec.Body, &rec.Viewed); err != nil {
			fmt.Println("Error reading DB. Iteration skipped.3")
			return
		}

		//просмотр непросмотренных сообщений
		if rec.Viewed == 0 && rec.Rcpt == observerId {
			rec.Viewed = 1
			viewMessageHandler(observerId, rec, false)
		}

		//добавляем недостающие данные
		rec.Observer = observerId
		// \n to <br>
		rec.Body = Nl2br(rec.Body)
		messages = append(messages, rec)
	}

	defer rows.Close()

	if err = rows.Err(); err != nil {
		fmt.Println("Error reading DB. Iteration skipped.3")
		return
	}

	//убираем ненужные данные
	observer.Access = 0
	opponent.Access = 0

	//сериализуем всё в json
	serializedMessages, err := json.Marshal(messages)
	serializedObserver, err := json.Marshal(observer)
	serializedOpponent, err := json.Marshal(opponent)

	//формируем ws-сообщение
	message := JsonResponseMessage{
		Action: action,
		Body:   string(serializedMessages),
		Params: map[string]string{
			"observer": string(serializedObserver),
			"opponent": string(serializedOpponent),
		},
	}

	response := ResponseMessage{observerId, "", message}

	//отправляем обсерверу
	sendMessageToUser(response)
}

//просмотр сообщения и оповещение собеседника
func viewMessageHandler(observerId int, data JsonResponseMessage, check bool) {
	message, err := viewMessage(data, observerId, check)
	if err != nil {
		return
	}

	//отнимаем счетчик приватных сообщений
	response := ResponseMessage{observerId, "", JsonResponseMessage{
		Action: "decPrivate",
	}}

	sendMessageToUser(response)

	//оповещаем отправителя что сообщение прочитано
	response = ResponseMessage{message.From, "", JsonResponseMessage{
		Id:     message.Id,
		Action: "hasBeenViewed",
		Rcpt:   message.Rcpt, //id чата
	}}

	sendMessageToUser(response)
}

func loadConfig(path string) *mini.Config {
	flowPrintln("Loading configuration...")
	flowPrintln("▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓ ▓")
	fmt.Println("")
	exists, err := fileExists(path)
	fatal(err)

	if !exists {
		cfgFile, err := os.OpenFile(path, os.O_APPEND|os.O_CREATE, 0666)
		fatal(err)
		defer cfgFile.Close()
		configs := `connection = tcp(192.168.0.101:3306)
			host = "127.0.0.1"
			port = "3306"
			sslmode = true
			dbname = "mysite"
			user = "root"
			pass = ""
			WsIP = "192.168.0.101:8081"
			TcpIP = "192.168.0.101:8079"
			authKeyName = web_profit_session
			authUrl = http://web-profit.tk/ws/getauth
			sslkey = userdata/config/cert_files/wp.key
			sslcert = userdata/config/cert_files/wp.crt`

		cfgFile.WriteString(configs)

		fmt.Println("Конфиг не загружен, используются значения по умолчанию. Будет создан файл .wsConfig с дефолтными настройками!\r\n ")
	}

	cfg, err := mini.LoadConfiguration(path)
	fatal(err)

	return cfg
}

/*
func isAuth(session string) bool{
	client := &http.Client{}
	req, err := http.NewRequest("GET", "http://web-profit.tk/ws/getauth", nil)
	fatal(err)

	req.Header.Set("Cookie", "web_profit_session="+session+"; XSRF=x")
	resp, err := client.Do(req)
	fatal(err)

	var data JsonResponse
	body, err := ioutil.ReadAll(resp.Body)

	err = json.Unmarshal(body, &data)

	defer resp.Body.Close()
	if data.Status == "success" {
		return true
	}
	return false
}
*/
