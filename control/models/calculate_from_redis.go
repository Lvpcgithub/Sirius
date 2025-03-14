package models

import (
	"control/config"
	"database/sql"
	"encoding/json"
	"fmt"
	"log"
	"time"

	"github.com/gomodule/redigo/redis"
)

func CalculateAvgDelay(conn redis.Conn,db *sql.DB, ip1 string, ip2 string){
	var totalDelay float64
	totalDelay = 0
	key := fmt.Sprintf("%s:%s", ip1, ip2)
	//fmt.Println("key:", key)
	// 获取最新的10条数据
	values, err := redis.Values(conn.Do("LRANGE", key, -10, -1))
	if err != nil {
		log.Fatalf("Failed to retrieve data from Redis: %v", err)
	}
	// 如果没有数据，返回0
	if len(values) == 0 {
		log.Printf("No data found for key: %s", key)
	}

	// 解析每条数据并累加延迟
	for _, value := range values {
		var result config.ProbeResult
		err := json.Unmarshal(value.([]byte), &result)
		if err != nil {
			log.Printf("Failed to parse Redis value: %v", err)
			continue // 跳过无法解析的数据
		}
		totalDelay += float64(result.Delay)
	}
	// 计算平均延迟
	avgDelay := totalDelay / float64(len(values))
	// 插入数据库
	InsertLinkInfo(db,ip1,ip2,avgDelay,time.Now().Format("2006-01-02 15:04:05"))
}