package MisraSocket

import (
	"MisraSockets/util"
	"bufio"
	"fmt"
	"net"
	"os"
	"strconv"
	"sync"
	"time"
)

type StoredToken int

const (
	NONE StoredToken = iota
	PING
	PONG
	BOTH
)

type TokenType int

const (
	PING_TOKEN TokenType = iota
	PONG_TOKEN
)

type MisraSocket struct {
	sendConnection *net.Conn
	listenConn     *net.Conn
	m              int64
	ping           int64
	pong           int64
	storedToken    StoredToken
	readChannel    chan int64
}

func NewMisraSocket() *MisraSocket {
	return &MisraSocket{
		sendConnection: nil,
		listenConn:     nil,
		m:              0,
		ping:           1,
		pong:           -1,
		storedToken:    NONE,
		readChannel:    make(chan int64, 2),
	}
}

func (socket *MisraSocket) send(tokenType TokenType) {
	switch tokenType {
	case PING_TOKEN:
		_, err := fmt.Fprintln(*socket.sendConnection, socket.ping)
		if err != nil {
			util.LogError(err.Error())
			return
		}
		socket.m = socket.ping
		if socket.storedToken == PING {
			socket.storedToken = NONE
		}
		if socket.storedToken == BOTH {
			socket.storedToken = PONG
		}
	case PONG_TOKEN:
		_, err := fmt.Fprintln(*socket.sendConnection, socket.pong)
		if err != nil {
			util.LogError(err.Error())
			return
		}
		socket.m = socket.pong
		if socket.storedToken == PONG {
			socket.storedToken = NONE
		}
		if socket.storedToken == BOTH {
			socket.storedToken = PING
		}
	}
	var typeToSend string
	switch tokenType {
	case PING_TOKEN:
		typeToSend = "PING"
	case PONG_TOKEN:
		typeToSend = "PONG"
	}
	util.LogSuccess("Token send: %s", typeToSend)
}

func (socket *MisraSocket) InitOutgoingConnection(sendAddress string) {
	util.LogInfo("Trying to connect to %s", sendAddress)
	var err error
	var conn net.Conn
	for {
		conn, err = net.Dial("tcp", sendAddress)
		if err != nil {
			util.LogError("Error connecting to port %s. Retrying...", sendAddress)
			time.Sleep(1 * time.Second)
			continue
		}
		util.LogInfo("Successfully connected to port %s", sendAddress)
		break
	}
	socket.sendConnection = &conn
}

func (socket *MisraSocket) SendInitMessage() {
	socket.storedToken = BOTH
	socket.send(PING_TOKEN)
	socket.send(PONG_TOKEN)
}

func (socket *MisraSocket) Listen(listenPort string, wg *sync.WaitGroup) {
	defer wg.Done()
	util.LogInfo("Listening on port %s", listenPort)
	ln, err := net.Listen("tcp", ":"+listenPort)
	if err != nil {
		util.LogError(err.Error())
		os.Exit(1)
	}

	conn, err := ln.Accept()
	if err != nil {
		util.LogError(err.Error())
		os.Exit(1)
	}
	util.LogSuccess("Connection established with: %s", conn.RemoteAddr().String())
	socket.listenConn = &conn
}

func (socket *MisraSocket) HandleMessages() {
	go socket.readFromConnection()
	for {
		switch socket.storedToken {
		case NONE:
			select {
			case token := <-socket.readChannel:
				socket.receiveToken(token)
			}
		case PING:
			util.LogSuccess("Ping token acquired, entering critical section")
			time.Sleep(1 * time.Second)
			util.LogSuccess("Leaving critical section")
			select {
			case token := <-socket.readChannel:
				socket.receiveToken(token)
				continue
			default:
				socket.send(PING_TOKEN)
			}
		case PONG:
			socket.send(PONG_TOKEN)
		case BOTH:
			util.LogInfo("Both token acquired, incarnating")
			socket.incarnate()
			util.LogInfo("New token values PING: %d, PONG: %d", socket.ping, socket.pong)
			socket.send(PING_TOKEN)
			socket.send(PONG_TOKEN)
		}

	}
}

func (socket *MisraSocket) receiveToken(token int64) {
	if util.Abs(token) < util.Abs(socket.m) {
		util.LogInfo("Old token, ignoring")
		return
	}
	if token == socket.m {
		if socket.m > 0 {
			util.LogWarn("Pong token lost, proceed with regeneration")
			socket.regenerateTokens()
			return
		}
		if socket.m < 0 {
			util.LogWarn("Ping token lost, proceed with regeneration")
			socket.regenerateTokens()
			return
		}
	}
	if token > 0 {
		socket.ping = token
		socket.pong = -socket.ping
		switch socket.storedToken {
		case NONE:
			socket.storedToken = PING
		case PONG:
			socket.storedToken = BOTH
		case PING, BOTH:
			util.LogError("Something went very wrong!!!")
		}
		return
	}
	if token < 0 {
		socket.pong = token
		socket.ping = util.Abs(token)
		switch socket.storedToken {
		case NONE:
			socket.storedToken = PONG
		case PING:
			socket.storedToken = BOTH
		case PONG, BOTH:
			util.LogError("Something went very wrong!!!")
		}
		return
	}
}

func (socket *MisraSocket) regenerateTokens() {
	socket.storedToken = BOTH
}

func (socket *MisraSocket) incarnate() {
	socket.ping++
	socket.pong = -socket.ping
}

func (socket *MisraSocket) readFromConnection() {
	scanner := bufio.NewScanner(*socket.listenConn)
	for {
		scanner.Scan()
		token, err := strconv.ParseInt(scanner.Text(), 10, 64)
		if err != nil {
			util.LogError(err.Error())
			continue
		}
		socket.readChannel <- token
		continue
	}
}
