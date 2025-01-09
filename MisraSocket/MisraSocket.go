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
}

func NewMisraSocket() *MisraSocket {
	return &MisraSocket{
		sendConnection: nil,
		listenConn:     nil,
		m:              0,
		ping:           1,
		pong:           -1,
		storedToken:    NONE,
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
	var sendedType string
	switch tokenType {
	case PING_TOKEN:
		sendedType = "PING"
	case PONG_TOKEN:
		sendedType = "PONG"
	}
	util.LogSuccess("Token send: %s", sendedType)
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
	scanner := bufio.NewScanner(*socket.listenConn)
	for {
		err := (*socket.listenConn).SetReadDeadline(time.Time{})
		if err != nil {
			util.LogError(err.Error())
		}
		switch socket.storedToken {
		case NONE:
			scanner.Scan()
			token, err := strconv.ParseInt(scanner.Text(), 10, 64)
			if err != nil {
				util.LogError(err.Error())
				continue
			}
			socket.receiveToken(token)
			continue
		case PING:
			util.LogSuccess("Ping token acquired, entering critical section")
			time.Sleep(1 * time.Second)
			util.LogSuccess("Leaving critical section")
			err := (*socket.listenConn).SetReadDeadline(time.Now().Add(100 * time.Millisecond))
			if err != nil {
				util.LogError(err.Error())
			}
			if scanner.Scan() {
				token, err := strconv.ParseInt(scanner.Text(), 10, 64)
				if err != nil {
					util.LogError(err.Error())
					continue
				}
				socket.receiveToken(token)
				continue
			} else {
				socket.send(PING_TOKEN)
			}

		case PONG:
			socket.send(PONG_TOKEN)
		case BOTH:
			util.LogInfo("Both token acquired, incarnating")
			socket.incarnate()
			util.LogInfo("New token values PING: %d, PONG: %d", socket.ping, socket.pong)
			socket.send(PONG_TOKEN)
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
		}
		if socket.m < 0 {
			util.LogWarn("Ping token lost, proceed with regeneration")
			socket.regenerateTokens()
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
	}
}

func (socket *MisraSocket) regenerateTokens() {
	socket.incarnate()
	socket.storedToken = BOTH
}

func (socket *MisraSocket) incarnate() {
	socket.ping++
	socket.pong = -socket.ping
}
