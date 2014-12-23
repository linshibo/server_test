package main

import (
	"io"
	"encoding/binary"
	"encoding/json"
	"github.com/felixge/tcpkeepalive"
	"net"
	"sanguo/base/log"
	"strings"
	"sync"
	"time"
	"bufferedManager"
)

var b *bufferedManager.BufferManager
func init(){
	b = bufferedManager.NewBufferManager(1024)
}

type Session struct {
	ip   net.IP

	conn *net.TCPConn //the tcp connection from client

	recvChan      chan *bufferedManager.Token //data from client
	ipcChan       chan []byte //internet process connection message
	sendChan      chan []byte //data to  client
	closeNotiChan chan bool   //结束标记 此session 的其他goroutine 可以select 这个管道来判断是否关闭
	exitChan      chan bool   //dispatch

	ok   bool
	lock sync.Mutex

	uid  int
}


func NewSession(connection *net.TCPConn) (sess *Session) {
	var client Session

	client.ip = net.ParseIP(strings.Split(connection.RemoteAddr().String(), ":")[0])
	client.conn = connection

	client.recvChan = make(chan *bufferedManager.Token, 1024)
	client.ipcChan = make(chan []byte, 1024)
	client.closeNotiChan = make(chan bool)
	client.exitChan = make(chan bool)
	client.ok = true

	log.Debug("New Connection", &client)

	kaConn, err := tcpkeepalive.EnableKeepAlive(connection)
	if err != nil {
		log.Debug("EnableKeepAlive err ", err)
	} else {
		kaConn.SetKeepAliveIdle(120 * time.Second)
		kaConn.SetKeepAliveCount(4)
		kaConn.SetKeepAliveInterval(5 * time.Second)
	}
	return &client
}

func (sess *Session) Kickout() {
	sess.Close()
	log.Debug(sess, "Kickout")
	t:=time.Now()
	select {
	case <-sess.exitChan:
	}
	log.Trace(sess, "Kickout Succ:", time.Now().Sub(t))
}

func foo2(i []byte){
	log.Trace(i)
}
func foo(i interface{}){
	back_json , _ := json.Marshal(i)
	log.Trace(back_json)
}
func handle(){
	type data struct {
		a int
		a1 int
		a2 int
		a3 int
		a4 int
		a5 int
		a6 int
		a7 int
		a8 int
		a9 int
		a10 int
		a11 int
	}
	var msg data
	back_json , _ := json.Marshal(msg)
	foo2(back_json )
}

func (sess *Session) handleSend(ch chan []byte) {
	conn := sess.conn
	uid := sess.uid
	defer func(){
		log.Trace("Session Send Exit", sess, uid)
		sess.Close()
	}()
	for {
		select {
		case msg, ok := <-ch:
			if !ok {
				return
			}
			_, err := conn.Write(msg)
			if err != nil {
				log.Error("conn write error", err)
				return
			}
			log.Trace("sssend",binary.LittleEndian.Uint16(msg[2:4]), len(ch))
		case _ = <-sess.closeNotiChan:
			return
		}
	}
}

func (sess *Session) Close() {
	sess.lock.Lock()
	if sess.ok {
		sess.ok = false
		close(sess.closeNotiChan)
		sess.conn.Close()
		log.Trace("Sess Close Succ", sess, sess.uid)
	}
	sess.lock.Unlock()
}

func (sess *Session) handleRecv() {
	defer func(){
		if err := recover(); err != nil {
			log.Critical("Panic", err)
		}
		log.Trace("Session Recv Exit", sess, sess.uid)
		sess.Close()
	}()
	ch := sess.recvChan
	header := make([]byte, 2)
	for {
		/**block until recieve len(header)**/
		n, err := io.ReadFull(sess.conn, header)
		if n == 0 && err == io.EOF {
			//Opposite socket is closed
			log.Warn("Socket Read EOF And Close", sess)
			break
		} else if err != nil {
			//Sth wrong with this socket
			log.Warn("Socket Wrong:", err)
			break
		}
		size := binary.LittleEndian.Uint16(header) + 4
		//data := make([]byte, size)
		t := b.GetToken(int(size)) 
		n, err = io.ReadFull(sess.conn, t.Data)
		if n == 0 && err == io.EOF {
			log.Warn("Socket Read EOF And Close", sess)
			break
		} else if err != nil {
			log.Warn("Socket Wrong:", err)
			break
		}
		ch <- t //send data to Client to process
	}
}

func (sess *Session) handleDispatch() {
	defer func(){
		log.Trace("Session Dispatch Exit",  sess, sess.uid)
		sess.Close()
	}()
	for {
		select {
		case msg, _ := <-sess.recvChan:
			//log.Debug("msg", msg)
			msg.Return()
			//handle()
			//if !UserRequestProxy(sess, msg) {
				//log.Error("Something wrong in process user msg")
				//return 
			//}
			sess.SendDirectly("helloworldhelloworldhelloworldhelloworldhelloworldhelloworldhelloworld", 1)
		case msg, _:= <-sess.ipcChan:
			log.Debug("msg", msg)
			//if !IPCRequestProxy(sess, &msg) {
				//log.Error("Something wrong in process IPC msg")
				//return
			//}
		case <-sess.closeNotiChan:
				return
		}
	}
}

func (sess *Session) Start() {
	defer func() {
		if err := recover(); err != nil {
			log.Critical("Panic", err)
		}
	}()
	go sess.handleRecv()

	sess.handleDispatch()

	close(sess.ipcChan)
	close(sess.recvChan)
	close(sess.exitChan)
	log.Warn("Session Start Exit", sess, sess.uid)
}


func (sess *Session) SendDirectly(back interface{}, op int) bool {
	back_json, err := json.Marshal(back)
	if err != nil {
		log.Error("Can't encode json message ", err, back)
		return false
	}
	log.Debug(sess.uid, "OUT cmd:", op, string(back_json))
	_, err = sess.conn.Write(back_json)
	if err != nil {
		log.Error("send fail", err)
		return false
	}
	return true
}

