package errSrv

import (
	"github.com/kataras/iris/v12"
	"github.com/kataras/iris/v12/context"
	"regexp"
	"strconv"
	"strings"
)

func initArtApi(app *iris.Application) {
	post := app.Party("/post")
	post.Get("/{slug}", getPost)
	posts := app.Party("/posts")
	posts.Get("/{page}", getPosst)

	edit := app.Party("/edit")
	auth(edit.Get, "/{page}", getEdits)
	// save
	auth(edit.Put, "/", savArt)
	//publish
	auth(edit.Post, "/", artPub)
	// unpublish
	auth(edit.Patch, "/{id}", unPub)
	// del
	auth(edit.Delete, "/{id}", delArt)
}

func getPosst(ctx *context.Context) {
	page := ctx.Params().GetIntDefault("page", 1)
	count := ctx.URLParamIntDefault("count", 5)
	if page == 0 {
		page = 1
	}
	if count == 0 {
		count = 5
	}
	p := make([]PubLisArt, 0)
	db.Model(&Art{}).Offset((page-1)*count).
		Limit(count).Where("arts.updated != ?", 0).
		Order("created desc, updated desc").
		Find(&p)
	for i, a := range p {
		a.Content = fixContent(a.Content)
		p[i] = a
	}
	ls := &ListPubPost{
		Posts: p,
		Total: (sys.TotalPubPosts + count - 1) / count,
		Cur:   page,
	}
	ctx.JSON(ls)
}

func fixContent(c string) string {
	c = strings.ReplaceAll(c, "\n", "")
	c = regexp.MustCompile("!?\\[.*?]").ReplaceAllString(c, "")
	c = regexp.MustCompile("!?\\(.*?\\)").ReplaceAllString(c, "")
	c = regexp.MustCompile("!?```.*?```").ReplaceAllString(c, "")
	return c
}

func getEdits(ctx *context.Context) {
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
	tx.Order("save_at desc").Find(&p)
	ls := &ListPost{
		Posts: p,
		Total: (t + count - 1) / count,
		Cur:   page,
	}
	ctx.JSON(ls)
}

func getPost(ctx *context.Context) {
	p := &Art{}
	err := db.Preload("Author").
		First(p, "slug = ?", ctx.Params().Get("slug")).Error
	if err == nil {
		pp := p.PubArt
		pp.AID = p.ID
		ctx.JSON(pp)
	} else {
		handleErr(ctx, err)
	}
}

func unPub(ctx *context.Context) {
	pa := ctx.Params()
	id, err := pa.GetUint("id")
	p := &Art{
		ID: id,
	}
	if err == nil {
		err = db.Model(p).Update("updated", 0).Error
	}
	if err != nil {
		handleErr(ctx, err)
	} else {
		countPos()
		ctx.StatusCode(200)
	}
}

func delArt(ctx *context.Context) {
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
		//if err == nil {
		//	err = db.Where("a_id = ?", id).Delete(&ArtHis{}).Error
		//}
	}
	if err == nil {
		ctx.WriteString(strconv.Itoa(int(id)))
	} else {
		countPos()
		handleErr(ctx, err)
	}
}

func savArt(ctx *context.Context) {
	p := &Art{}
	ctx.ReadJSON(p)
	err := p.Save()
	if err != nil {
		handleErr(ctx, err)
	} else {
		ctx.WriteString(strings.Join([]string{
			strconv.Itoa(int(p.ID)),
			strconv.Itoa(int(p.SaveAt)),
		}, "\u0001"))
	}
}

func artPub(ctx *context.Context) {
	p := &Art{}
	ctx.ReadJSON(p)
	v := p.Updated
	err := p.Publish()
	if err != nil {
		handleErr(ctx, err)
	} else {
		if v == 0 {
			countPos()
		}
		ctx.StatusCode(200)
		ctx.WriteString(strings.Join([]string{
			strconv.Itoa(int(p.ID)),
			p.Slug,
			strconv.Itoa(int(p.Updated)),
			nTags,
			dTags,
		}, "\u0001"))
	}
}
