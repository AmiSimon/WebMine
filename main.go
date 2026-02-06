package main

import (
	"Skyfield1888/WebMine/backend"
	"fmt"
	"log"
	"net/http"
)

func main() {
	backend.DecodeConfig()
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

	//App Setting Handeler
	http.HandleFunc("/settings/set", backend.ChangeAppSettingsHandler)
	http.HandleFunc("/settings/view", backend.AppSettingsTableHandler)

	// main website handeler
	http.Handle("/", http.FileServer(http.Dir("frontend/static")))
	fmt.Println("Server listening on :8082")
	log.Fatal(http.ListenAndServe(":"+backend.SavedAppConfig.WebAppConfig.Port, nil))
}
