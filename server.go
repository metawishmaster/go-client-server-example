package main

import (
	"bufio"
	"flag"
	"fmt"
	"net"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Server struct {
	port     string
	mu       sync.Mutex
	clients  int
	requests int
}

func NewServer(port string) *Server {
	return &Server{
		port:    port,
		clients: 0,
	}
}

func (s *Server) processNumbers(data string) (string, error) {
	strNumbers := strings.Split(strings.TrimSpace(data), ",")
	numbers := make([]int, 0, len(strNumbers))

	for _, strNum := range strNumbers {
		strNum = strings.TrimSpace(strNum)
		if strNum == "" {
			continue
		}

		num, err := strconv.Atoi(strNum)
		if err != nil {
			return "", fmt.Errorf("convertion error '%s': %v", strNum, err)
		}
		numbers = append(numbers, num)
	}

	sort.Ints(numbers)

	sortedStr := make([]string, len(numbers))
	for i, num := range numbers {
		sortedStr[i] = strconv.Itoa(num)
	}

	return strings.Join(sortedStr, ","), nil
}

func (s *Server) handleConnection(conn net.Conn) {
	defer conn.Close()

	clientAddr := conn.RemoteAddr().String()

	s.mu.Lock()
	s.clients++
	clientNum := s.clients
	s.mu.Unlock()

	fmt.Printf("[client #%d] connected from: %s\n", clientNum, clientAddr)

	conn.SetReadDeadline(time.Now().Add(30 * time.Second))

	reader := bufio.NewReader(conn)
	data, err := reader.ReadString('\n')
	if err != nil {
		fmt.Printf("[client #%d] read error: %v\n", clientNum, err)
		return
	}

	s.mu.Lock()
	s.requests++
	s.mu.Unlock()

	fmt.Printf("[client #%d] recieved %d numbers\n", clientNum, len(strings.Split(data, ",")))

	result, err := s.processNumbers(data)
	if err != nil {
		fmt.Printf("[client #%d] processing error: %v\n", clientNum, err)
		conn.Write([]byte("ERROR: " + err.Error() + "\n"))
		return
	}

	conn.SetWriteDeadline(time.Now().Add(10 * time.Second))

	_, err = conn.Write([]byte(result + "\n"))
	if err != nil {
		fmt.Printf("[client #%d] send error: %v\n", clientNum, err)
		return
	}

	fmt.Printf("[client #%d] sorted and sent %d numbers\n",
		clientNum, len(strings.Split(result, ",")))
	fmt.Printf("[client #%d] disconnected\n", clientNum)
}

func (s *Server) Run() {
	listener, err := net.Listen("tcp", ":"+s.port)
	if err != nil {
		fmt.Printf("server error: %v\n", err)
		return
	}
	defer listener.Close()

	fmt.Printf("server started on port %s\n", s.port)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("accept error: %v\n", err)
			continue
		}

		go s.handleConnection(conn)
	}
}

func main() {
	port := flag.String("port", "4040", "server port")
	flag.Parse()

	server := &Server{
		port:    *port,
		clients: 0,
	}
	server.Run()
}
