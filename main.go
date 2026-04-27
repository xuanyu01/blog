/*
程序启动入口。
*/
package main

import (
	"blog/app"
	"log"
)

// main 创建应用并启动服务。
func main() {
	application, err := app.New()
	if err != nil {
		log.Fatal(err)
	}

	if err := application.Run(); err != nil {
		log.Fatal(err)
	}
}
