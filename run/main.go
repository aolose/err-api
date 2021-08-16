package main

import (
	"errSrv"
)

func main() {
	errSrv.Connect()
	db := errSrv.DB()
	post := &errSrv.Post{}
	post.SetPublic(errSrv.PublicPost{
		Title:   "title000",
		Content: "content0",
		Slug:    "12",
		Author: errSrv.Author{
			Name: "Tom",
		},
	})
	db.Where(post).FirstOrCreate(post)
	errSrv.Run("localhost:8880")
}
