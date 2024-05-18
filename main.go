package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/gorilla/websocket"
)

type Message string

var (
    clients = make(map[*websocket.Conn]bool)
    broadcast = make(chan Message)
)

var upgrader = websocket.Upgrader{
    CheckOrigin: func(r *http.Request) bool {
        return true 
    },
}

func handleConnections(w http.ResponseWriter, r *http.Request) {
    ws, err := upgrader.Upgrade(w, r, nil)
    if err != nil {
        log.Fatal(err)
    }
    defer ws.Close()
    clients[ws] = true

    fmt.Println("connection established")
    for {
        _, msg, err := ws.ReadMessage()
        if err != nil {
            log.Println("read:", err)
            break
        }
        log.Printf("recv: %s", msg)

        broadcast <- Message(msg)
    }
}

func handleMessage() {
    for {
        msg := <-broadcast
        for client := range clients {
            err := client.WriteMessage(websocket.TextMessage, []byte(msg))
            if err != nil {
                log.Printf("Error: %v", err)
                client.Close()
                delete(clients, client)
            }
        }
    }
}

func main() {
    fs := http.FileServer(http.Dir("./frontend"))
    http.Handle("/", fs)

    http.HandleFunc("/ws", handleConnections)

    go handleMessage()

    log.Println("Server started on :8000")
    err := http.ListenAndServe(":8000", nil)
    if err != nil {
        log.Fatal("ListenAndServe: ", err)
    }
}

