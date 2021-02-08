package main

import (
	"database/sql"
	"encoding/json"
	"flag"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"github.com/gorilla/websocket"
	"net/http"
	"os"
	"time"
)

var (
	//общее подключение к бд
	db *sql.DB
	//пул соединений с ws
	state = WsState{users: make(map[int][]UserState)}
	//загрузка конфига
	cfg = loadConfig(*flag.String("conf",".wsConfig", "Path to config file"))
	//адрес для ws
	wsIp = flag.String("ws", cfg.String("WsIP", "192.168.0.101:8081"), "http service address")
	//адрес для tcp
	tcpIp = flag.String("tcp", cfg.String("TcpIP", "192.168.0.101:8081"), "http service address")
	//вкл/выкл SSL
	sslMode = cfg.String("sslmode", "true")
	//ключи для SSL
	sslCert = cfg.String("sslcert", "wp.crt")
	sslKey  = cfg.String("sslkey", "wp.key")
	//токен из куков, по которому можно авторизироваться
	cookieAuthKey = cfg.String("authKeyName", "web_profit_session")
	//url для получения авторизации и id от основного сайта
	authCenter = cfg.String("authUrl", "https://web-profit.tk/ws/getauth")
	//модификация апгрейдера соединения
	upgrader = websocket.Upgrader{
		//разрешаем коннект с другого хоста
		//todo: сделать проверку на origin чтоб не кидали конект с других сайтов
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}
	errorLogFile, _ = os.OpenFile(".wsErrors", os.O_APPEND|os.O_CREATE, 0666)

	timeLimits = RequestLimits{Writing: 5}
)

func main() {

	flag.Parse()

	var err error

	// соединение с базой
	db, err = sql.Open("mysql", fmt.Sprintf("%s:%s@%s/%s?charset=utf8mb4,utf8&",
		cfg.String("user", "user"),
		cfg.String("pass", "pass"),
		cfg.String("connection", "connection"),
		cfg.String("dbname", "dbname"),
	))
	fatal(err)
	defer db.Close()

	//чек соединения с бд
	checkDb()

	//tcp сервер (для приема HTTP сообщений от php-сервера)
	flowPrintln("Start listening " + *tcpIp + "...")
	go startTcpServer(*tcpIp)
	flowPrintln("HTTP server is running.")

	// прослушка порта (для приема WebSocket соединений)
	flowPrintln("Start listening " + *wsIp + "...")
	flowPrintln("WebSocket server is running.")
	flowPrintln("\r\nProgram started successfully.")

	if sslMode == "true" {
		http.ListenAndServeTLS(*wsIp, sslCert, sslKey, Handler{}) // calls the ServeHTTP method of the Handler structure
	} else {
		http.ListenAndServe(*wsIp, Handler{})
	}

	//ui.Main(setupUI)
}

func (h Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	//апгрейд соединения
	c, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Print("upgrade:", err)
		return
	}

	//отловить куки
	session := getCookieByName(r.Cookies(), cookieAuthKey)

	//если не авторизирован, бездействие
	id := session2Id(session)
	if id == 0 { //** Скорее всего соединение не закрывается при отключении клиента
		fmt.Println("auth: fail")
		return //** todo: сделать горутину ожидающую ws.OpClose и закрывающую соединение
	}

	//инициализация каналов
	messageChannel := make(chan JsonResponseMessage)
	closeChannel := make(chan int)

	//подготовка дефолтных пользовательских данных

	subscribes := Subscribes {
		1,
		0,
		0,
		0,
	}

	userState := UserState{
		&subscribes,
		&messageChannel,
	}

	//добавляем пользователя в общий пул

	state.Lock()
	state.users[id] = append(state.users[id], userState)
	state.Unlock()

	//запуск горутин
	go reader(c, &messageChannel, &id)
	go writer(c, &messageChannel, &closeChannel, &id)
}

func reader(c *websocket.Conn, messageChannel *chan JsonResponseMessage, id *int) {
	//отложенные функции
	defer func() { *messageChannel <- JsonResponseMessage{Action: "closeConnection"} }()

	var counters RequestLimits

	for {
		//читаем сообщения из соединения
		messageType, message, err := c.ReadMessage()
		if err != nil {
			fmt.Println("read:", err)
			break
		}

		fmt.Printf("Message type %d from id %d: %s\n", messageType, *id, message)

		//раскидываем json
		var data JsonResponseMessage
		json.Unmarshal(message, &data)

		switch data.Action {
		case "sendMessage":
			messageHandle(*id, data)
		case "viewMessage":
			viewMessageHandler(*id, data, true)
		case "getMessages":
			getMessagesHandle(*id, data, data.Id)
		case "writing":
			currentTime := time.Now().Unix()
			//не кидать сообщение если не прошло 5сек после предыдущего
			if (currentTime - counters.Writing) > timeLimits.Writing {
				writingHandle(*id, data)
				counters.Writing = currentTime
			}
		case "deleteMessage":
			//todo
		case "admDeleteMessage":
			//todo
		case "editMessage":
			//todo
		default:
		}
	}
}

func writer(c *websocket.Conn, messageChannel *chan JsonResponseMessage, closeChannel *chan int, id *int) {
	//отложенные функции
	defer c.Close()
	defer deleteUser(id, messageChannel) //удаляем пользователя из общего пула (отлож)
	for {
		select {
		case response := <-*messageChannel:
			if response.Action == "closeConnection" {
				fmt.Println("Close connection")
				return
			}
			response.Body = Nl2br(response.Body)

			//сериализуем сообщение
			message, err := json.Marshal(response)
			err = c.WriteMessage(1, message)
			if err != nil {
				fmt.Println("write:", err)
			}

		case <-*closeChannel:
			fmt.Println("Close connection")
			return
		}
	}
}
