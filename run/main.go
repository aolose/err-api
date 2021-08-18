package main

import (
	"errSrv"
	"strconv"
	"time"
)

func main() {
	errSrv.Connect()
	db := errSrv.DB()
	for i := 0; i < 60; i++ {
		post := &errSrv.Post{
			ID: uint(i + 1),
		}
		n := time.Now().Unix() * 1e3
		pf := strconv.Itoa(i)
		post.DraftTitle = "Draft-" + pf
		post.Slug = "Slug-" + pf
		post.DraftContent = "Draft ccc content -- " + pf
		post.Created = n
		post.DraftUpdate = n
		post.Updated = n
		switch i % 3 {
		case 0:
			post.Draft = 1
		case 1:
			post.Publish = 1
			post.SetPublic(errSrv.PublicPost{
				Title:   "title-" + pf,
				Content: "content-" + pf,
				Slug:    "post" + pf,
				Author: errSrv.Author{
					Name: "Tom",
				},
			})
		case 2:
			post.Draft = 1
			post.Publish = 1
		}
		db.Model(&errSrv.Post{}).FirstOrCreate(post, &errSrv.Post{ID: post.ID})
	}
	errSrv.Run("localhost:8880")
}
