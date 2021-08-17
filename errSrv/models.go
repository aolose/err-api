package errSrv

import (
	"gorm.io/gorm"
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
	ID uint `gorm:"primarykey" json:"-"`
	CreateDate
	Name     string
	PostID   uint
	Post     Post
	Mail     uint
	Content  uint
	Token    string
	Comments []Comment `gorm:"foreignKey:ID"`
}

type PublicPost struct {
	Content  string `json:"content"`
	AuthorID uint   `json:"-"`
	Author   Author `json:"author"`
	Ress     []Res  `json:"-"`
	Pwd      string `json:"-"`
	Slug     string `json:"slug"`
	Title    string `json:"title"`
	Updated  int64  `json:"updated"gorm:"autoCreateTime:milli"`
	CreateDate
}
type EditPost struct {
	ID      uint   `json:"id"`
	Content string `json:"content"`
	Res     []uint `json:"res"`
	Pwd     string `json:"pwd"`
	Slug    string `json:"slug"`
	Title   string `json:"title"`
	Status  int    `json:"status"`
	Updated int64  `json:"updated"`
}

type DraftPost struct {
	DraftSlug    string `json:"slug"`
	DraftTitle   string `json:"draft_title"`
	DraftContent string `json:"draft_content"`
	DraftRess    []Res  `json:"-"`
	Saved        int64  `json:"updated"gorm:"autoCreateTime:milli"`
}

type Post struct {
	ID     uint `gorm:"primarykey" json:"id"`
	ResID  uint `json:"-"`
	Status int  `json:"status"`
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
func (p *EditPost) ToPub() *Post {
	pp := &Post{}
	return pp
}
func (p *EditPost) ToDraft() *Post {
	pp := &Post{}
	return pp
}
func (p *Post) GetEdit() *EditPost {
	e := &EditPost{
		ID:      p.ID,
		Title:   p.Title,
		Content: p.Content,
		Pwd:     p.Pwd,
		Res:     make([]uint, len(p.Ress)),
		Status:  p.Status,
		Slug:    p.Slug,
		Updated: p.Updated,
	}
	if p.DraftContent != "" {
		e.Content = p.DraftContent
	}
	if p.Status != 1 {
		e.Res = make([]uint, len(p.DraftRess))
		e.Title = p.DraftTitle
		e.Content = p.DraftContent
		e.Slug = p.DraftSlug
		e.Updated = p.Saved
		for i, r := range p.DraftRess {
			e.Res[i] = r.ID
		}
	} else {
		for i, r := range p.Ress {
			e.Res[i] = r.ID
		}
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
