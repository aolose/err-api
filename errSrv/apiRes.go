package errSrv

import (
	"bytes"
	"github.com/disintegration/imaging"
	"github.com/h2non/filetype"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm/clause"
	"log"
	"math/rand"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"time"
)

const maxSize = 8 * iris.MB

var fileCache map[string][]multipart.File
var fileInfoCache map[string][3]string
var fileFirstCache map[string][]byte
var chs map[int64]chan string

func initResApi(app *iris.Application) {
	app.HandleDir("/r",
		iris.Dir("./dist"),
		iris.DirOptions{
			ShowList: false,
		},
	)
	app.Get("/msg", auth(msg))
	app.Post("/upload", upload)
	res := app.Party("/res")
	res.Get("/{page}", resLs)
	res.Patch("/{id}/{name}", resCh)
	res.Delete("/", resDel)
}
func resCh(ctx iris.Context) {
	pm := ctx.Params()
	nm := pm.Get("name")
	id := pm.Get("id")
	err := db.Model(&Res{}).Where("id = ?", id).Update("name", nm).Error
	handleErr(ctx, err)
}

func resDel(ctx iris.Context) {
	id := ctx.URLParam("id")
	if id == "" {
		ctx.StatusCode(200)
	} else {
		ids := strings.Split(id, ",")
		err := db.Delete(&Res{}, "ID in ?", ids).Error
		countRes()
		handleErr(ctx, err)
		addJob(func() {
			for _, i := range ids {
				er := os.Remove("./dist/" + i)
				_ = os.Remove("./dist/" + i + ".png")
				if er != nil {
					log.Printf("del %s file fail: %v \n", i, er)
				}
			}
		})
	}
}
func resLs(ctx iris.Context) {
	pg := ctx.Params().GetIntDefault("page", 1)
	count := ctx.URLParamIntDefault("c", 20)
	search := ctx.URLParam("k")
	var c int64
	t := sys.TotalRes
	ls := make([]Res, 0)
	tx := db.Offset((pg - 1) * count).Limit(count)
	if search != "" {
		v := "%" + search + "%"
		q := "name Like ? OR type Like ?"
		tx = tx.Where(q, v, v)
		db.Table("res").Where(q, v, v).Count(&c)
		t = int(c)
	}
	err := tx.Order("date desc").Find(&ls).Error
	if err == nil {
		ll := &ListRes{
			List:  ls,
			Total: (t + count - 1) / count,
			Cur:   pg,
		}
		_, _ = ctx.JSON(ll)
	} else {
		handleErr(ctx, err)
	}
}

func upload(ctx iris.Context) {
	ctx.SetMaxRequestBodySize(maxSize)
	key := ctx.FormValue("title")
	nm := ctx.FormValue("name")
	if fileCache == nil {
		fileCache = make(map[string][]multipart.File)
		fileInfoCache = make(map[string][3]string)
		fileFirstCache = make(map[string][]byte)
	}
	pt := ctx.FormValue("part")
	part, _ := strconv.Atoi(pt)
	total, _ := strconv.Atoi(ctx.FormValue("total"))
	total = total - 1
	if _, ok := fileCache[key]; !ok {
		fileCache[key] = make([]multipart.File, total)
	}
	file, _, _ := ctx.FormFile("data")

	if pt == "0" {
		buf := new(bytes.Buffer)
		_, _ = buf.ReadFrom(file)
		bt := buf.Bytes()
		kind, _ := filetype.Match(bt)
		fileInfoCache[key] = [3]string{nm, kind.MIME.Value, kind.Extension}
		fileFirstCache[key] = bt
	} else {
		fileCache[key][part-1] = file
	}
	done := true
	for i := 0; i < total; i++ {
		if fileCache[key][i] == nil {
			done = false
			break
		}
	}
	if done {
		addJob(func() {
			_ = os.Mkdir("dist", os.ModePerm)
			fn := "dist/" + key
			f, _ := os.OpenFile(fn, os.O_WRONLY|os.O_CREATE, 0666)
			defer func(f *os.File) {
				err := f.Close()
				if err != nil {
				}
			}(f)
			fk := fileFirstCache[key]
			che := fileCache[key]
			ii, er := f.Write(fk)
			i := int64(ii)
			for _, ff := range che {
				buf := new(bytes.Buffer)
				_, _ = buf.ReadFrom(ff)
				s, _ := f.Seek(i, 0)
				ii, er = f.WriteAt(buf.Bytes(), s)
				if er != nil {
					log.Printf("write file err %v", er)
				}
				i = int64(ii) + i
			}
			inf, _ := fileInfoCache[key]
			re := Res{
				Name: inf[0],
				Type: inf[1],
				Ext:  inf[2],
				Size: i,
				ID:   key,
				Date: time.Now().Unix(),
			}
			db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(re)
			delete(fileCache, key)
			delete(fileInfoCache, key)
			delete(fileFirstCache, key)
			if strings.HasPrefix(re.Type, "image") {
				thumbnail(f)
			}
			for _, cha := range chs {
				if cha != nil {
					cha <- strings.Join([]string{key, pt}, ",")
				}
			}
		})
	} else {
		for _, cha := range chs {
			if cha != nil {
				nf := fileInfoCache[key]
				cha <- strings.Join([]string{key, pt, nf[1], nf[2]}, ",")
			}
		}
	}
	ctx.StatusCode(200)
	countRes()
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
				cha = nil
				delete(chs, key)
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

func thumbnail(f *os.File) {
	n := f.Name()
	_ = f.Close()
	img, err := imaging.Open(n)
	if err == nil {
		img := imaging.Resize(img, 0, 200, imaging.Lanczos)
		err = imaging.Save(img, n+".png")
		if err != nil {
			log.Printf("thumbnail err %v", err)
		}
	}
}
