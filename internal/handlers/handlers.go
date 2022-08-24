package handlers

import (
	"fmt"
	"log"
	"net/http"
	"sort"

	"github.com/CloudyKit/jet/v6"
	"github.com/gorilla/websocket"
)

var wsChan = make(chan WsPayload)

var clients = make(map[WebSocketConnection]string)

var views = jet.NewSet(
	jet.NewOSFileSystemLoader("./html"),
	jet.InDevelopmentMode(),
)

var upgradeConnection = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func Home(w http.ResponseWriter, r *http.Request) {
	err := renderPage(w, "home.jet", nil)
	if err != nil {
		log.Println(err)
	}
}

type WebSocketConnection struct {
	*websocket.Conn
}

type WsJsonResponse struct {
	Action         string   `json:"action"`
	Message        string   `json:"message"`
	MessageType    string   `json:"message_type"`
	ConnectedUsers []string `json:"connected_users"`
}

type WsPayload struct {
	Action   string              `json:"action"`
	Username string              `json:"username"`
	Message  string              `json:"message"`
	Conn     WebSocketConnection `json:"-"`
}

func WsEndPoint(w http.ResponseWriter, r *http.Request) {
	ws, err := upgradeConnection.Upgrade(w, r, nil)
	if err != nil {
		log.Println(err)
	}

	log.Println("Connected")

	var res WsJsonResponse
	res.Message = `<em><small>Connctes to Server</small></em>`

	conn := WebSocketConnection{Conn: ws}
	clients[conn] = ""

	err = ws.WriteJSON(res)
	if err != nil {
		log.Println(err)
	}

	go ListenForWS(&conn)
}

func ListenForWS(conn *WebSocketConnection) {
	defer func() {
		if r := recover(); r != nil {
			log.Println("Error", fmt.Sprintf("%v", r))
		}
	}()
	var payload WsPayload

	for {
		err := conn.ReadJSON(&payload)
		if err != nil {

		} else {
			payload.Conn = *conn
			wsChan <- payload
		}
	}
}

func ListenToWSChannel() {
	var res WsJsonResponse

	for {
		e := <-wsChan
		switch e.Action {
		case "username":
			clients[e.Conn] = e.Username
			users := getUserList()
			res.Action = "list_users"
			res.ConnectedUsers = users
			broadcastToAll(res)

		case "left":
			res.Action = "list_users"
			delete(clients, e.Conn)
			users := getUserList()
			res.ConnectedUsers = users
			broadcastToAll(res)

		case "broadcast":
			res.Action = "list_users"
			res.Message = fmt.Sprintf("%s : %s", e.Username, e.Message)
			broadcastToAll(res)

		}
		// res.Action = "got"
		// res.Message = fmt.Sprintf("Msg %s", e.Action)
		// broadcastToAll(res)

	}
}

func getUserList() []string {
	var userlist []string
	for _, x := range clients {
		if x != "" {
			userlist = append(userlist, x)
		}
	}
	sort.Strings(userlist)
	return userlist
}

func broadcastToAll(res WsJsonResponse) {
	for client := range clients {
		err := client.WriteJSON(res)
		if err != nil {
			log.Println("ws error")
			_ = client.Close()
			delete(clients, client)
		}
	}
}

func renderPage(w http.ResponseWriter, tmpl string, data jet.VarMap) error {
	view, err := views.GetTemplate(tmpl)
	if err != nil {
		log.Println(err)
		return err
	}

	err = view.Execute(w, data, nil)
	if err != nil {
		log.Println(err)
		return err
	}
	return nil
}
