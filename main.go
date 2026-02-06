package main

import (
	"Skyfield1888/WebMine/backend"
	"fmt"
	"log"
	"net/http"
)

func main() {
	log.Println("Starting Minecraft server WebSocket controller")
	//Console
	http.HandleFunc("/console/ws", backend.WsHandler)
	http.HandleFunc("/console/start", backend.StartHandler)
	http.HandleFunc("/console/stop", backend.StopHandeler)
	http.HandleFunc("/console/restart", backend.RestartHandeler)
	http.HandleFunc("/console/view", backend.ConsoleHandeler)
	//Properties Handeler
	http.HandleFunc("/properties/set", backend.ChangePropertiesHandler)
	http.HandleFunc("/properties/view", backend.PropertiesTableHandler)

	// main website handeler
	http.Handle("/", http.FileServer(http.Dir("frontend/static")))
	fmt.Println("Server listening on :8082")
	fmt.Fatal(http.ListenAndServe(":8082", nil))
}
