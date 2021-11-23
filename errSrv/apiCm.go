package errSrv

import (
	"errors"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"regexp"
	"strconv"
	"strings"
)

var totalCm int64

func initCmApi(app *iris.Application) {
	cm := app.Party("c")
	cm.Post("/", cmCreate)
	cm.Get("/{id}/{page}", cmLs)
	auth(cm.Get, "/m/{page}", cmList)
	auth(cm.Get, "/{page}", pageQuery(Comment{}, &totalCm, "art_id", "%content%"))
	auth(cm.Delete, "/", cmDel)
	cm.Delete("/{id}", cmDel2)
	auth(cm.Post, "/{id}", cmOpt)
	sys.CmLife = 3600 * 24 * 2 // 2day
	countCm()
}

type CMLs struct {
	R     []Comment `json:"r"`
	C     []Comment `json:"ls"`
	Cur   int       `json:"cur"`
	Total int       `json:"total"`
}

func cmLs(ctx iris.Context) {
	pms := ctx.Params()
	page := pms.GetIntDefault("page", 1)
	id := pms.GetUintDefault("id", 0)
	count := ctx.URLParamIntDefault("count", 10)
	var err error
	if id == 0 {
		err = errors.New("article not exist")
	}
	if err == nil {
		if page == 0 {
			page = 1
		}
		if count == 0 {
			count = 5
		}

		cm := make([]Comment, 0)
		cr := make([]Comment, 0)
		t := int64(0)
		db.Model(&Comment{}).Where("art_id=? AND status>?", id, 0).Count(&t)
		db.Model(&Comment{}).Offset((page-1)*count).Limit(count).
			Where("art_id=? AND status>?", id, 0).
			Order("created desc").
			Find(&cm)
		ci := make([]uint, 0)
		cn := make([]uint, 0)
		for _, c := range cm {
			ci = append(ci, c.ID)
		}
		for _, c := range cm {
			if c.Reply > 0 {
				for _, i := range ci {
					if i == c.Reply {
						break
					}
				}
				cn = append(cn, c.Reply)
			}
		}
		if len(cn) > 0 {
			db.Where("reply IN ?", cn).Find(&cr)
		}
		ck := ctx.GetCookie("cm_tk")
		if ck != "" {
			n := now() - sys.CmLife
			for i, c := range cm {
				if c.Token == ck && c.Saved > n {
					cm[i].Own = 1
				}
			}
		}
		ctx.JSON(CMLs{
			R:     cr,
			C:     cm,
			Total: (int(t) + count - 1) / count,
			Cur:   page,
		})
	}
	if err != nil {
		handleErr(ctx, err)
	}
}

func cmList(ctx iris.Context) {

}

func countCm() {
	db.Model(&Comment{}).Count(&totalCm)
}

func cmCreate(ctx iris.Context) {
	ck := ctx.GetCookie("cm_tk")
	var err error
	if sys.DisCm == 1 {
		err = errors.New("comment close")
	}
	c := &Comment{}
	if err == nil {
		err = ctx.ReadJSON(c)
	}
	a := &Art{}
	if err == nil {
		err = db.First(a, c.ArtID).Error
	}
	if err == nil {
		if a.DisCm == 1 {
			err = errors.New("comment close")
		}
	}
	if err == nil {
		c.Name = strings.TrimSpace(c.Name)
		c.Content = strings.TrimSpace(c.Content)
		la := len(c.Content)
		lb := len(c.Name)
		allow := regexp.MustCompile("^[0-9a-z· \u4e00-\u9fa5]+$")
		if lb == 0 || la == 0 {
			err = errors.New("name or comment is empty")
		} else if lb > sys.CnLen || la > sys.CmLen {
			err = errors.New("name or comment too long")
		} else if !allow.MatchString(c.Name) {
			err = errors.New("illegal name format")
		}
	}
	if err == nil {
		if sys.AuditCm == 0 && a.AuditCm == 0 {
			c.Status = 1
		} else {
			c.Status = 0
		}
		if c.ID > 0 && ck != "" {
			d := &Comment{
				ID: c.ID,
			}
			err = db.Where("token = ?", ck).First(d).Error
			if err == nil {
				if d.Saved+sys.CmLife < now() {
					err = errors.New("can't edit")
				}
			}
			if err == nil {
				d.Content = c.Content
				db.Save(d)
			}
		} else {
			if ck == "" {
				ck = enc(uuid.New().String())
			}
			c.ID = 0
			c.IP = getIP(ctx)
			c.From = ""
			c.Token = ck
			c.Saved = now()
			err = db.Save(c).Error
			if err == nil {
				totalCm = totalCm + 1
			}
		}
	}
	ctx.SetCookie(&iris.Cookie{
		Name:     "cm_tk",
		Value:    ck,
		HttpOnly: true,
		MaxAge:   int(sys.CmLife),
		SameSite: iris.SameSiteLaxMode,
		Path:     "/",
	}, iris.CookieAllowSubdomains())
	if err == nil {
		ctx.WriteString(strconv.Itoa(int(c.ID)))
	} else {
		handleErr(ctx, err)
	}
}

func cmOpt(ctx iris.Context) {

}

func cmDel2(ctx iris.Context) {
	ck := ctx.GetCookie("cm_tk")
	id, err := ctx.Params().GetUint("id")
	if err == nil && id > 0 && ck != "" {
		err = db.Delete(&Comment{}, id).Error
	}
	handleErr(ctx, err)
}
func cmDel(ctx iris.Context) {
	id := ctx.URLParam("id")
	if id == "" {
		ctx.StatusCode(200)
	} else {
		ids := strings.Split(id, ",")
		err := db.Delete(&Comment{}, "ID in ?", ids).Error
		handleErr(ctx, err)
		countCm()
		ctx.StatusCode(200)
	}
}
