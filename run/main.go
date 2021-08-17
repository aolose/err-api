package main

import (
	"errSrv"
	"strconv"
)

func main() {
	errSrv.Connect()
	db := errSrv.DB()
	for i := 0; i < 30; i++ {
		post := &errSrv.Post{}
		post.Status = 1
		pf := strconv.Itoa(i)
		post.SetPublic(errSrv.PublicPost{
			Title:   "title-" + pf,
			Content: "content-" + pf,
			Slug:    "post" + pf,
			Author: errSrv.Author{
				Name: "Tom",
			},
		})
		db.Where(post).FirstOrCreate(post)
	}
	errSrv.Run("localhost:8880")
}
