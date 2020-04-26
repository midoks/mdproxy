package main

import (
	"database/sql"
	"fmt"
	_ "github.com/go-sql-driver/mysql"
	"sync"
	"testing"
	"time"
)

func MySQL_PoolSetting_Test(maxConn, idleConn, queryNum int) int64 {
	start := time.Now().UnixNano()
	db, _ := sql.Open("mysql", "root:root@tcp(127.0.0.1:3306)/mysql")

	defer db.Close()

	db.SetMaxOpenConns(maxConn)
	db.SetMaxIdleConns(idleConn)
	var wg sync.WaitGroup
	for i := 0; i < queryNum; i++ {
		wg.Add(1)
		go func(i int) {
			mSql := "SELECT `Host`,`User`,`Plugin` FROM `mysql`.`user` limit 1"

			rows, _ := db.Query(mSql)
			rows.Close()

			wg.Done()
		}(i)
	}
	wg.Wait()

	end := time.Now().UnixNano()
	time.Sleep(1000)
	return (end - start) / int64(time.Millisecond)
}

func Test_MySQL_Press(t *testing.T) {
	var cos int64
	for i := 1; i < 100; i++ {
		for j := 0; j < 4; j++ {
			cos = MySQL_PoolSetting_Test(i, i, 10000)
			fmt.Println("conn:", i, " cos:", cos, "ms")
		}
	}
}

// go test -v
// go test -bench=. -benchtime=3s  -run=none
func Benchmark_MySQL_Press(b *testing.B) {

	var cos int64

	for i := 0; i < b.N; i++ {
		for i := 1; i < 100; i++ {
			for j := 0; j < 4; j++ {
				cos = MySQL_PoolSetting_Test(i, i, 10000)
				fmt.Println("conn:", i, " cos:", cos, "ms")
			}
		}
	}
}
