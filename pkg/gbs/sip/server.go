package sip

import (
	"bufio"
	"bytes"
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
)

var bufferSize uint16 = 65535 - 20 - 8 // IPv4 max size - IPv4 Header size - UDP Header size

// RequestHandler RequestHandler
// type RequestHandler func(req *Request, tx *Transaction)

// Server Server
type Server struct {
	udpaddr net.Addr
	conn    Connection

	txs *transacionts

	hmu   *sync.RWMutex
	route map[string][]HandlerFunc

	port *Port
	host net.IP

	tcpPort     *Port
	tcpListener *net.TCPListener

	tcpaddr net.Addr
}

// NewServer NewServer
func NewServer() *Server {
	activeTX = &transacionts{txs: map[string]*Transaction{}, rwm: &sync.RWMutex{}}
	srv := &Server{
		hmu:   &sync.RWMutex{},
		txs:   activeTX,
		route: make(map[string][]HandlerFunc),
	}
	return srv
}

func (s *Server) addRoute(method string, pattern string, handler ...HandlerFunc) {
	s.hmu.Lock()
	defer s.hmu.Unlock()
	key := method + "-" + pattern
	s.route[key] = handler
}

func (s *Server) Register(handler ...HandlerFunc) {
	s.addRoute(MethodRegister, "", handler...)
}

func (s *Server) Message(handler ...HandlerFunc) {
	s.addRoute(MethodMessage, "", handler...)
}

func (s *Server) getTX(key string) *Transaction {
	return s.txs.getTX(key)
}

func (s *Server) mustTX(key string) *Transaction {
	tx := s.txs.getTX(key)
	if tx == nil {
		tx = s.txs.newTX(key, s.conn)
	}
	return tx
}

// ListenUDPServer ListenUDPServer
func (s *Server) ListenUDPServer(addr string) {
	udpaddr, err := net.ResolveUDPAddr("udp", addr)
	if err != nil {
		panic(fmt.Errorf("net.ResolveUDPAddr err[%w]", err))
	}
	s.port = NewPort(udpaddr.Port)
	s.host, err = ResolveSelfIP()
	if err != nil {
		panic(fmt.Errorf("net.ListenUDP resolveip err[%w]", err))
	}
	udp, err := net.ListenUDP("udp", udpaddr)
	if err != nil {
		panic(fmt.Errorf("net.ListenUDP err[%w]", err))
	}
	s.conn = newUDPConnection(udp)
	var (
		raddr net.Addr
		num   int
	)
	buf := make([]byte, bufferSize)
	parser := newParser()
	defer parser.stop()
	go s.handlerListen(parser.out)
	for {
		num, raddr, err = s.conn.ReadFrom(buf)
		if err != nil {
			// logrus.Errorln("udp.ReadFromUDP err", err)
			continue
		}
		parser.in <- newPacket(append([]byte{}, buf[:num]...), raddr, s.conn)
	}
}

// ListenTCPServer 启动 TCP 服务器并监听指定地址。
func (s *Server) ListenTCPServer(ctx context.Context, addr string) {
	// 解析传入的地址为 TCP 地址
	tcpaddr, err := net.ResolveTCPAddr("tcp", addr)
	// 如果解析地址失败，则抛出错误
	if err != nil {
		panic(fmt.Errorf("net.ResolveUDPAddr err[%w]", err))
	}
	// 保存解析后的 TCP 地址到服务器结构体
	s.tcpaddr = tcpaddr
	// 创建新的端口实例并保存到服务器结构体
	s.tcpPort = NewPort(tcpaddr.Port)

	// 创建 TCP 监听器
	tcp, err := net.ListenTCP("tcp", tcpaddr)
	// 如果创建监听器失败，则抛出错误
	if err != nil {
		panic(fmt.Errorf("net.ListenUDP err[%w]", err))
	}
	// 确保在方法退出时关闭 TCP 监听器
	// 当这个关闭时 所有的设备的socket都会被关闭
	// defer tcp.Close()
	// 保存 TCP 监听器到服务器结构体
	s.tcpListener = tcp
	// 无限循环接受连接

	for {
		select {
		case <-ctx.Done():
			slog.Info("ListenTCPServer Has Been Exits")
			return
		default:
			conn, err := tcp.Accept()
			if err != nil {
				slog.Error("net.ListenTCP", "err", err, "addr", addr)
				return
			}
			go s.ProcessTcpConn(conn)
		}
	}
}

// ProcessTcpConn 处理传入的 TCP 连接。
func (s *Server) ProcessTcpConn(conn net.Conn) {
	// 确保在方法退出时关闭连接
	defer conn.Close() // 关闭连接
	// 创建一个新的缓冲读取器，用于从连接中读取数据
	reader := bufio.NewReader(conn)
	// lenBuf := make([]byte, 2)
	// 创建一个新的 TCP 连接实例
	c := NewTCPConnection(conn)

	parser := newParser()
	defer parser.stop()
	go s.handlerListen(parser.out)

	for {
		// 初始化一个缓冲区，用于存储读取的数据
		var buffer bytes.Buffer
		// 初始化 body 长度
		bodyLen := 0
		for {
			// 读取一行数据，以 '\n' 为结束符
			line, err := reader.ReadBytes('\n')
			// 如果读取过程中出错，则退出方法
			if err != nil {
				// logrus.Debugln("tcp conn read error:", err)
				return
			}
			// 将读取的数据写入缓冲区
			buffer.Write(line)
			// 如果读取到的行长度小于等于2且 body 长度小于等于0，则跳出循环
			if len(line) <= 2 {
				if bodyLen <= 0 {
					break
				}

				// 读取 body 数据
				// read body
				bodyBuf := make([]byte, bodyLen)
				n, err := io.ReadFull(reader, bodyBuf)
				// 如果读取 body 数据时出错，则记录错误并退出循环
				if err != nil || n != bodyLen {
					slog.Error(`error while read full`, "err", err)
					// err process
				}
				// 将读取的 body 数据写入缓冲区
				buffer.Write(bodyBuf)
				break
			}

			// 如果读取到 "Content-Length" 头部，则解析 body 长度
			if strings.Contains(string(line), "Content-Length") {
				// 以: 对line做分割
				s := strings.Split(string(line), ":")
				value := strings.Trim(s[len(s)-1], " \r\n")
				bodyLen, err = strconv.Atoi(value)
				// 如果解析 "Content-Length" 头部失败，则记录错误并退出循环
				if err != nil {
					slog.Error("parse Content-Length failed")
					break
				}
			}
		}

		parser.in <- newPacket(buffer.Bytes(), conn.RemoteAddr(), c)
	}
}

func NewTCPConnection(baseConn net.Conn) Connection {
	conn := &connection{
		baseConn: baseConn,
		laddr:    baseConn.LocalAddr(),
		raddr:    baseConn.RemoteAddr(),
		logKey:   "tcpConnection",
	}
	return conn
}

func (s *Server) handlerListen(msgs chan Message) {
	var msg Message
	for {
		msg = <-msgs
		switch tmsg := msg.(type) {
		case *Request:
			req := tmsg
			req.SetDestination(s.udpaddr)
			s.handlerRequest(req)
		case *Response:
			resp := tmsg
			resp.SetDestination(s.udpaddr)
			s.handlerResponse(resp)
		default:
			// logrus.Errorln("undefind msg type,", tmsg, msg.String())
		}
	}
}

func (s *Server) handlerRequest(msg *Request) {
	tx := s.mustTX(getTXKey(msg))
	// logrus.Traceln("receive request from:", msg.Source(), ",method:", msg.Method(), "txKey:", tx.key, "message: \n", msg.String())
	s.hmu.RLock()
	handlers, ok := s.route[msg.Method()]
	s.hmu.RUnlock()
	if !ok {
		// logrus.Errorln("not found handler func,string:", msg.Method(), msg.String())
		go handlerMethodNotAllowed(msg, tx)
		return
	}

	ctx := newContext(msg, tx)
	ctx.handlers = handlers
	go ctx.Next()
}

func (s *Server) handlerResponse(msg *Response) {
	tx := s.getTX(getTXKey(msg))
	if tx == nil {
		// logrus.Infoln("not found tx. receive response from:", msg.Source(), "message: \n", msg.String())
	} else {
		// logrus.Traceln("receive response from:", msg.Source(), "txKey:", tx.key, "message: \n", msg.String())
		tx.receiveResponse(msg)
	}
}

// Request Request
func (s *Server) Request(req *Request) (*Transaction, error) {
	viaHop, ok := req.ViaHop()
	if !ok {
		return nil, fmt.Errorf("missing required 'Via' header")
	}
	viaHop.Host = s.host.String()
	viaHop.Port = s.port
	if viaHop.Params == nil {
		viaHop.Params = NewParams().Add("branch", String{Str: GenerateBranch()})
	}
	if !viaHop.Params.Has("rport") {
		viaHop.Params.Add("rport", nil)
	}

	tx := s.mustTX(getTXKey(req))
	return tx, tx.Request(req)
}

func handlerMethodNotAllowed(req *Request, tx *Transaction) {
	resp := NewResponseFromRequest("", req, http.StatusMethodNotAllowed, http.StatusText(http.StatusMethodNotAllowed), []byte{})
	tx.Respond(resp)
}
