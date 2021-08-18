package errSrv

import "github.com/kataras/iris/v12"

func Run(addr string) {
	app := iris.New()
	app.UseGlobal(func(ctx iris.Context) {
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Max-Age", "3600")
		ctx.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS")
		if ctx.Method() == "OPTION" {
			ctx.StatusCode(204)
		} else {
			ctx.Next()
		}
	})
	app.Get("/edits/{page}", auth(getEdits))
	post := app.Party("/post")
	post.Get("/{slug}", getPost)
	post.Post("/", setPost)
	posts := app.Party("/posts")

	posts.Get("/{page}", getPosst)
	_ = app.Listen(addr)
}

type ListPubPost struct {
	Posts []*PublicPost `json:"ls"`
	Total int           `json:"total"`
	Cur   int           `json:"cur"`
}

type ListPost struct {
	Posts []*EditPost `json:"ls"`
	Total int         `json:"total"`
	Cur   int         `json:"cur"`
}

func auth(next func(ctx iris.Context)) func(ctx iris.Context) {
	return func(ctx iris.Context) {
		// -- todo check
		next(ctx)
	}
}

func getPosst(ctx iris.Context) {
	page := ctx.Params().GetIntDefault("page", 1)
	count := ctx.URLParamIntDefault("count", 5)
	if page == 0 {
		page = 1
	}
	if count == 0 {
		count = 5
	}
	p := []*Post{}
	pp := []*PublicPost{}
	db.Offset((page-1)*count).Limit(count).Where("publish = ?", 1).Find(&p)
	for _, i := range p {
		pp = append(pp, i.GetPublic())
	}
	ls := &ListPubPost{
		Posts: pp,
		Total: (sys.TotalPubPosts + count - 1) / count,
		Cur:   page,
	}
	ctx.JSON(ls)
}
func getEdits(ctx iris.Context) {
	page := ctx.Params().GetIntDefault("page", 1)
	count := ctx.URLParamIntDefault("count", 20)
	search := ctx.URLParam("k")
	if page == 0 {
		page = 1
	}
	if count == 0 {
		count = 20
	}
	p := []*Post{}
	var c int64
	t := sys.TotalPosts
	tx := db.Offset((page - 1) * count).Limit(count)
	if search != "" {
		v := "%" + search + "%"
		tx = tx.Where("title Like ? OR content Like ?", v, v)
		db.Table("post").Where("title Like ? OR content Like ?", v, v).Count(&c)
		t = int(c)
	}
	tx.Order("updated desc, created desc").Find(&p)
	pp := make([]*EditPost, len(p))
	for i, v := range p {
		pp[i] = v.GetEdit()
	}
	ls := &ListPost{
		Posts: pp,
		Total: (t + count - 1) / count,
		Cur:   page,
	}
	ctx.JSON(ls)
}

func getPost(ctx iris.Context) {
	p := &Post{Publish: 1}
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
