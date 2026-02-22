package debug

import (
	"fmt"
	"net"
	"os"
	"sync"
	"time"
)

type DebugServer struct {
	listener net.Listener
	clients  map[net.Conn]bool
	mu       sync.RWMutex
	running  bool
}

var server *DebugServer

func StartDebugServer() error {
	if server != nil {
		return nil
	}

	socketPath := "/tmp/recliner_debug.sock"
	os.Remove(socketPath) // Remove existing socket if it exists

	listener, err := net.Listen("unix", socketPath)
	if err != nil {
		return fmt.Errorf("failed to start debug server: %v", err)
	}

	server = &DebugServer{
		listener: listener,
		clients:  make(map[net.Conn]bool),
		running:  true,
	}

	go server.acceptConnections()
	return nil
}

func StopDebugServer() {
	if server == nil {
		return
	}

	server.running = false
	server.listener.Close()

	server.mu.Lock()
	for client := range server.clients {
		client.Close()
	}
	server.mu.Unlock()
}

func SendDebugMessage(category, msg string) {
	if server == nil {
		return
	}

	timestamp := time.Now().Format("15:04:05")
	fullMsg := fmt.Sprintf("[%s] [%s] %s\n", timestamp, category, msg)

	server.mu.RLock()
	for client := range server.clients {
		go func(c net.Conn) {
			_, err := c.Write([]byte(fullMsg))
			if err != nil {
				server.mu.Lock()
				delete(server.clients, c)
				c.Close()
				server.mu.Unlock()
			}
		}(client)
	}
	server.mu.RUnlock()
}

func (s *DebugServer) acceptConnections() {
	for s.running {
		conn, err := s.listener.Accept()
		if err != nil {
			if s.running {
				fmt.Printf("Debug server accept error: %v\n", err)
			}
			break
		}

		s.mu.Lock()
		s.clients[conn] = true
		s.mu.Unlock()

		// Send welcome message
		conn.Write([]byte("=== ReCLIner Debug Server Connected ===\n"))
	}
}
