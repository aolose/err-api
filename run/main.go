package main

import (
	"errSrv"
	"strconv"
)

func main() {
	errSrv.Connect()
	db := errSrv.DB()
	for i := 0; i < 60; i++ {
		post := &errSrv.Post{}
		post.Status = i % 3
		pf := strconv.Itoa(i)
		if i%3 == 1 {
			post.SetPublic(errSrv.PublicPost{
				Title:   "title-" + pf,
				Content: "content-" + pf,
				Slug:    "post" + pf,
				Author: errSrv.Author{
					Name: "Tom",
				},
			})
		} else {
			post.DraftTitle = "Draft-" + pf
			post.DraftContent = "Draft ccc content -- " + pf
		}
		db.FirstOrCreate(post, &errSrv.Post{
			PublicPost: errSrv.PublicPost{
				Title: post.Title,
			},
			DraftPost: errSrv.DraftPost{
				DraftTitle: post.DraftTitle,
			},
		})
	}
	errSrv.Run("localhost:8880")
}
