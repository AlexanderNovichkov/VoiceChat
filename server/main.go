package main

import "server/app"

var PORT = ":8081"

func main() {
	application := app.App{}
	application.Run(PORT)
}
