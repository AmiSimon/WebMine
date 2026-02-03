package main

import (
	"Skyfield1888/WebMine/backend"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Minecraft server WebSocket controller")
	http.HandleFunc("/console/ws", backend.WsHandler)
	http.HandleFunc("/console/start", backend.StartHandler)
	http.HandleFunc("/console/stop", backend.StopHandeler)
	http.HandleFunc("/console/restart", backend.RestartHandeler)
	http.HandleFunc("/console/view", backend.ConsoleHandeler)
	http.Handle("/", http.FileServer(http.Dir("frontend/static")))
	log.Println("Server listening on :8082")
	log.Fatal(http.ListenAndServe(":8082", nil))
}
