package errSrv

import "github.com/kataras/iris/v12"

func Run(addr string) {
	app := iris.New()
	post := app.Party("/post")
	post.Get("/{slug}", getPost)
	post.Post("/", setPost)
	_ = app.Listen(addr)
}

func getPost(ctx iris.Context) {
	p := &Post{Status: 0}
	tx := db.Preload("Author").First(p, "slug = ?", ctx.Params().Get("slug"))
	if tx.Error != nil {
		println(tx.Error)
	}
	ctx.JSON(p.GetPublic())
}

func setPost(ctx iris.Context) {

}
func getRes(ctx iris.Context) {

}
func setRes(ctx iris.Context) {

}
func getComments(ctx iris.Context) {

}
func setComment(ctx iris.Context) {

}
