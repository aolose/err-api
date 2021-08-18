package errSrv

import (
	"gorm.io/gorm"
	"time"
)

type CreateDate struct {
	Created   int64          `json:"created"gorm:"autoCreateTime:milli"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`
}

type System struct {
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
	Path   string
	Size   int64
	PostID uint
	Posts  []Post
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
	PostID  uint   `json:"-"`
	reply   uint   `json:"reply"`
	Content string `json:"content"`
	Link    string `json:"link"`
}

type PublicPost struct {
	Content  string `json:"content"`
	AuthorID uint   `json:"-"`
	Author   Author `json:"author"`
	Ress     []Res  `json:"-"`
	CanCom   int    `json:"comm"`
	Pwd      string `json:"-"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Updated  int64  `json:"updated"`
	Created  int64  `json:"created"`
	Banner   string `json:"banner"`
	Desc     string `json:"desc"`
	Tags     []Tag  `json:"tags"`
	TagID    uint   `json:"-"`
}
type EditPost struct {
	ID        uint   `json:"id"`
	Content   string `json:"content"`
	Res       []Res  `json:"res"`
	Pwd       string `json:"pwd"`
	Slug      string `json:"slug"`
	Title     string `json:"title"`
	Updated   int64  `json:"updated"`
	Updated2  int64  `json:"update2"`
	Create2   int64  `json:"create2"`
	CanCom    int    `json:"comm"`
	Published int    `json:"publish"`
	Draft     int    `json:"draft"`
	Banner    string `json:"banner"`
	Desc      string `json:"desc"`
	Tags      []Tag  `json:"tags"`
}

type DraftPost struct {
	DraftTitle   string `json:"draft_title"`
	DraftContent string `json:"draft_content"`
	DraftRess    []Res  `json:"-"`
	DraftUpdate  int64  `json:"updated"`
}

type Tag struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	PostID uint   `json:"-"`
	Posts  []Post `json:"-"`
}

type Post struct {
	ID             uint           `gorm:"primarykey" json:"id"`
	OverrideUpdate int64          `json:"-"`
	OverrideCreate int64          `json:"-"`
	DeletedAt      gorm.DeletedAt `gorm:"index" json:"-"`
	Publish        int            `json:"publish"`
	Draft          int            `json:"draft"`
	ResID          uint           `json:"-"`
	DraftPost
	PublicPost
}

func (p *Post) SetPublic(pu PublicPost) *Post {
	p.PublicPost = pu
	return p
}
func (p *Post) SetDraft(pu DraftPost) *Post {
	p.DraftPost = pu
	return p
}

func (p *Post) GetPublic() *PublicPost {
	return &p.PublicPost
}
func (p *Post) Save() error {
	n := time.Now().Unix() * 1e3
	if p.Created == 0 {
		p.Created = n
	}
	if p.Publish == 1 {
		if p.OverrideUpdate != 0 {
			p.Updated = p.OverrideUpdate
		} else if p.Updated == 0 {
			p.Updated = n
		}
	} else {
		if p.OverrideUpdate != 0 {
			p.DraftUpdate = p.OverrideUpdate
		} else if p.DraftUpdate == 0 {
			p.DraftUpdate = n
		}
	}
	if p.OverrideCreate != 0 {
		p.Created = p.OverrideCreate
	}
	if p.ID != 0 {
		return db.Create(p).Error
	} else {
		return db.Save(p).Error
	}
}

func (p *EditPost) GetPost() *Post {
	pp := &Post{}
	if p.ID != 0 {
		db.First(pp, p.ID)
	}
	return pp
}

func ToPub(pp *Post, p *EditPost) {
	pp.Publish = 1
	pp.Draft = 0
	pp.Content = p.Content
	pp.Title = p.Title
	pp.Ress = p.Res
	pp.DraftPost = DraftPost{}
}
func ToDraft(pp *Post, p *EditPost) {
	pp.OverrideUpdate = p.Updated2
	pp.OverrideCreate = p.Create2
	pp.DraftContent = p.Content
	pp.DraftTitle = p.Title
	pp.DraftRess = p.Res

	pp.Banner = p.Banner
	pp.Desc = p.Desc
	pp.Pwd = p.Pwd
	pp.Slug = p.Slug
	pp.CanCom = p.CanCom
	pp.Tags = p.Tags
}

func (p *EditPost) Publish() error {
	pp := p.GetPost()
	ToDraft(pp, p)
	ToPub(pp, p)
	return pp.Save()
}
func (p *EditPost) Save() error {
	pp := p.GetPost()
	ToDraft(pp, p)
	return pp.Save()
}
func (p *EditPost) Unpublish() error {
	if p.ID == 0 {
		return nil
	}
	pp := p.GetPost()
	pp.Draft = 1
	pp.Publish = 0
	pp.PublicPost = PublicPost{}
	return pp.Save()
}

func (p *Post) GetEdit() *EditPost {
	e := &EditPost{
		ID:        p.ID,
		Title:     p.Title,
		Content:   p.Content,
		Pwd:       p.Pwd,
		Slug:      p.Slug,
		Updated:   p.Updated,
		Res:       p.Ress,
		CanCom:    p.CanCom,
		Banner:    p.Banner,
		Desc:      p.Desc,
		Published: p.Publish,
		Draft:     p.Draft,
		Tags:      p.Tags,
	}
	if p.Draft == 1 {
		e.Title = p.DraftTitle
		e.Content = p.DraftContent
		e.Updated = p.DraftUpdate
		e.Res = p.DraftRess
	}
	return e
}
func (p *Post) GetDraft() *DraftPost {
	return &p.DraftPost
}

func dbInit() {
	db.AutoMigrate(&Post{})
	db.AutoMigrate(&Comment{})
	db.AutoMigrate(&Guest{})
}
