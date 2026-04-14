/*
该文件是程序的启动入口，负责创建应用并启动服务
*/
package main

import (
	"blog/app"
	"log"
)

// main 是程序的主入口函数
func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
