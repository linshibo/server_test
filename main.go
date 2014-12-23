package main

import (
	"sanguo/base/log"
	"fmt"
	"runtime"
	"math/rand"
	"time"
	"net"
	"os"
)

type GameServer struct {
	Host   string
}


func (server *GameServer) Start() {
	// load system data
	log.Debug("/*************************SREVER START********************************/")

	tcpAddr, err := net.ResolveTCPAddr("tcp4", server.Host)
	if err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
	go func(){
		for{
			select {
			case <-time.After(30*time.Second):
				LookUp("read memstats")
			}
		}
	}()
	listener, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		log.Error(err.Error())
		os.Exit(-1)
	}
	log.Debug("/*************************SERVER SUCC********************************/")
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			continue
		}
		log.Debug("Accept a new connection ", conn.RemoteAddr())
		go handleClient(conn)
	}
}

func handleClient(conn *net.TCPConn) {
	sess := NewSession(conn)
	sess.Start()
}

func main() {
	rand.Seed(time.Now().Unix())

	runtime.GOMAXPROCS(runtime.NumCPU())

	log.SetLevel(0)
	
	filew := log.NewFileWriter("log", true)
	err := filew.StartLogger()
	if err != nil {
		fmt.Println("Failed start log",err)
		return
	}

	var server GameServer
	server.Host = "127.0.0.1:9999"
	server.Start()
}

