package main

import (
	"errSrv"
)

func main() {
	errSrv.Connect()
	errSrv.Run("localhost:8880")
}
