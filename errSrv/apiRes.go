package errSrv

import (
	"bytes"
	"github.com/h2non/bimg"
	"github.com/h2non/filetype"
	"github.com/kataras/iris/v12"
	"gorm.io/gorm/clause"
	"log"
	"mime/multipart"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

var ml = 0
var mg sync.Map

func setMsg(m string) {
	mg.Store(ml, m)
	ml++
}
func getMsg() string {
	if ml == 0 {
		return ""
	}
	ml--
	m, ok := mg.LoadAndDelete(ml)
	if ok {
		return m.(string)
	}
	return ""
}

const maxSize = 8 * iris.MB

var fileCache map[string][]multipart.File
var fileInfoCache map[string][3]string
var fileFirstCache map[string][]byte
var wg sync.WaitGroup

func initResApi(app *iris.Application) {
	wg = sync.WaitGroup{}
	app.HandleDir("/r",
		iris.Dir("./dist"),
		iris.DirOptions{
			ShowList: false,
		},
	)
	auth(app.Get, "/msg", msg)
	auth(app.Post, "/upload", upload)
	res := app.Party("/res")
	auth(res.Get, "/{page}", resLs)
	auth(res.Patch, "/{id}/{name}", resCh)
	auth(res.Delete, "/", resDel)
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
				_ = os.Remove("./dist/" + i + ".webp")
				if er != nil {
					log.Printf("del %s file fail: %v \n", i, er)
				}
			}
		})
	}
}

func wait(fn ...func()) {
	l := len(fn)
	wg.Add(l)
	for _, f := range fn {
		ff := f
		go func() {
			defer wg.Done()
			ff()
		}()
	}
	wg.Wait()
}

func resLs(ctx iris.Context) {
	pg := ctx.Params().GetIntDefault("page", 1)
	count := ctx.URLParamIntDefault("c", 20)
	img := ctx.URLParamIntDefault("img", 0)
	search := ctx.URLParam("k")
	var c int64
	t := sys.TotalRes
	ls := make([]Res, 0)
	tx := db.Offset((pg - 1) * count).Limit(count)
	if search != "" || img != 0 {
		v := "%" + search + "%"
		q := "name Like ? OR type Like ?"
		if search == "" {
			q = "type like ?"
			tx = tx.Where(q, "image%")
			db.Table("res").Where(q, "image%").Count(&c)
		} else if img != 0 {
			q = q + "name Like ? And  type like ?"
			tx = tx.Where(q, v, "image%")
			db.Table("res").Where(q, v, "image%").Count(&c)
		} else {
			tx = tx.Where(q, v, v)
			db.Table("res").Where(q, v, v).Count(&c)
		}
		t = int(c)
	}
	err := tx.Order("date desc").Find(&ls).Error
	if err == nil {
		ll := &ListResult{
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
	key := ctx.FormValue("key")
	nm := ctx.FormValue("nm")
	if fileCache == nil {
		fileCache = make(map[string][]multipart.File)
		fileInfoCache = make(map[string][3]string)
		fileFirstCache = make(map[string][]byte)
	}
	if nm != "" {
		a := strings.Split(nm, ".")
		c := a[len(a)-1]
		tp := ctx.FormValue("tp")
		if c != "" {
			wait(func() { fileInfoCache[key] = [3]string{nm, tp, c} })
		}
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
		kind, e := filetype.Match(bt)
		if e == nil && kind.MIME.Value != "" {
			wait(func() { fileInfoCache[key] = [3]string{nm, kind.MIME.Value, kind.Extension} })
		}
		wait(
			func() { fileFirstCache[key] = bt },
		)
	} else {
		wait(func() { fileCache[key][part-1] = file })
	}
	done := fileFirstCache[key] != nil
	fc := fileCache[key]
	for i := 0; i < total; i++ {
		if fc[i] == nil {
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
			for n := 0; n < total; n++ {
				ff := che[n]
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
			delete(fileCache, key)
			delete(fileInfoCache, key)
			delete(fileFirstCache, key)
			if strings.HasPrefix(inf[1], "image") {
				i = int64(thumbnail(f, i))
			}
			re := &Res{
				Name: inf[0],
				Type: inf[1],
				Ext:  inf[2],
				Size: i,
				ID:   key,
				Date: now(),
			}
			db.Clauses(clause.OnConflict{
				UpdateAll: true,
			}).Create(re)
			setMsg(strings.Join([]string{key, pt, "", "", strconv.Itoa(int(i))}, ","))
		})
	} else {
		nf := fileInfoCache[key]
		setMsg(strings.Join([]string{key, pt, nf[1], nf[2]}, ","))
	}
	ctx.StatusCode(200)
	countRes()
}

func msg(ctx iris.Context) {
	if strings.Contains(ctx.GetHeader("accept"), "text/event-stream") {
		flusher, ok := ctx.ResponseWriter().Flusher()
		if !ok {
			return
		}
		ctx.ContentType("text/event-stream")
		ctx.Header("Cache-Control", "no-cache")
		go func() {
			flusher.Flush()
			time.Sleep(time.Second * 3)
		}()
		for {
			if ctx.IsCanceled() {
				break
			}
			c := getMsg()
			if c != "" {
				_, _ = ctx.Writef("data: %s\n\n", c)
			} else {
				time.Sleep(time.Millisecond * 500)
			}
			flusher.Flush()
		}
	} else {
		ctx.StatusCode(404)
	}
}

func compressImg(buf []byte, n, nm string, sm bool, s int64) (error, int) {
	var img []byte
	var err error
	sz, _ := bimg.NewImage(buf).Size()
	if sm && sz.Height > 500 {
		img, err = bimg.NewImage(buf).Convert(bimg.WEBP)
		img, err = bimg.NewImage(img).Resize(0, 500)
	} else {
		tp := bimg.NewImage(buf).Type()
		if tp == "gif" {
			img = buf
		} else {
			img, err = bimg.NewImage(buf).Convert(bimg.WEBP)
			if err == nil && sz.Width > 3840 {
				img, _ = bimg.NewImage(img).Resize(3080, 0)
			}
		}
	}
	if err == nil {
		log.Printf("compress: before %d after %d \n", s, len(img))
		if s == 0 || len(img) < int(s) {
			err = bimg.Write(n+nm, img)
		} else {
			return nil, int(s)
		}
	}
	return err, len(img)
}

func thumbnail(f *os.File, s int64) int {
	n := f.Name()
	_ = f.Close()
	i := 0
	buffer, err := bimg.Read(n)

	if err == nil {
		err, _ = compressImg(buffer, n, ".webp", true, 0)
		if err == nil {
			err, i = compressImg(buffer, n, "", false, s)
		}
		if err != nil {
			log.Printf("thumbnail err %v", err)
		}
	}
	return i
}
