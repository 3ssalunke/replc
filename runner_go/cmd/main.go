package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/3ssalunke/replc/runner_go/pkg/fs"
	"github.com/3ssalunke/replc/runner_go/pkg/s3"
	"github.com/3ssalunke/replc/runner_go/pkg/term"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
)

type WSOutgoingMessage struct {
	Event   string `json:"event"`
	Content string `json:"content"`
}

type WSIncomingMessageContent struct {
	Dir      string `json:"dir"`
	FilePath string `json:"path"`
	Content  string `json:"content"`
}

type WSIncomingMessage struct {
	Event   string                   `json:"event"`
	Content WSIncomingMessageContent `json:"content"`
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

	// host := r.Host
	// replId := strings.Split(host, ":")[0]
	replId := "opencomputerto"
	socketId := uuid.New()
	terminal := term.NewTerminalManager()

	// Send workspace dir content to client on connection
	workspaceDirPath, err := filepath.Abs(filepath.Join("..", "workspace"))
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

	for {
		mt, message, err := conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error reading incoming message from connection %s: %v", conn.RemoteAddr().String(), err)
			} else {
				log.Printf("connection %s closed: %v", conn.RemoteAddr().String(), err)
				err := terminal.CloseTerminal(socketId)
				if err != nil {
					log.Printf("error closiing terminal %v", err)
				}
			}
			break
		}
		if mt != websocket.TextMessage {
			log.Printf("error processing incoming non text message")
			continue
		}

		var wsMessage WSIncomingMessage
		err = json.Unmarshal(message, &wsMessage)
		if err != nil {
			log.Printf("error unmarshaling ws message to struct: %v", err)
			continue
		}

		switch wsMessage.Event {
		case FETCHDIR:
			dirPath, err := filepath.Abs(filepath.Join("..", "workspace", wsMessage.Content.Dir))
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
			filePath, err := filepath.Abs(filepath.Join("..", "workspace", wsMessage.Content.FilePath))
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
			filePath, err := filepath.Abs(filepath.Join("..", "workspace", wsMessage.Content.FilePath))
			if err != nil {
				log.Printf("error creating path to file %v", err)
			} else {
				err := fs.SaveFile(filePath, wsMessage.Content.Content)
				if err != nil {
					log.Printf("error saving file content to runner instance: %v", err)
				} else {
					err = s3.SaveToS3(fmt.Sprintf("replcs/%s", replId), wsMessage.Content.FilePath, wsMessage.Content.Content)
					if err != nil {
						log.Printf("error saving file content to s3 bucket: %v", err)
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
			}
			continue
		case REQUESTTERMINAL:
			err := terminal.CreateTerminal(socketId)
			if err != nil {
				log.Printf("error creating new terminal: %v", err)
			}
			continue
		case TERMINALDATA:
			output, err := terminal.WriteToTerminal(socketId, wsMessage.Content.Content)
			if err != nil {
				log.Printf("error writing to terminal: %v", err)
			}
			log.Println(output)
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
