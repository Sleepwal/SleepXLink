package network

import (
	"SleepXLink/global"
	"SleepXLink/logo"
	"fmt"
	"net"

	"SleepXLink/iface"
)

type Server struct {
	Name      string             `info:"服务器名称"`
	IPVersion string             `info:"IP版本"`
	IP        string             `info:"服务器监听的IP"`
	Port      int                `info:"服务器监听的端口"`
	MsgHandle iface.IMsgHandle   // server的消息管理模块
	ConnMgr   iface.IConnManager // server的连接管理模块

	//该Server的连接创建时Hook函数
	OnConnStart func(conn iface.IConnection)
	//该Server的连接断开时的Hook函数
	OnConnStop func(conn iface.IConnection)
}

/**
* 返回一个Server对象
**/
func NewServer() iface.IServer {
	s := &Server{
		Name:      global.SXL_CONFIG.Name,
		IPVersion: "tcp4",
		IP:        global.SXL_CONFIG.Host,
		Port:      global.SXL_CONFIG.Port,
		MsgHandle: NewMsgHandle(),
		ConnMgr:   NewConnManager(),
	}

	return s
}

func (s *Server) Start() {
	// 0.开启消息队列和worker工作池
	s.MsgHandle.StartWorkerPool()

	// 1.获取TCP的Addr
	addr, err := net.ResolveTCPAddr(s.IPVersion, fmt.Sprintf("%s:%d", s.IP, s.Port))
	if err != nil {
		fmt.Println("network resolve addr error: ", err)
	}

	// 2.监听地址
	listener, err := net.ListenTCP(s.IPVersion, addr)
	if err != nil {
		fmt.Println("listen TCP error: ", err)
		return
	}

	fmt.Println("start SleepXLink server success, ", s.Name, " is listening...")

	var cid uint32 = 0

	// 3.阻塞，等待客户端连接，处理客户端业务
	for {
		conn, err := listener.AcceptTCP()
		if err != nil {
			fmt.Println("accept error: ", err)
			continue
		}

		// 判断是否超过最大连接数量
		if s.ConnMgr.Len() >= global.SXL_CONFIG.MaxConn {
			//TODO 给客户端响应，超出最大连接的错误
			fmt.Println("-------> [Server]---", s.Name, " Server is full, too many connections! max conn num = ", global.SXL_CONFIG.MaxConn)
			conn.Close()
			continue
		}

		// 处理新连接，用链接模块处理
		handleConn := NewConnection(s, conn, cid, s.MsgHandle)
		cid++

		// 启动一个goroutine处理业务
		go handleConn.Start()
	}

}

func (s *Server) Stop() {
	// TODO 资源、状态、链接信息 停止或回收
	fmt.Println("[SleepXLink]---", s.Name, " Server Stop!")
	s.ConnMgr.ClearConn()
}

func (s *Server) Serve() {
	logo.InitLogo()
	global.SXL_LOG.Info("[SleepXLink]---" + global.SXL_CONFIG.Name + " Server Start!")
	fmt.Println("[SleepXLink]---Server IP:", global.SXL_CONFIG.Host)
	fmt.Println("[SleepXLink]---Server Port:", global.SXL_CONFIG.Port)
	fmt.Println("[SleepXLink]---Server Version:", global.SXL_CONFIG.Version,
		", Server MaxConn:", global.SXL_CONFIG.MaxConn,
		", Server MaxPackageSize:", global.SXL_CONFIG.MaxPackageSize)

	//Serve要处理其他业务，不能再Start中阻塞，故开启goroutine
	go s.Start()
	defer s.Stop()

	// TODO 启动服务器后的额外业务

	//阻塞
	select {}
}

func (s *Server) AddRouter(msgID uint32, router iface.IRouter) {
	s.MsgHandle.AddRouter(msgID, router)
	fmt.Println("AddRouter success!")
}

func (s *Server) GetConnMgr() iface.IConnManager {
	return s.ConnMgr
}

// SetOnConnStart 设置该Server的连接创建时Hook函数
func (s *Server) SetOnConnStart(hookFunc func(iface.IConnection)) {
	s.OnConnStart = hookFunc
}

// SetOnConnStop 设置该Server的连接断开时的Hook函数
func (s *Server) SetOnConnStop(hookFunc func(iface.IConnection)) {
	s.OnConnStop = hookFunc
}

// CallOnConnStart 调用连接OnConnStart Hook函数
func (s *Server) CallOnConnStart(conn iface.IConnection) {
	if s.OnConnStart != nil {
		fmt.Println("---> CallOnConnStart....")
		s.OnConnStart(conn)
	}
}

// CallOnConnStop 调用连接OnConnStop Hook函数
func (s *Server) CallOnConnStop(conn iface.IConnection) {
	if s.OnConnStop != nil {
		fmt.Println("---> CallOnConnStop....")
		s.OnConnStop(conn)
	}
}
