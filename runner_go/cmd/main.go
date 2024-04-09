package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	fs "github.com/3ssalunke/replc/runner_go/pkg"
	"github.com/gorilla/websocket"
)

type WebsocketMessage struct {
	Event   string `json:"event"`
	Content string `json:"content"`
}

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by returing true
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Failed to upgrade to websocket:", err)
		return
	}
	defer conn.Close()

	conn.WriteMessage(websocket.TextMessage, []byte("connection successfull"))

	host := r.Host
	replID := strings.Split(host, ":")[0]
	log.Println(replID)

	workspaceDirPath, err := filepath.Abs(filepath.Join("..", "..", "replc"))
	if err != nil {
		log.Println("error creating path to workspace directory", err)
	} else {
		rootContent, err := fs.FetchDir(workspaceDirPath, "")
		if err != nil {
			log.Println("error getting workspace directory content", err)
		} else {
			rootContentString, err := json.Marshal(rootContent)
			if err != nil {
				log.Println("error converting workspace directory content to json string", err)
			} else {
				wsMessage, err := json.Marshal(WebsocketMessage{
					Event:   "loaded",
					Content: string(rootContentString),
				})
				if err != nil {
					log.Println("error converting websocket message to json string", err)
				} else {
					conn.WriteMessage(websocket.TextMessage, wsMessage)
				}
			}
		}
	}

	// Infinite loop to handle incoming messages
	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			log.Println("Error reading message:", err)
			break
		}
		log.Printf("%s, %d", string(message), mt)
	}
}

func main() {
	http.HandleFunc("/ws", handleWebSocket)

	port := os.Getenv("PORT")
	if port == "" {
		port = "3001"
	}

	log.Printf("Starting runner websocket server on port %s\n", port)
	err := http.ListenAndServe(fmt.Sprintf(":"+port), nil)
	if err != nil {
		log.Fatal("Failed to start runner websocker server:", err)
	}
}
