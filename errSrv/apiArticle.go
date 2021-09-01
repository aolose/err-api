package errSrv

import (
	"github.com/kataras/iris/v12"
	"strconv"
	"strings"
)

func initArtApi(app *iris.Application) {
	post := app.Party("/post")
	post.Get("/{slug}", getPost)
	posts := app.Party("/posts")
	posts.Get("/{page}", getPosst)

	edit := app.Party("/edit")
	edit.Get("/{page}", auth(getEdits))
	// save
	edit.Put("/", auth(savArt))
	//publish
	edit.Post("/", auth(artPub))
	// unpublish
	edit.Patch("/{id}/{ver}", auth(unPub))
	// del
	edit.Delete("/{id}", auth(delArt))
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
	p := []PubLisArt{}
	db.Model(&Art{}).Offset((page-1)*count).
		Limit(count).Where("arts.version != ?", -1).Take(&p)
	ls := &ListPubPost{
		Posts: p,
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
	p := make([]Art, 0)
	var c int64
	t := sys.TotalPosts
	tx := db.Offset((page - 1) * count).Limit(count)
	if search != "" {
		v := "%" + search + "%"
		tx = tx.Where("title Like ? OR content Like ?", v, v)
		db.Table("post").Where("title Like ? OR content Like ?", v, v).Count(&c)
		t = int(c)
	}
	tx.Order("updated desc, save_at desc").Find(&p)
	ls := &ListPost{
		Posts: p,
		Total: (t + count - 1) / count,
		Cur:   page,
	}
	ctx.JSON(ls)
}

func getPost(ctx iris.Context) {
	p := &Art{}
	err := db.Preload("Author").First(p, "slug = ?", ctx.Params().Get("slug")).Error
	if err == nil {
		pp := p.PubArt
		ctx.JSON(pp)
	} else {
		handleErr(ctx, err)
	}
}

func unPub(ctx iris.Context) {
	pa := ctx.Params()
	ver := -1
	id, err := pa.GetUint("id")
	if err == nil {
		ver, err = pa.GetInt("ver")
	}
	p := &Art{
		ID: id,
	}
	if err == nil {
		err = db.Model(p).Update("version", ver).Error
	}
	if err != nil {
		handleErr(ctx, err)
	} else {
		if ver == -1 {
			countPos()
		}
		ctx.StatusCode(200)
	}
}

func delArt(ctx iris.Context) {
	id, err := ctx.Params().GetUint("id")
	if err == nil {
		a := &Art{ID: id}
		err = db.Find(a).Error
		if err == nil {
			err = delTags(id, strings.Split(a.Tags, " ")...)
		}
	}
	if err == nil {
		err = db.Delete(&Art{ID: id}).Error
		if err == nil {
			err = db.Where("a_id = ?", id).Delete(&ArtHis{}).Error
		}
	}
	if err == nil {
		ctx.WriteString(strconv.Itoa(int(id)))
	} else {
		countPos()
		handleErr(ctx, err)
	}
}

func savArt(ctx iris.Context) {
	p := &Art{}
	ctx.ReadJSON(p)
	err := p.Save()
	if err != nil {
		handleErr(ctx, err)
	} else {
		ctx.WriteString(strings.Join([]string{
			strconv.Itoa(int(p.ID)),
			strconv.Itoa(int(p.Version)),
			strconv.Itoa(int(p.SaveAt)),
		}, "\u0001"))
	}
}

func artPub(ctx iris.Context) {
	p := &Art{}
	ctx.ReadJSON(p)
	v := p.Version
	err := p.Publish()
	if err != nil {
		handleErr(ctx, err)
	} else {
		if v == -1 {
			countPos()
		}
		ctx.StatusCode(200)
		ctx.WriteString(strings.Join([]string{
			strconv.Itoa(int(p.ID)),
			p.Slug,
			strconv.Itoa(int(p.Version)),
			strconv.Itoa(int(p.Updated)),
			nTags,
			dTags,
		}, "\u0001"))
	}
}
