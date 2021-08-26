package errSrv

import (
	"fmt"
	"github.com/kataras/iris/v12"
	"strconv"
	"strings"
	"time"
)

func Run(addr string) {
	go func() {
		for {
			time.Sleep(time.Second * 5)
			n := time.Now()
			if nextSyncSys.Before(n) {
				syncSys()
				nextSyncSys = nextSyncSys.Add(time.Hour * 999)
			}
		}
	}()
	app := iris.New()
	app.UseRouter(func(ctx iris.Context) {
		r := ctx.Request()
		fmt.Printf("%v \t %v\n", r.Method, r.URL)
		ctx.Header("Access-Control-Allow-Origin", "*")
		ctx.Header("Access-Control-Allow-Credentials", "true")
		if ctx.Method() == iris.MethodOptions {
			ctx.Header("Access-Control-Allow-Methods", "GET,POST,PUT,PATCH,DELETE,HEAD,OPTIONS")
			ctx.Header("Access-Control-Max-Age", "86400")
			ctx.StatusCode(204)
		} else {
			ctx.Next()
		}
	})
	app.OnAnyErrorCode(func(ctx iris.Context) {
		r := ctx.Request()
		fmt.Printf("Error %v \t %v %v\n", r.Method, r.URL, ctx.GetErr())
		ctx.Next()
	})
	app.Get("/msg", func(ctx iris.Context) {
		flusher, ok := ctx.ResponseWriter().Flusher()
		if !ok {
			return
		}
		ctx.ContentType("text/event-stream")
		ctx.Header("Cache-Control", "no-cache")
		now := time.Now()
		ctx.Writef("data: The server time is: %s\n\n", now)
		flusher.Flush()
	})
	post := app.Party("/post")
	post.Get("/{slug}", getPost)
	post.Get("/ctx", auth(getCtx))
	posts := app.Party("/posts")
	posts.Get("/{page}", getPosst)

	tag := app.Party("/tag")
	tag.Get("/ls", func(ctx iris.Context) {
		tg := make([]string, len(tagsCache))
		i := 0
		for k, _ := range tagsCache {
			tg[i] = k
			i++
		}
		ctx.StatusCode(200)
		ctx.JSON(tg)
	})
	tag.Get("/all", func(ctx iris.Context) {
		var gs []Tag
		err := db.Find(&gs).Error
		if err == nil {
			ctx.StatusCode(200)
			ctx.JSON(gs)
		} else {
			handleErr(ctx, err)
		}
	})
	tag.Get("/{name}/{page}", func(ctx iris.Context) {

	})
	his := app.Party("/his")
	his.Get("/{id}/{ver}", func(ctx iris.Context) {})
	his.Delete("/{id}/{ver}", func(ctx iris.Context) {})

	edit := app.Party("/edit")
	edit.Get("/{page}", auth(getEdits))
	// save
	edit.Put("/", func(ctx iris.Context) {
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
	})
	//publish
	edit.Post("/", func(ctx iris.Context) {
		p := &Art{}
		ctx.ReadJSON(p)
		v := p.Version
		err := p.Publish()
		if err != nil {
			handleErr(ctx, err)
		} else {
			if v == -1 {
				nextSysSync(time.Second * 2)
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
	})
	// unpublish
	edit.Patch("/{id}/{ver}", func(ctx iris.Context) {
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
				nextSysSync(time.Second * 2)
			}
			ctx.StatusCode(200)
		}
	})
	// del
	edit.Delete("/{id}", func(ctx iris.Context) {
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
			nextSysSync(time.Second * 2)
			handleErr(ctx, err)
		}
	})

	_ = app.Listen(addr)
}

func handleErr(ctx iris.Context, err error) {
	fmt.Errorf("%v", err)
	ctx.StatusCode(500)
	ctx.JSON(err)
}

func getCtx(ctx iris.Context) {
	id, err := ctx.URLParamInt("id")
	ver := ctx.URLParamInt64Default("ver", -1)
	if err == nil {
		c := &ArtHis{}
		err = db.Take(c, ArtHis{
			AID:     uint(id),
			Version: ver,
		}).Error
		if err == nil {
			ctx.JSON(iris.Map{
				"c": c.Content,
			})
		}
	}
	if err != nil {
		handleErr(ctx, err)
	}
}

type ListPubPost struct {
	Posts []PubLisArt `json:"ls"`
	Total int         `json:"total"`
	Cur   int         `json:"cur"`
}

type ListPost struct {
	Posts []Art `json:"ls"`
	Total int   `json:"total"`
	Cur   int   `json:"cur"`
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
	p := []Art{}
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
