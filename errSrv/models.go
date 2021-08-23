package errSrv

import (
	"gorm.io/gorm"
	"math/rand"
	"strconv"
	"time"
)

type CreateDate struct {
	Created   int64          `json:"created"gorm:"autoCreateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

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
	Banner  string `json:"banner"`
	Desc    string `json:"desc"`
	Title   string `json:"title" grom:"-"`
	Slug    string `json:"slug" gorm:"not null;index"`
	Content string `json:"content" grom:"-"`
}

type PubArt struct {
	AuthorID     uint   `json:"-"`
	Author       Author `json:"author"`
	Ress         []Res  `json:"-"`
	AllowComment int    `json:"comable"`
	Pwd          string `json:"-"`
	Updated      int64  `json:"updated"`
	Created      int64  `json:"created"`
	Tags         []Tag  `json:"tags"`
	TagID        uint   `json:"-"`
	PubLisArt
}

type Tag struct {
	ID    uint  `gorm:"primarykey" json:"id"`
	ArtID uint  `json:"-"`
	Arts  []Art `json:"-"`
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

func save(p *Art) error {
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
	err := save(p)
	return err
}

func (p *Art) Publish() error {
	n := time.Now().Unix()
	p.Updated = n
	if p.OverrideUpdate != 0 {
		p.Updated = p.Updated
	}
	if p.OverrideSlug != "" {
		p.Slug = p.OverrideSlug
	} else {
		p.Slug = trans(p.Title)
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
	err := save(p)
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

func dbInit() {
	db.AutoMigrate(&System{})
	db.AutoMigrate(&Art{})
	db.AutoMigrate(&ArtHis{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&Res{})
	db.AutoMigrate(&Author{})
	db.AutoMigrate(&Comment{})
	db.AutoMigrate(&Guest{})
}
