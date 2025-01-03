package MisraSocket

import (
	"MisraSockets/util"
	"bufio"
	"fmt"
	"net"
	"strconv"
)

type Action int

const (
	PassThrough Action = iota
	Incarnated
	CriticalSection
)

type MisraSocket struct {
	listenPort string
	sendPort   string
	sendIp     string
	m          int
	ping       int
	pong       int
	ringSize   int
	action     Action
	hasPing    bool
	hasPong    bool
}

func NewSocket(listenPort string, sendIp string, sendPort string, ringSize int) *MisraSocket {
	return &MisraSocket{
		listenPort: listenPort,
		sendPort:   sendPort,
		sendIp:     sendIp,
		ringSize:   ringSize,
		ping:       1,
		pong:       -1,
		m:          0,
		action:     PassThrough,
	}
}

func (socket *MisraSocket) Listen() {
	listener, err := net.Listen("tcp", "localhost:"+socket.listenPort)
	if err != nil {
		fmt.Printf("Failed to listen on port: %s, %s", socket.listenPort, err)
	}
	defer listener.Close()
	fmt.Printf("Listening on port %s", socket.listenPort)

	for {
		conn, err := listener.Accept()
		if err != nil {
			fmt.Printf("Failed to accept connection on port: %s, %s", socket.listenPort, err)
			continue
		}
		socket.handleConnection(conn)
	}
}

func (socket *MisraSocket) handleConnection(conn net.Conn) {
	defer conn.Close()

	scanner := bufio.NewScanner(conn)
	for scanner.Scan() {
		input := scanner.Text()

		pingPong, err := strconv.Atoi(input)
		if err != nil {
			fmt.Printf("Failed to convert input to int: %s, %s", input, err)
		}
		socket.incomingMessage(pingPong)
	}
}

func (socket *MisraSocket) incomingMessage(pingPong int) {
	if socket.m == pingPong {
		socket.regenerate(pingPong)
		return
	}
	if pingPong > 0 {
		socket.hasPing = true
	}
	if pingPong < 0 {
		socket.hasPong = true
	}
	if socket.hasPing && socket.hasPong {
		socket.incarnate(pingPong)
	}

}

func (socket *MisraSocket) regenerate(pingPong int) {
	socket.ping = util.Abs(pingPong)
	socket.pong = -socket.ping
	socket.incarnate(socket.ping)
}

func (socket *MisraSocket) incarnate(pingPong int) {
	socket.ping = (util.Abs(pingPong) + 1) % socket.ringSize
	socket.pong = -socket.ping
	socket.action = Incarnated
}

func (socket *MisraSocket) send(pingPong int) {
	address := fmt.Sprintf("%s:%s", socket.sendIp, socket.sendPort)
	conn, err := net.Dial("tcp", address)
	defer conn.Close()
	if err != nil {
		fmt.Printf("Failed to connect to server: %s, %s", address, err)
	}
	switch socket.action {
	case PassThrough:
		switch {
		case pingPong > 0:
			socket.hasPing = false
		case pingPong < 0:
			socket.hasPong = false
		}
		conn.Write([]byte(fmt.Sprintf("%d", pingPong)))
	case Incarnated:
		conn.Write([]byte(fmt.Sprintf("%d", socket.ping)))
		conn.Write([]byte(fmt.Sprintf("%d", socket.pong)))
		socket.hasPing = false
	case CriticalSection:
		return
	}
}
