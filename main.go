package main

import (
	"errSrv"
)

func main() {
	errSrv.Connect()
	errSrv.Run("127.0.0.1:8880")
}
