package models

import (
	"database/sql"
	"fmt"
)

// 查询ip列表
func QueryIp(db *sql.DB) ([]string, error) {
	rows, err := db.Query("SELECT DISTINCT ip FROM system_info")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	var ips []string
	for rows.Next() {
		var ip string
		if err := rows.Scan(&ip); err != nil {
			fmt.Println("扫描行出错：",err)
			return nil, err
		}
		ips = append(ips, ip)
	}

	if err := rows.Err(); err != nil {
		return nil, err
	}

	return ips, nil
}
