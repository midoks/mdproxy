package main

// import (
// 	"bufio"
// 	"errors"
// 	"fmt"
// 	"io"
// 	"log"
// 	"math/rand"
// 	"net"
// 	"runtime"
// 	_ "strings"
// 	"sync/atomic"
// 	"time"

// 	. "github.com/midoks/mdproxy/mysql"
// )

// type Server struct {
// 	listener net.Listener
// }

// func NewServer() (*Server, error) {
// 	s := new(Server)

// 	s.listener, err = net.Listen(netProto, s.addr)

// 	if err != nil {
// 		return nil, err
// 	}
// 	return s, nil
// }

// func (s *Server) Run() error {
// 	conn, err := s.listener.Accept()
// 	if err != nil {
// 		log.Error("accept error %s", err.Error())
// 		continue
// 	}

// 	go s.onConn(conn)
// }

// func (s *Server) onConn(conn net.Conn) error {

// }
