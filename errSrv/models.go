package errSrv

import (
	"gorm.io/gorm"
)

type CreateDate struct {
	Created   int64          `json:"created"gorm:"autoCreateTime:milli"`
	Updated   int64          `json:"updated"gorm:"autoCreateTime:milli"`
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
	CreateDate
}

type DraftPost struct {
	DraftTitle   string `json:"draft_title"`
	DraftContent string `json:"draft_content"`
	DraftRess    []Res
}

type Post struct {
	ID     uint `gorm:"primarykey" json:"-"`
	ResID  uint `json:"-"`
	Status int  `json:"-"`
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
func (p *Post) GetDraft() *DraftPost {
	return &p.DraftPost
}

func dbInit() {
	db.AutoMigrate(&Post{})
	db.AutoMigrate(&Comment{})
	db.AutoMigrate(&Guest{})
}
