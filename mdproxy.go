package main

import (
	_ "database/sql"
	"fmt"
	"net"
	"strings"
	"time"
	_ "time"

	_ "github.com/go-sql-driver/mysql"
)

//单独处理客户端的请求
func clientHandle(conn net.Conn) {
	//设置当客户端3分钟内无数据请求时，自动关闭conn
	conn.SetReadDeadline(time.Now().Add(time.Minute * 3))
	defer conn.Close()

	//循环的处理客户的请求
	for {
		data := make([]byte, 256)
		//从conn中读取数据
		n, err := conn.Read(data)
		//如果读取数据大小为0或出错则退出
		if n == 0 || err != nil {
			break
		}
		//去掉两端空白字符
		cmd := strings.TrimSpace(string(data[0:n]))
		//发送给客户端的数据
		rep := ""
		if cmd == "string" {
			rep = "hello,client \r\n"
		} else if cmd == "time" {
			rep = time.Now().Format("2006-01-02 15:04:05")
		}
		//发送数据
		conn.Write([]byte(rep))
	}
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

		//接收用户的请求
		buf := make([]byte, 1024) //1024大小的缓冲区
		n, err1 := conn.Read(buf)
		if err1 != nil {
			fmt.Println("err1 = ", err1)
			return
		}

		fmt.Println("buf = ", string(buf[:n]))

		// defer conn.Close() //关闭当前用户链接
	}

	// tcpaddr, err := net.ResolveTCPAddr("tcp", "127.0.0.1:3308")
	// tcplisten, err := net.ListenTCP("tcp", tcpaddr)
	// fmt.Println("err", err)
	// for {
	// 	conn, err3 := tcplisten.Accept()
	// 	if err3 != nil {
	// 		continue
	// 	}
	// 	go clientHandle(conn)
	// }

	// db, _ := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/dd")
	// db.SetMaxOpenConns(10)
	// db.SetMaxIdleConns(5)
	// //连接数据库查询
	// for i := 0; i < 100; i++ {
	// 	go func(i int) {
	// 		mSql := "select * from dd_logs"
	// 		rows, _ := db.Query(mSql)
	// 		rows.Close() //这里如果不释放连接到池里，执行5次后其他并发就会阻塞
	// 		fmt.Println("第 ", i)
	// 	}(i)
	// }

	// for {
	// 	time.Sleep(time.Second)
	// }
}
