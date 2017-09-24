package main

import (
	"log"
	"net/http"

	"github.com/googollee/go-socket.io"
)

type PlayerInfo struct {
	ID    string  `json:"id"`
	X     float64 `json:"x"`
	Y     float64 `json:"y"`
	Angle float64 `json:"angle"`
}

var (
	playerLst []PlayerInfo
)

func main() {
	server, err := socketio.NewServer(nil)
	if err != nil {
		log.Fatal(err)
	}
	server.On("connection", func(socket socketio.Socket) {
		socketID := socket.Id()
		log.Println("on connection", socketID)
		socket.Join("chat")

		socket.On("chat message", func(msg string) {
			log.Println("emit:", socket.Emit("chat message", msg))
			socket.BroadcastTo("chat", "chat message", msg)
		})

		socket.On("new_player", func(playerInfo PlayerInfo) {
			log.Println("hello-new-player", playerInfo)

			current_info := PlayerInfo{
				ID:    socketID,
				X:     playerInfo.X,
				Y:     playerInfo.Y,
				Angle: 0,
			}

			socket.Emit("player_ready", current_info)

			for i := range playerLst {
				existingPlayer := playerLst[i]
				log.Println("existingPlayer", existingPlayer)
				playerInfo := PlayerInfo{
					ID:    existingPlayer.ID,
					X:     existingPlayer.X,
					Y:     existingPlayer.Y,
					Angle: 0,
				}
				socket.Emit("new_enemyPlayer", playerInfo)
			}

			socket.BroadcastTo("chat", "new_enemyPlayer", current_info)
			playerLst = append(playerLst, current_info)
		})
		socket.On("disconnection", func() {
			log.Println("on disconnect")
			jsonSender := struct {
				ID string `json:"id"`
			}{ID: socketID}
			// jsonSender, _ := json.Marshal(mapDataID)
			socket.BroadcastTo("chat", "remove_player", jsonSender)
			index := findPlayerIndexByID(socketID)
			if index > -1 {
				playerLst = append(playerLst[:index], playerLst[index+1:]...)
			}
			log.Println("removing player", socketID)
		})

		socket.On("player_move", func(playerInfo PlayerInfo) {
			index := findPlayerIndexByID(playerInfo.ID)
			targetPlayer := playerLst[index]
			if index > -1 {
				targetPlayer.X = playerInfo.X
				targetPlayer.Y = playerInfo.Y
				socket.BroadcastTo("chat", "player_move", targetPlayer)
			}

		})
	})
	server.On("error", func(so socketio.Socket, err error) {
		log.Println("error:", err)
	})

	// http.Handle("/socket.io/", server)
	http.HandleFunc("/socket.io/", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "http://localhost:3000")
		w.Header().Set("Access-Control-Allow-Credentials", "true")
		server.ServeHTTP(w, r)
	})
	http.Handle("/", http.FileServer(http.Dir("./asset")))
	PORT := "2000"
	log.Println("Serving at localhost:" + PORT + "...")
	log.Fatal(http.ListenAndServe(":2000", nil))
}

func findPlayerIndexByID(ID string) int {
	for i := range playerLst {
		if playerLst[i].ID == ID {
			return i
		}
	}
	return -1
}
