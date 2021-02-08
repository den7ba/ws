package main

import "sync"

//структура диалога в базе
type Dialog struct {
	Id              int    `json:"id"`
	LastId          string `json:"last_id"`
	LastMessageBody string `json:"last_message_body"`
	LastMessageId   string `json:"last_message_id"`
	Member1         string `json:"member1"`
	Member2         string `json:"member2"`
	Viewed          string `json:"viewed"`
	Deleted         string `json:"deleted"`
}

//состояние ws сервера
type WsState struct {
	sync.Mutex
	users map[int][]UserState
}

type UserState struct {
	Subscribes *Subscribes
	Channel *chan JsonResponseMessage
}

type Subscribes struct {
	System		int
	Messages 	int
	Admins 		int
	Specials	int
}

//ответ от Laravel
type JsonResponse struct {
	Id     int    `json:"id"`
	Status string `json:"status"`
}

// User
type User struct {
	Id        int    `json:"id"`
	Firstname string `json:"firstname"`
	Lastname  string `json:"lastname"`
	Access    int    `json:"access"`
	Private   int    `json:"private"`
}

//приходящие от клиента сообщения
type JsonActionMessage struct {
	Action   string      `json:"action"`
	Body     string      `json:"body"`
	Observer int         `json:"observer"`
	From     int         `json:"from"`
	Rcpt     int         `json:"rcpt"`
	Params   interface{} `json:"params"`
}

//сообщения от сервера к клиенту
type JsonResponseMessage struct {
	Id       int         `json:"id"`
	Action   string      `json:"action"`
	Body     string      `json:"body"`
	From     int         `json:"from"`
	Rcpt     int         `json:"rcpt"`
	Observer int         `json:"observer"`
	Dialog   int         `json:"dialog_id"`
	Viewed   int         `json:"viewed"`
	Params   interface{} `json:"params"`
}

//сообщение горутине
type ResponseMessage struct {
	Id      int
	Session string
	Message JsonResponseMessage
}

//костыль ы
type Handler struct{}

//ответ от Laravel
type RequestLimits struct {
	Writing     int64
	SendMessage int64
}
