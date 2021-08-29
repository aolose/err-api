package errSrv

import (
	"bytes"
	"fmt"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm/clause"
	"math/rand"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"
)

var chs map[int64]chan string

func Run(addr string) {
	chs = make(map[int64]chan string)
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
		ctx.Header("Access-Control-Allow-Origin", r.Header.Get("origin"))
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
	app.Get("/msg", auth(msg))
	const maxSize = 8 * iris.MB

	var fileCache map[string][]multipart.File
	var fileInfoCache map[string][2]string

	app.Post("/upload", func(ctx iris.Context) {
		ctx.SetMaxRequestBodySize(maxSize)
		key := ctx.FormValue("title")
		nm := ctx.FormValue("name")
		tp := ctx.FormValue("type")
		if fileCache == nil {
			fileCache = make(map[string][]multipart.File)
			fileInfoCache = make(map[string][2]string)
		}
		if nm != "" {
			fileInfoCache[key] = [2]string{nm, tp}
		}
		pt := ctx.FormValue("part")
		part, _ := strconv.Atoi(pt)
		total, _ := strconv.Atoi(ctx.FormValue("total"))
		if _, ok := fileCache[key]; !ok {
			fileCache[key] = make([]multipart.File, total)
		}
		file, _, _ := ctx.FormFile("data")
		fileCache[key][part] = file
		done := true
		for i := 0; i < total; i++ {
			if fileCache[key][i] == nil {
				done = false
				break
			}
		}
		if done {
			fn := "dist/" + key
			f, _ := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0666)
			defer f.Close()
			i := int64(0)
			for _, ff := range fileCache[key] {
				buf := new(bytes.Buffer)
				_, _ = buf.ReadFrom(ff)
				s, _ := f.Seek(i, 0)
				ii, _ := f.WriteAt(buf.Bytes(), s)
				i = int64(ii) + i
			}
			inf, _ := fileInfoCache[key]
			re := Res{
				Name: inf[0],
				Type: inf[1],
				Size: i,
				ID:   key,
			}
			db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(re)
			delete(fileCache, key)
			delete(fileInfoCache, key)
		}
		ctx.StatusCode(200)
		for _, cha := range chs {
			if cha != nil {
				cha <- strings.Join([]string{key, pt}, ",")
			}
		}
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
func msg(ctx iris.Context) {
	key := rand.Int63n(1e8)
	flusher, ok := ctx.ResponseWriter().Flusher()
	if !ok {
		return
	}
	cha := make(chan string)
	chs[key] = cha
	ctx.ContentType("text/event-stream")
	ctx.Header("Cache-Control", "no-cache")
	go func() {
		ctx.StatusCode(200)
		flusher.Flush()
		time.Sleep(time.Second)
	}()
	for {
		if ctx.IsCanceled() {
			go func() {
				_ = <-cha
				close(cha)
			}()
			time.Sleep(time.Millisecond * 10)
			cha = nil
			delete(chs, key)
			break
		}
		c := <-cha
		_, _ = ctx.Writef("data: %s\n\n", c)
		flusher.Flush()
	}
}
