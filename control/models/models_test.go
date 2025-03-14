package models

import (
	"control/config"
	"control/dao"
	"encoding/json"
	"fmt"
	"log"
	"testing"
	"time"
)

// 测试查询IP列表
func TestQueryIp(t *testing.T) {
	// 连接到数据库
	db := dao.ConnectToDB()
	defer db.Close()
	ips,_ :=QueryIp(db)
	fmt.Println("Query result:", ips)
	// 例如：确保查询结果不为空
	if len(ips) == 0 {
		t.Error("Expected at least one IP, but got none")
	}
}
//测试redis计算方法
func TestCalculateAvgDelay(t *testing.T) {
	conn := dao.ConnRedis()
	db := dao.ConnectToDB()
	defer conn.Close()
	defer db.Close()
	var i int64
	for i = 1; i < 4; i++ {
		key := "192.168.1.1:192.168.2.2"
		value, _ := json.Marshal(config.ProbeResult{
			SourceIP:      "192.168.1.1",
			DestinationIP: "192.168.2.2",
			Delay:         i,
			Timestamp:     time.Now().Format("2006-01-02 15:04:05"),
		})
		_, lpushErr := conn.Do("LPUSH", key, value)
		if lpushErr != nil {
			log.Printf("Error storing result in redis: %v", lpushErr)
		}
	}
	CalculateAvgDelay(conn,db,"192.168.1.1","192.168.2.2")
}