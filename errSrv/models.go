package errSrv

import (
	"math/rand"
	"sort"
	"strconv"
	"strings"
)

type ArtVisitDetail struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	Ip    string `gorm:"index`
	Aid   uint   `gorm:"index`
	Date  uint   `gorm:"index`
	Count int
}

func visitRec(art *Art, ip string) {
	da := uint(now() / 3600 / 24)
	a := &ArtVisitDetail{
		Aid:  art.ID,
		Ip:   ip,
		Date: da,
	}
	db.Where(a).FirstOrCreate(a)
	db.Model(a).Update("count", a.Count+1)
	db.Model(art).Update("read", art.Read+1)
}

type BlCAche struct {
	idx []int
	ips [][]string
}

var blackCache *BlCAche

func (bc *BlCAche) add(ip string) *[]string {
	l0 := len(ip)
	l := len(bc.idx)
	for i := 0; i < l; i++ {
		if bc.idx[i] == l0 {
			bc.ips[i] = append(bc.ips[i], ip)
			return &bc.ips[i]
		}
	}
	bc.idx = append(bc.idx, l0)
	bc.ips = append(bc.ips, []string{ip})
	return nil
}
func (bc *BlCAche) rm(ip string) {
	l0 := len(ip)
	l := len(bc.idx)
	for i := 0; i < l; i++ {
		if bc.idx[i] == l0 {
			for n, p := range bc.ips[i] {
				if p == ip {
					a := bc.ips[n:]
					bc.ips = bc.ips[:n-1]
					for _, v := range a {
						bc.ips = append(bc.ips, v)
					}
					return
				}
			}
			return
		}
	}
}

func (bc *BlCAche) load() {
	bk := make([]BlackList, 0)
	db.Find(&bk)
	for _, b := range bk {
		bc.add(b.IP)
	}
	for _, p := range bc.ips {
		sort.Strings(p)
	}
}

func (bc *BlCAche) has(ip string) bool {
	l := len(ip)
	l1 := len(bc.idx)
	for i := 0; i < l1; i++ {
		if bc.idx[i] == l {
			ls := bc.ips[i]
			l2 := len(ls)
			s := 0
			e := l2
			for n := l2 / 2; n >= s && n < e && e > s; {
				v := ls[n]
				if v == ip {
					return true
				}
				if v > ip {
					s = n + 1
					n = (s + e) / 2
				} else {
					e = n
					n = (s + e) / 2
				}
			}
		}
	}
	return false
}

type BlackList struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	Saved  int64  `json:"saved"`
	IP     string `json:"ip"`
	Type   int    `json:"type"`
	Life   int64  `json:"life"`
	Reason string `json:"reason"`
}

const (
	BkComment = 1 << iota
	BkLogin
)

func isBKType(t int) bool {
	return 0b110&t != 0
}

type BKManager []BlackList

func (b *BKManager) add(bl BlackList) {
	bl.Saved = now()
	db.Create(bl)
	s := blackCache.add(bl.IP)
	if s != nil {
		sort.Strings(*s)
	}
}

func (b *BKManager) rm(id int) {
	db.Delete(&BlackList{}, id)
	blackCache.load()
}

type ListPubPost struct {
	Posts []PubLisArt `json:"ls"`
	Total int         `json:"total"`
	Cur   int         `json:"cur"`
}
type ListRes struct {
	List  []Res `json:"ls"`
	Total int   `json:"total"`
	Cur   int   `json:"cur"`
}

type ListPost struct {
	Posts []Art `json:"ls"`
	Total int   `json:"total"`
	Cur   int   `json:"cur"`
}

var tagsCache = map[string][]uint{}

type RQA struct {
	ID    uint   `gorm:"primarykey" json:"id"`
	Q     string `gorm:"index" json:"q"`
	A     string `json:"a"`
	Saved int64  `json:"saved"`
}

type Qa struct {
	RQA
	Params string `json:"p"`
}

type System struct {
	ID            uint   `json:"-"`
	Admin         string `json:"usr"`
	Pwd           string `json:"pwd"`
	LoginProtect  bool   `json:"qal"`
	CommentQA     bool   `json:"qac"`
	Token         string `gorm:"-" json:"-"`
	TotalPubPosts int    `gorm:"-" json:"-"`
	TotalPosts    int    `gorm:"-" json:"-"`
	TotalRes      int    `gorm:"-" json:"-"`
}

type Res struct {
	Date int64  `json:"date"`
	Type string `json:"type"`
	Ext  string `json:"ext"`
	ID   string `json:"id" gorm:"primarykey" `
	Name string `json:"name"`
	Size int64  `json:"size"`
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
	ID      uint   `gorm:"primarykey" json:"id"`
	Avatar  int    `json:"avatar"`
	Name    string `json:"name"`
	ArtID   uint   `json:"-"`
	Reply   uint   `json:"reply"`
	Content string `json:"content"`
	Link    string `json:"link"`
}

type PubLisArt struct {
	Banner   string `json:"banner"`
	Desc     string `json:"desc"`
	PubTitle string `json:"title"`
	Slug     string `json:"slug" gorm:"not null;index"`
	Content  string `json:"content" grom:"-"`
	Updated  int64  `json:"updated"`
	Created  int64  `json:"created"`
}

type PubArt struct {
	Read         int    `json:"rd"`
	PubContent   string `json:"pubCont"`
	AuthorID     uint   `json:"-"`
	Author       Author `json:"author"`
	AllowComment int    `json:"cm"`
	Pwd          string `json:"pwd"`
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
	n := now()
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
		if pub {
			db.Model(p).Update("banner", p.Banner)
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
	n := now()
	p.Updated = n
	if p.OverrideUpdate != 0 {
		p.Updated = p.Updated
	}
	if p.OverrideSlug != "" {
		p.Slug = p.OverrideSlug
	} else {
		p.Slug = slug(p.Title)
	}
	c := slugCount(p.Slug, p.ID)
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
	db.AutoMigrate(&Qa{})
	db.AutoMigrate(&BlackList{})
	db.AutoMigrate(&Art{})
	db.AutoMigrate(&ArtHis{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&TagArt{})
	db.AutoMigrate(&Res{})
	db.AutoMigrate(&Author{})
	db.AutoMigrate(&Comment{})
	db.AutoMigrate(&Guest{})
	db.AutoMigrate(&ArtVisitDetail{})
}
