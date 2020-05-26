package main

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"log"
	"net"
	"runtime"
	_ "strings"
	"time"

	// "crypto/sha1"
	"math/rand"
	"sync/atomic"

	. "github.com/midoks/mdproxy/mysql"
)

const (
	MinProtocolVersion byte   = 10
	MaxPayloadLen      int    = 1<<24 - 1
	TimeFormat         string = "2006-01-02 15:04:05"
	ServerVersion      string = "mdproxy 0.1"
)

var (
	ErrBadConn       = errors.New("connection was bad")
	ErrMalformPacket = errors.New("Malform packet error")

	ErrTxDone = errors.New("sql: Transaction has already been committed or rolled back")
)

var DEFAULT_CAPABILITY uint32 = CLIENT_LONG_PASSWORD | CLIENT_LONG_FLAG |
	CLIENT_CONNECT_WITH_DB | CLIENT_PROTOCOL_41 |
	CLIENT_TRANSACTIONS | CLIENT_SECURE_CONNECTION

func RandomBuf(size int) []byte {
	buf := make([]byte, size)
	rand.Seed(time.Now().UTC().UnixNano())
	for i := 0; i < size; i++ {
		buf[i] = byte(rand.Intn(127))
		if buf[i] == 0 || buf[i] == byte('$') {
			buf[i]++
		}
	}
	return buf
}

var baseConnId uint32 = 10000

type Conn struct {
	conn net.Conn
}

func writeInitialHandshake(conn net.Conn) error {

	salt := RandomBuf(20)
	connectionId := atomic.AddUint32(&baseConnId, 1)

	fmt.Println(connectionId)

	data := make([]byte, 4, 128)

	//min version 10
	data = append(data, 10)

	//server version[00]
	data = append(data, ServerVersion...)
	data = append(data, 0)

	//connection id
	data = append(data, byte(connectionId), byte(connectionId>>8), byte(connectionId>>16), byte(connectionId>>24))

	//auth-plugin-data-part-1
	data = append(data, salt[0:8]...)

	//filter [00]
	data = append(data, 0)

	//capability flag lower 2 bytes, using default capability here
	data = append(data, byte(DEFAULT_CAPABILITY), byte(DEFAULT_CAPABILITY>>8))

	//charset, utf-8 default
	data = append(data, uint8(DEFAULT_COLLATION_ID))

	//status
	data = append(data, byte(SERVER_STATUS_AUTOCOMMIT), byte(SERVER_STATUS_AUTOCOMMIT>>8))

	//below 13 byte may not be used
	//capability flag upper 2 bytes, using default capability here
	data = append(data, byte(DEFAULT_CAPABILITY>>16), byte(DEFAULT_CAPABILITY>>24))

	//filter [0x15], for wireshark dump, value is 0x15
	data = append(data, 0x15)

	//reserved 10 [00]
	data = append(data, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0)

	//auth-plugin-data-part-2
	data = append(data, salt[8:]...)

	//filter [00]
	data = append(data, 0)

	length := len(data) - 4
	data[0] = byte(length)
	data[1] = byte(length >> 8)
	data[2] = byte(length >> 16)
	data[3] = 0

	fmt.Println("length", length)
	fmt.Println(data, string(data))

	conn.Write(data)

	return nil
}

func readHandshakeResponse(conn net.Conn) ([]byte, error) {
	header := []byte{0, 0, 0, 0}

	if _, err := io.ReadFull(conn, header); err != nil {
		fmt.Println("readHandshakeResponse", err)
	}

	length := int(uint32(header[0]) | uint32(header[1])<<8 | uint32(header[2])<<16)
	if length < 1 {
		fmt.Printf("invalid payload length %d", length)
	}

	sequence := uint8(header[3])

	fmt.Println("length", length)
	fmt.Println("sequence", sequence)

	// if sequence != p.Sequence {
	// 	fmt.Printf("invalid sequence %d != %d", sequence, p.Sequence)
	// }

	data := make([]byte, length)
	if _, err := io.ReadFull(conn, data); err != nil {
		fmt.Println("readHandshakeResponse", err)
	} else {
		if length < MaxPayloadLen {
			return data, nil
		}

		var buf []byte
		buf, err = readHandshakeResponse(conn)
		if err != nil {
			return nil, ErrBadConn
		} else {
			return append(data, buf...), nil
		}
	}
	return nil, ErrBadConn
}

//单独处理客户端的请求
func clientHandle(conn net.Conn) {

	defer func() {
		if err := recover(); err != nil {
			const size = 4096
			buf := make([]byte, size)
			buf = buf[:runtime.Stack(buf, false)]
			log.Println("onConn panic: %v\n%s", err, buf)
		}

		conn.Close()
	}()
	fmt.Println("clientHandle start")
	//设置当客户端3分钟内无数据请求时，自动关闭conn
	conn.SetReadDeadline(time.Now().Add(time.Minute * 3))

	// data := make([]byte, 4096)
	//循环的处理客户的请求
	for {
		//从conn中读取数据
		// n, err := conn.Read(data)
		// //如果读取数据大小为0或出错则退出
		// if n == 0 || err != nil {
		// 	fmt.Println("eerr")
		// 	conn.Close()
		// 	break
		// }

		fmt.Println("for in")

		rb := bufio.NewReaderSize(conn, 1024)
		header := []byte{0, 0, 0, 0}
		vv, err := io.ReadFull(rb, header)
		if err != nil {
			fmt.Println("err:", err, vv)
			break
		}
		fmt.Println("vvv:", vv, err)

		//去掉两端空白字符
		// cmd := strings.TrimSpace(string(data[0:n]))
		// //发送给客户端的数据
		// rep := ""
		// if cmd == "string" {
		// 	rep = "hello,client \r\n"
		// } else if cmd == "time" {
		// 	rep = time.Now().Format("2006-01-02 15:04:05")
		// }
		// fmt.Println(cmd, rep)
		//发送数据
		// conn.Write([]byte(rep))
	}
	// conn.Close()
}

func main() {

	listener, err := net.Listen("tcp", "127.0.0.1:3308")
	if err != nil {
		fmt.Println("err = ", err)
		return
	}

	defer listener.Close()

	for {
		//阻塞等待用户链接
		conn, err := listener.Accept()
		if err != nil {
			fmt.Println("err = ", err)
			return
		}

		writeInitialHandshake(conn)
		data, err := readHandshakeResponse(conn)
		fmt.Println("data1:", string(data), err)

		data2 := make([]byte, 4, 128)
		data2 = append(data2, 7, 0, 0, 2, 0, 0, 0, 2, 0, 0, 0)
		conn.Write(data2)

		fmt.Println("data2 ok:", data2)

		go clientHandle(conn)
	}
}
