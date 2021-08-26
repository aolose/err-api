package errSrv

import (
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"strings"
	"time"
)

type CreateDate struct {
	Created   int64          `json:"created"gorm:"autoCreateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

var tagsCache = map[string][]uint{}

type System struct {
	ID            uint
	Admin         string
	Pwd           string
	Token         string
	TotalPubPosts int
	TotalPosts    int
}

type Res struct {
	ID     uint `gorm:"primarykey" json:"-"`
	Type   string
	Name   string
	Remark string
	Pwd    string
	Size   int64
	ArtID  uint
	Arts   []Art
}

type Guest struct {
	ID   uint   `gorm:"primarykey" json:"-"`
	Name string `json:"name"`
	Link string ` json:"link"`
}

type Author struct {
	ID   uint   `gorm:"primarykey" json:"-"`
	Name string `json:"name"`
}

type Comment struct {
	ID uint `gorm:"primarykey" json:"id"`
	CreateDate
	Avatar  int    `json:"avatar"`
	Name    string `json:"name"`
	ArtID   uint   `json:"-"`
	reply   uint   `json:"reply"`
	Content string `json:"content"`
	Link    string `json:"link"`
}

type PubLisArt struct {
	Banner   string `json:"banner"`
	Desc     string `json:"desc"`
	PubTitle string `json:"pubTitle"`
	Slug     string `json:"slug" gorm:"not null;index"`
	Content  string `json:"content" grom:"-"`
}

type PubArt struct {
	PubContent   string `json:"pubCont"`
	AuthorID     uint   `json:"-"`
	Author       Author `json:"author"`
	Ress         []Res  `json:"-"`
	AllowComment int    `json:"comable"`
	Pwd          string `json:"-"`
	Updated      int64  `json:"updated"`
	Created      int64  `json:"created"`
	Tags         string `json:"tags"`
	PubLisArt
}

type Tag struct {
	Name string `gorm:"primarykey" json:"name"`
	Pic  string `json:"pic"`
	Desc string `json:"desc"`
}

type TagArt struct {
	Name string `gorm:"index" json:"name"`
	AID  uint   `gorm:"index" json:"aid"`
}

type ArtHis struct {
	AID     uint  ` gorm:"index" json:"id"`
	Version int64 `gorm:"index" json:"ver"`
	Content string
	Title   string
}

type Art struct {
	ID             uint   `gorm:"primarykey" json:"id"`
	OverrideUpdate int64  `json:"update2"`
	OverrideCreate int64  `json:"create2"`
	OverrideSlug   string `json:"slug2"`
	ResID          uint   `json:"-"`
	Version        int64  `json:"ver"`
	Content        string `json:"content"`
	Title          string `json:"title" gorm:"index;not null"`
	SaveAt         int64  `json:"saved"`
	PubArt
}

func (p *Art) SetPublic(pu PubArt) *Art {
	p.PubArt = pu
	return p
}

func save(p *Art, pub bool) error {
	if !pub {
		p.PubArt = PubArt{}
	}
	n := time.Now().Unix()
	p.SaveAt = n
	if p.ID == 0 {
		if p.Created == 0 {
			p.Created = n
		}
		if p.OverrideCreate != 0 {
			p.Created = p.OverrideCreate
		}
		return db.Create(p).Error
	} else {
		if p.OverrideCreate != 0 {
			p.Created = p.OverrideCreate
		}
		return db.Model(p).Updates(p).Error
	}
}

func (p *Art) Save() error {
	err := save(p, false)
	return err
}

var nTags string
var dTags string

func (p *Art) Publish() error {
	nTags = ""
	dTags = ""
	var err error
	if p.ID != 0 {
		c := &Art{ID: p.ID}
		err = db.Find(c).Error
		if err != nil {
			return err
		}
		if c.Tags != p.Tags {
			a := strings.Split(c.Tags, " ")
			b := strings.Split(p.Tags, " ")
			la := len(a)
			lb := 0
			m := make(map[string]int)
			for _, k := range a {
				m[k] = m[k] - 1
			}
			for _, k := range b {
				if m[k] == -1 {
					la--
				} else {
					lb++
				}
				m[k] = m[k] + 1
			}
			od := make([]string, la)
			ne := make([]string, lb)
			la = 0
			lb = 0
			for k, v := range m {
				if v == -1 {
					od[la] = k
					la++
				}
				if v == 1 {
					ne[lb] = k
					lb++
				}
			}
			err = delTags(p.ID, od...)
			if err == nil {
				err = addTags(p.ID, ne...)
			}
			dTags = strings.Join(od, " ")
			nTags = strings.Join(ne, " ")
		}
	}
	if err != nil {
		return err
	}
	n := time.Now().Unix()
	p.Updated = n
	if p.OverrideUpdate != 0 {
		p.Updated = p.Updated
	}
	if p.OverrideSlug != "" {
		p.Slug = p.OverrideSlug
	} else {
		p.Slug = slug(p.Title)
	}
	c := slugCount(p.Slug, 0)
	if c > 0 {
		if c > 99 {
			c = 99 + rand.Int63n(10000)
		}
	}
	if c > 0 {
		p.Slug += strconv.Itoa(int(c))
	}
	p.PubTitle = p.Title
	p.PubContent = p.Content
	err = save(p, true)
	if err == nil {
		if p.Version == -1 {
			p.Version = n
			err = db.Model(p).Update("version", n).Error
		}
		if err == nil {
			err = db.Where(ArtHis{
				AID:     p.ID,
				Version: p.Version,
			}).Assign(ArtHis{
				Content: p.Content,
				Title:   p.Title,
			}).FirstOrCreate(&ArtHis{}).Error
		}
	}
	return err
}

func hasTag(id uint, name string) (bool, []uint) {
	v, ok := tagsCache[name]
	if !ok {
		v = make([]uint, 0)
	}
	for _, i := range v {
		if i == id {
			return true, v
		}
	}
	return false, v
}

func addTags(id uint, name ...string) error {
	if len(name) == 0 {
		return nil
	}
	var old []Tag
	err := db.Where("name in ?", name).Find(&old).Error
	if err != nil {
		return err
	}
	nt := make([]Tag, len(name)-len(old))
	a := 0
loop:
	for _, n := range name {
		for _, t := range old {
			if t.Name == n {
				continue loop
			}
		}
		nt[a] = Tag{Name: n}
		a++
	}
	err = db.Create(&nt).Error
	if err == nil {
		ta := make([]TagArt, 0)
		for _, n := range name {
			ok, v := hasTag(id, n)
			if !ok {
				ta = append(ta, TagArt{AID: id, Name: n})
				tagsCache[n] = append(v, id)
			}
		}
		err = db.Create(&ta).Error
	}
	return err
}

func delTags(id uint, name ...string) error {
	if len(name) == 0 {
		return nil
	}
	ns := make([]string, 0)
	ts := make([]string, 0)
	for _, t := range name {
		ok, v := hasTag(id, t)
		if ok {
			ts = append(ts, t)
			if len(v) < 2 {
				delete(tagsCache, t)
				ns = append(ns, t)
			} else {
				vv := make([]uint, len(v)-1)
				n := 0
				for _, x := range v {
					if x != id {
						vv[n] = x
						n++
					}
				}
				tagsCache[t] = vv
			}
		}
	}
	var err error
	if len(ns) > 0 {
		err = db.Where("name in ?", ns).Delete(&Tag{}).Error
	}
	if len(ts) > 0 {
		if err == nil {
			err = db.Where("a_id = ? and name in ?", id, ts).Delete(&TagArt{}).Error
		}
	}
	return err
}

func dbInit() {
	db.AutoMigrate(&System{})
	db.AutoMigrate(&Art{})
	db.AutoMigrate(&ArtHis{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&TagArt{})
	db.AutoMigrate(&Res{})
	db.AutoMigrate(&Author{})
	db.AutoMigrate(&Comment{})
	db.AutoMigrate(&Guest{})
}
