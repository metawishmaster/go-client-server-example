package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"net"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type Client struct {
	serverAddr string
	timeout    time.Duration
	count      int
	min        int
	max        int
}

func (c *Client) generateNumbers() []int {
	rand.Seed(time.Now().UnixNano())

	numbers := make([]int, c.count)
	for i := 0; i < c.count; i++ {
		numbers[i] = rand.Intn(c.max-c.min+1) + c.min
	}

	return numbers
}

func (c *Client) numbersToString(numbers []int) string {
	strNumbers := make([]string, len(numbers))
	for i, num := range numbers {
		strNumbers[i] = strconv.Itoa(num)
	}
	return strings.Join(strNumbers, ",")
}

func (c *Client) Run(elapsedTime *time.Duration) error {
	endsWithPort := func(s string) bool {
		pattern := regexp.MustCompile(`:\d{1,5}$`)
		return pattern.MatchString(s)
	}

	if !endsWithPort(c.serverAddr) {
		c.serverAddr += ":8080"
	}
	/*
		ipNport := strings.Split(c.serverAddr, ":")
		//	strIp := ipNport[0]
		strPort := ipNport[1]

		port, err := strconv.Atoi(strPort)
		if err != nil || port < 0 || port > 65535 {
			return fmt.Errorf("invalid port", os.ErrInvalid)
		}
		//	ip, err := strconv.Atoi(strIp)
	*/
	host, port, err := net.SplitHostPort(c.serverAddr)
	fmt.Printf("connecting to %s:%s\n", host, port)

	conn, err := net.DialTimeout("tcp", c.serverAddr, c.timeout)
	if err != nil {
		return fmt.Errorf("connection error: %v", err)
	}
	defer conn.Close()

	numbers := c.generateNumbers()
	numbersStr := c.numbersToString(numbers)

	fmt.Printf("%d numbers generated (range: %d-%d): %v\n", c.count, c.min, c.max, numbers)

	conn.SetWriteDeadline(time.Now().Add(c.timeout))

	startTime := time.Now()
	_, err = conn.Write([]byte(numbersStr + "\n"))
	if err != nil {
		return fmt.Errorf("send error: %v", err)
	}

	conn.SetReadDeadline(time.Now().Add(c.timeout))

	reader := bufio.NewReader(conn)
	response, err := reader.ReadString('\n')
	if err != nil {
		return fmt.Errorf("reader error: %v", err)
	}

	response = strings.TrimSpace(response)
	if response != "GOTESTACK" {
		fmt.Printf("wrong ack message: %s\n", response)
		return nil //fmt.Errorf("connection error: %v", err)
	}
	response, err = reader.ReadString('\n')
	if err != nil {
		fmt.Printf("read error: %v\n", err)
		return fmt.Errorf("connection error: %v", err)
	}

	*elapsedTime = time.Since(startTime)

	response = strings.TrimSpace(response)

	if strings.HasPrefix(response, "ERROR:") {
		return fmt.Errorf("server error: %s", response)
	}

	sortedNumbers := strings.Split(response, ",")
	sortedInts := make([]int, len(sortedNumbers))

	for i, numStr := range sortedNumbers {
		num, _ := strconv.Atoi(numStr)
		sortedInts[i] = num
	}

	isSorted := true
	for i := 1; i < len(sortedInts); i++ {
		if sortedInts[i] < sortedInts[i-1] {
			isSorted = false
			break
		}
	}

	fmt.Printf("numbers are sorted: %v\n", isSorted)
	fmt.Printf("response: %v\n", sortedInts)

	return nil
}

func main() {
	serverAddr := flag.String("server", "localhost:8080", "адрес сервера")
	timeout := flag.Int("timeout", 30, "timeout in seconds")
	count := flag.Int("count", 10, "numbers count")
	min := flag.Int("min", 1, "minimum value")
	max := flag.Int("max", 1000, "maximum value")
	flag.Parse()

	client := Client{
		serverAddr: *serverAddr,
		timeout:    time.Duration(*timeout) * time.Second,
		count:      *count,
		min:        *min,
		max:        *max,
	}

	var elapsedTime time.Duration
	if err := client.Run(&elapsedTime); err != nil {
		fmt.Printf("error: %v\n", err)
	}
	fmt.Printf("elapsed time = %v\n", elapsedTime)
}
