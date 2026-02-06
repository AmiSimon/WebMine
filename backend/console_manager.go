package backend

import (
	"bufio"
	"encoding/json"
	"fmt"
	"html/template"
	"net/http"
	"os/exec"
	"sync"

	"github.com/gorilla/websocket"
)

const PathToMcServer = "../mcserver/"

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

var mcServer = &McServer{}

type McServer struct {
	cmd    *exec.Cmd
	stdin  *bufio.Writer
	mu     sync.Mutex
	active bool
	ws     *websocket.Conn
}

type HTMXMessage struct {
	Command string            `json:"command"`
	Headers map[string]string `json:"HEADERS"`
}

func (mc *McServer) Start() error {
	fmt.Println("Attempting to start Minecraft server")
	mc.mu.Lock()
	if mc.active {
		mc.mu.Unlock()
		fmt.Println("Server start failed: already running")
		if mc.ws != nil {
			mc.ws.WriteMessage(websocket.TextMessage, []byte("Server already running!"))
		}
		return fmt.Errorf("\nserver already running")
	}
	mc.active = true
	mc.mu.Unlock()

	command := "java"
	arg1 := "-Xmx1024M"
	arg2 := "-Xms1024M"
	arg3 := "-jar"
	arg4 := "server.jar"
	arg5 := "nogui"

	cmd := exec.Command(command, arg1, arg2, arg3, arg4, arg5)
	cmd.Dir = PathToMcServer
	mc.cmd = cmd

	fmt.Printf("\nExecuting command: %s %s in directory %s", command, arg1, PathToMcServer)

	// Pipes from cmd
	stdout, err := mc.cmd.StdoutPipe()
	if err != nil {
		fmt.Printf("\nError getting stdout pipe: %v", err)
		mc.active = false
		return err
	}

	stderr, err := mc.cmd.StderrPipe()
	if err != nil {
		fmt.Printf("\nError getting stderr pipe: %v", err)
		mc.active = false
		return err
	}

	stdin, err := mc.cmd.StdinPipe()
	if err != nil {
		fmt.Printf("\nError getting stdin pipe: %v", err)
		mc.active = false
		return err
	}
	mc.stdin = bufio.NewWriter(stdin)

	// Run the Command
	if err := mc.cmd.Start(); err != nil {
		fmt.Printf("\nError starting server: %v", err)
		mc.active = false
		return err
	}

	fmt.Println("Minecraft server process started successfully")

	// Pipe stdout to websocket
	go func() {
		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			text := scanner.Text()
			if mc.ws != nil {
				mc.ws.WriteMessage(websocket.TextMessage, []byte(text))
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("\nError reading stdout: %v", err)
		}
	}()

	// Pipe stderr to websocket
	go func() {
		scanner := bufio.NewScanner(stderr)
		for scanner.Scan() {
			text := scanner.Text()
			fmt.Printf("\n[MC-STDERR] %s", text)
			if mc.ws != nil {
				mc.ws.WriteMessage(websocket.TextMessage, []byte(text))
			}
		}
		if err := scanner.Err(); err != nil {
			fmt.Printf("\nError reading stderr: %v", err)
		}
	}()

	// Wait for process to finish
	go func() {
		err := mc.cmd.Wait()
		mc.mu.Lock()
		mc.active = false
		mc.mu.Unlock()

		if err != nil {
			fmt.Printf("\nServer process exited with error: %v", err)
		} else {
			fmt.Println("Server process exited normally")
		}
		if mc.ws != nil {
			mc.ws.WriteMessage(websocket.TextMessage, []byte("[Server stopped]"))
		}
	}()

	return nil
}

func (mc *McServer) Stop() error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.active {
		fmt.Println("Cannot stop server: not running")
		return fmt.Errorf("\nserver not running")
	}

	fmt.Println("Stopping Minecraft server...")

	_, err := mc.stdin.WriteString("stop\n")
	if err != nil {
		fmt.Printf("\nError writing stop command: %v", err)
		return err
	}

	if err := mc.stdin.Flush(); err != nil {
		fmt.Printf("\nError flushing stop command: %v", err)
		return err
	}

	fmt.Println("Stop command sent successfully")
	return nil
}

func (mc *McServer) Restart() error {
	mc.mu.Lock()
	if !mc.active {
		mc.mu.Unlock()
		fmt.Println("Server not running, starting ...")
		return mc.Start()
	}
	mc.mu.Unlock()

	fmt.Println("Restarting Minecraft server...")

	if err := mc.Stop(); err != nil {
		fmt.Printf("\nFailed to stop server: %v", err)
		return err
	}

	if mc.cmd != nil && mc.cmd.Process != nil {
		mc.cmd.Wait()
		fmt.Println("Server process has fully stopped")
	}

	if err := mc.Start(); err != nil {
		fmt.Printf("\nFailed to start server: %v", err)
		return err
	}

	fmt.Println("Minecraft server restarted successfully")
	return nil
}

func (mc *McServer) SendCommand(command string) error {
	mc.mu.Lock()
	defer mc.mu.Unlock()

	if !mc.active {
		fmt.Printf("\nCannot send command '%s': server not running", command)
		return fmt.Errorf("\nserver not running")
	}

	fmt.Printf("\nSending command to MC server: %s", command)

	// write command for minecraft input
	_, err := mc.stdin.WriteString(command + "\n")
	if err != nil {
		fmt.Printf("\nError writing command: %v", err)
		return err
	}

	// send now
	if err := mc.stdin.Flush(); err != nil {
		fmt.Printf("\nError flushing command: %v", err)
		return err
	}

	fmt.Printf("\nCommand sent successfully: %s", command)
	return nil
}

func (mc *McServer) SetWebSocket(ws *websocket.Conn) {
	mc.mu.Lock()
	defer mc.mu.Unlock()
	mc.ws = ws
}

func WsHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\nNew WebSocket connection from %s", r.RemoteAddr)

	ws, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		fmt.Printf("\nWebSocket upgrade error from %s: %v", r.RemoteAddr, err)
		return
	}
	defer func() {
		ws.Close()
		fmt.Printf("\nWebSocket connection closed for %s", r.RemoteAddr)
	}()

	mcServer.SetWebSocket(ws)

	for {
		_, message, err := ws.ReadMessage()
		var HTMXMessage HTMXMessage
		jsonErr := json.Unmarshal([]byte(message), &HTMXMessage)
		if jsonErr != nil {
			fmt.Println("Error decoding JSON:", err)
			return
		}
		if err != nil {
			fmt.Printf("\nWebSocket read error from %s: %v", r.RemoteAddr, err)
			break
		}
		command := HTMXMessage.Command
		fmt.Printf("\nReceived message from %s: %s", r.RemoteAddr, command)

		if err := mcServer.SendCommand(command); err != nil {
			fmt.Printf("\nFailed to send command: %v", err)
			ws.WriteMessage(websocket.TextMessage, []byte("Error: "+err.Error()))
		}
	}
}

func StartHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\nStart server request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := mcServer.Start(); err != nil {
		fmt.Printf("\nFailed to start server: %v", err)
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Minecraft server started",
	})
}

func StopHandeler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\nStop server request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := mcServer.Stop(); err != nil {
		fmt.Printf("\nFailed to stop server: %v", err)
		http.Error(w, fmt.Sprintf("Error: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Stop command sent to Minecraft server",
	})
}
func RestartHandeler(w http.ResponseWriter, r *http.Request) {
	fmt.Printf("\nRestart server request from %s", r.RemoteAddr)

	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	if err := mcServer.Restart(); err != nil {
		fmt.Printf("\nFailed to Restart server: %v", err)
		http.Error(w, fmt.Sprintf("\nError: %s", err.Error()), http.StatusBadRequest)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]string{
		"status":  "success",
		"message": "Restart command sent to Minecraft server",
	})
}

var consoleTemplate = template.Must(template.New("console.html").ParseFiles("./frontend/templates/console.html"))

func ConsoleHandeler(w http.ResponseWriter, r *http.Request) {
	consoleTemplate.Execute(w, nil)
}
