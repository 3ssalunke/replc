package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/3ssalunke/replc/runner_go/pkg/fs"
	"github.com/gorilla/websocket"
)

type WSOutgoingMessage struct {
	Event   string `json:"event"`
	Content string `json:"content"`
}

type WSIngoingMessageContent struct {
	Dir      string `json:"dir"`
	FilePath string `json:"path"`
	Content  string `json:"content"`
}

type WSIngoingMessage struct {
	Event   string                  `json:"event"`
	Content WSIngoingMessageContent `json:"content"`
}

const (
	DISCONNECT      = "disconnect"
	FETCHDIR        = "fetchDir"
	FETCHCONTENT    = "fetchContent"
	UPDATECONTENT   = "updateContent"
	REQUESTTERMINAL = "requestTerminal"
	TERMINALDATA    = "terminalData"
	LOADED          = "loaded"
	RESPONSE        = "response"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		// Allow all connections by returing true
		return true
	},
}

func handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("Failed to upgrade to websocket: %v", err)
		return
	}

	defer conn.Close()

	log.Printf("connection established with remote address %s", conn.RemoteAddr().String())

	conn.WriteMessage(websocket.TextMessage, []byte("connection successfull"))

	host := r.Host
	replID := strings.Split(host, ":")[0]
	log.Println(replID)

	// Send workspace dir content to client on connection
	workspaceDirPath, err := filepath.Abs(filepath.Join("..", "..", "replc"))
	if err != nil {
		log.Printf("error creating path to workspace directory %v", err)
	} else {
		rootContent, err := fs.FetchDir(workspaceDirPath, "")
		if err != nil {
			log.Printf("error getting workspace directory content: %v", err)
		} else {
			rootContentString, err := json.Marshal(rootContent)
			if err != nil {
				log.Printf("error converting workspace directory content to json string: %v", err)
			} else {
				wsMessage, err := json.Marshal(WSOutgoingMessage{
					Event:   LOADED,
					Content: string(rootContentString),
				})
				if err != nil {
					log.Printf("error converting websocket message to json string: %v", err)
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
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading incoming message from connection %s: %v", conn.RemoteAddr().String(), err)
			} else {
				log.Printf("connection %s closed: %v", conn.RemoteAddr().String(), err)
			}
			break
		}
		if mt != websocket.TextMessage {
			log.Printf("error processing incoming non text message")
			continue
		}

		var wsMessage WSIngoingMessage
		err = json.Unmarshal(message, &wsMessage)
		if err != nil {
			log.Printf("error unmarshaling ws message to struct: %v", err)
			continue
		}

		switch wsMessage.Event {
		case FETCHDIR:
			dirPath, err := filepath.Abs(filepath.Join("..", "..", "replc", wsMessage.Content.Dir))
			if err != nil {
				log.Printf("error creating path to workspace directory %v", err)
			} else {
				dirContent, err := fs.FetchDir(dirPath, wsMessage.Content.Dir)
				if err != nil {
					log.Printf("error getting workspace directory content: %v", err)
				} else {
					dirContentString, err := json.Marshal(dirContent)
					if err != nil {
						log.Printf("error converting workspace directory content to json string: %v", err)
					} else {
						wsMessage, err := json.Marshal(WSOutgoingMessage{
							Event:   RESPONSE,
							Content: string(dirContentString),
						})
						if err != nil {
							log.Printf("error converting websocket message to json string: %v", err)
						} else {
							conn.WriteMessage(websocket.TextMessage, wsMessage)
						}
					}
				}
			}
			continue
		case FETCHCONTENT:
			filePath, err := filepath.Abs(filepath.Join("..", "..", "replc", wsMessage.Content.FilePath))
			if err != nil {
				log.Printf("error creating path to file %v", err)
			} else {
				fileContent, err := fs.FetchContent(filePath)
				if err != nil {
					log.Printf("error getting file content: %v", err)
				} else {
					wsMessage, err := json.Marshal(WSOutgoingMessage{
						Event:   RESPONSE,
						Content: fileContent,
					})
					if err != nil {
						log.Printf("error converting websocket message to json string: %v", err)
					} else {
						conn.WriteMessage(websocket.TextMessage, wsMessage)
					}
				}
			}
			continue
		case UPDATECONTENT:
			filePath, err := filepath.Abs(filepath.Join("..", "..", "replc", wsMessage.Content.FilePath))
			if err != nil {
				log.Printf("error creating path to file %v", err)
			} else {
				err := fs.SaveFile(filePath, wsMessage.Content.Content)
				if err != nil {
					log.Printf("error getting file content: %v", err)
				} else {
					wsMessage, err := json.Marshal(WSOutgoingMessage{
						Event:   RESPONSE,
						Content: "file content updated successfully",
					})
					if err != nil {
						log.Printf("error converting websocket message to json string: %v", err)
					} else {
						conn.WriteMessage(websocket.TextMessage, wsMessage)
					}
				}
			}
			continue
		case REQUESTTERMINAL:
			continue
		case TERMINALDATA:
			continue
		default:
			log.Println("invalid message event")
		}
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