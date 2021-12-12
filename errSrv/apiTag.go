package errSrv

import "github.com/kataras/iris/v12"
import "github.com/kataras/iris/v12/context"

func initTagsApi(app *iris.Application) {
	tag := app.Party("/tag")
	auth(tag.Get, "/ls", getTags)
	tag.Get("/all", getTags2)
	tag.Get("/{name}/{page}", getTagArt)
}

func getTags(ctx *context.Context) {
	tg := make([]string, len(tagsCache))
	i := 0
	for k := range tagsCache {
		tg[i] = k
		i++
	}
	ctx.StatusCode(200)
	ctx.JSON(tg)
}

func getTags2(ctx *context.Context) {
	var gs []Tag
	err := db.Find(&gs).Error
	if err == nil {
		ctx.StatusCode(200)
		ctx.JSON(gs)
	} else {
		handleErr(ctx, err)
	}
}

func getTagArt(ctx *context.Context) {
	page := ctx.Params().GetIntDefault("page", 1)
	name := ctx.Params().GetStringDefault("name", "")
	count := ctx.URLParamIntDefault("count", 5)
	if page == 0 {
		page = 1
	}
	if count == 0 {
		count = 5
	}
	if name == "" {
		ctx.StatusCode(404)
	} else {
		p := make([]PubLisArt, 0)
		if ids, ok := tagsCache[name]; ok {
			db.Model(&Art{}).Offset((page-1)*count).
				Limit(count).Where("arts.updated != ? and arts.id in ?", 0, ids).
				Order("created desc, updated desc").
				Find(&p)
			for i, a := range p {
				a.Content = fixContent(a.Content)
				p[i] = a
			}
			ls := &ListPubPost{
				Posts: p,
				Total: (len(ids) + count - 1) / count,
				Cur:   page,
			}
			ctx.JSON(ls)
		}
	}
}
