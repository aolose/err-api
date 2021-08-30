package errSrv

import "github.com/kataras/iris/v12"

func initTagsApi(app *iris.Application) {
	tag := app.Party("/tag")
	tag.Get("/ls", getTags)
	tag.Get("/all", getTags2)
	tag.Get("/{name}/{page}", getTagArt)
}

func getTags(ctx iris.Context) {
	tg := make([]string, len(tagsCache))
	i := 0
	for k, _ := range tagsCache {
		tg[i] = k
		i++
	}
	ctx.StatusCode(200)
	ctx.JSON(tg)
}

func getTags2(ctx iris.Context) {
	var gs []Tag
	err := db.Find(&gs).Error
	if err == nil {
		ctx.StatusCode(200)
		ctx.JSON(gs)
	} else {
		handleErr(ctx, err)
	}
}

func getTagArt(ctx iris.Context) {

}
