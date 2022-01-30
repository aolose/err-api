package errSrv

import (
	"github.com/kataras/iris/v12/context"
	"math/rand"
	"net"
	"strconv"
	"strings"
)

// FirewallRule State
// 0001 - skip log
// 0010 - block comment
// 0100 - block access
type FirewallRule struct {
	ID     uint   `gorm:"primarykey" json:"id"`
	IP     string `json:"ip"`
	Path   string `json:"pa"`
	UA     string `json:"ua"`
	Refer  string `json:"rf"`
	Saved  int64  `json:"sv"`
	State  int    `json:"st"`
	Active bool   `json:"at"`
	Remark string `json:"rk"`
	TmpID  int64  `json:"ti" gorm:"-"`
}

const (
	SkipLog = 1 << iota
	BlockLogin
	BlockComment
	BlockAccess
)

func parseCtx(c *context.Context) (net.IP, string, string, string) {
	return net.ParseIP(getIP(c)),
		strings.ToLower(c.Path()),
		strings.ToLower(c.GetHeader("User-Agent")),
		strings.ToLower(c.GetReferrer().String())
}

func firewall(c *context.Context, state int) bool {
	ip, path, ua, refer := parseCtx(c)
	for _, filter := range firewallRules {
		if filter.Active &&
			(filter.hasIp(ip)|
				filter.hasUA(ua)|
				filter.hasRefer(refer)|
				filter.hasPath(path)) == 1 &&
			(filter.State&state) > 0 {
			return true
		}
	}
	return false
}

func (f *FirewallRule) hasIp(p net.IP) int {
	if f.IP != "" {
		_, subnet, _ := net.ParseCIDR(f.IP)
		if subnet != nil && subnet.Contains(p) {
			return 1
		} else {
			if p.Equal(net.ParseIP(f.IP)) {
				return 1
			}
		}
		return -1
	}
	return 0
}

func (f *FirewallRule) hasPath(path string) int {
	if f.Path != "" {
		if strings.HasPrefix(path, strings.ToLower(f.Path)) {
			return 1
		}
		return -1
	}
	return 0
}
func (f *FirewallRule) hasRefer(refer string) int {
	if f.Refer != "" {
		if strings.Contains(refer, strings.ToLower(f.Refer)) {
			return 1
		}
		return -1
	}
	return 0
}
func (f *FirewallRule) hasUA(ua string) int {
	if f.UA != "" {
		if strings.Contains(ua, strings.ToLower(f.UA)) {
			return 1
		}
		return -1
	}
	return 0
}

func BlockIpTemporary(ip string) {
	for _, filter := range firewallRules {
		if filter.Active &&
			filter.hasIp(net.ParseIP(ip)) == 1 &&
			filter.State&BlockLogin > 0 {
			return
		}
	}
	firewallRules = append(firewallRules, &FirewallRule{
		TmpID:  rand.Int63n(1e10),
		IP:     ip,
		Saved:  now(),
		Active: true,
		State:  BlockLogin,
	})
}

type Notice struct {
	ID   uint   `gorm:"primarykey" json:"i"`
	Date int64  `json:"d"`
	Msg  string `json:"m"`
	Read int    `json:"r"`
	Type int    `json:"c"`
	Meta string `json:"t"`
}

type AccessLog struct {
	ID    uint   `gorm:"primarykey" json:"k"`
	Ip    string `gorm:"index" json:"i"`
	Saved int64  `json:"s"`
	Date  int64  `json:"-"`
	Refer string `json:"r"`
	From  string `gorm:"-"  json:"f"`
	Path  string `json:"p"`
	UA    string `json:"u"`
}

type BlackList struct {
	ID     uint   `gorm:"primarykey" json:"k"`
	Saved  int64  `json:"s"`
	IP     string `json:"i"`
	From   string `gorm:"-" json:"f"`
	Type   int    `json:"t"`
	Life   int64  `json:"l"`
	Reason string `json:"r"`
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

type System struct {
	ID            uint   `json:"-"`
	Admin         string `json:"usr"`
	Pwd           string `json:"pwd"`
	LoginProtect  bool   `json:"qal"`
	CommentQA     bool   `json:"qac"`
	CmLen         int    `json:"cLen"`
	CmLife        int64  `json:"cLife"`
	DisCm         int    `json:"disCm"`
	AuditCm       int    `json:"auCm"`
	CnLen         int    `json:"cnLen"`
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
type ArtInf struct {
	Slug  string `json:"s"`
	Title string `json:"t"`
	Date  int64  `json:"d"`
}
type Comment struct {
	ID      uint   `gorm:"primarykey" json:"i"`
	Inf     ArtInf `gorm:"-" json:"x"`
	Avatar  int    `json:"a"`
	Name    string `json:"n"`
	ArtID   uint   `json:"d"`
	Reply   uint   `json:"r"`
	Content string `json:"c"`
	Link    string `json:"l"`
	Status  int    `json:"s"`
	Token   string `json:"-"`
	Saved   int64  `json:"t"`
	Own     int    `json:"o" gorm:"-"`
	IP      string `json:"-"`
	From    string `json:"f"`
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
	AID        uint   `json:"aid" grom:"-"`
	Read       int    `json:"rd"`
	PubContent string `json:"pubCont"`
	AuthorID   uint   `json:"-"`
	Author     Author `json:"author"`
	Pwd        string `json:"pwd"`
	Tags       string `json:"tags"`
	DisCm      int    `json:"disCm"`
	AuditCm    int    `json:"auCm"`
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

//type ArtHis struct {
//	ID      uint  `gorm:"primarykey" json:"id"`
//	AID     uint  ` gorm:"index" json:"aid"`
//	Version int64 `gorm:"index" json:"ver"`
//	Content string
//	Title   string
//}

type Art struct {
	ID             uint   `gorm:"primarykey" json:"id"`
	OverrideUpdate int64  `json:"update2"`
	OverrideCreate int64  `json:"create2"`
	OverrideSlug   string `json:"slug2"`
	Content        string `json:"content"`
	Title          string `json:"title" gorm:"index;not null"`
	SaveAt         int64  `json:"saved"`
	PubArt
}

func save(p *Art, pub bool) error {
	var err error
	o := p
	n := now()
	if !pub {
		p = &Art{
			ID:      p.ID,
			Content: p.Content,
			Title:   p.Title,
			SaveAt:  n,
		}
	} else {
		if p.OverrideCreate != 0 {
			p.Created = p.OverrideCreate
		}
		p.OverrideCreate = 0
	}
	if p.ID == 0 {
		if p.Created == 0 {
			p.Created = n
		}
		err = db.Create(p).Error
	} else {
		if pub && p.Banner == "" {
			db.Model(p).Update("banner", p.Banner)
		}
		err = db.Model(p).Updates(p).Error
	}
	o.ID = p.ID
	o.SaveAt = p.SaveAt
	return err
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
		p.Updated = p.OverrideUpdate
	}
	p.OverrideUpdate = 0
	if p.OverrideSlug != "" {
		p.Slug = p.OverrideSlug
	} else {
		p.Slug = slug(p.Title)
	}
	p.OverrideSlug = ""
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
	//todo: rm article his
	//if err == nil {
	//	err = db.Create(&ArtHis{
	//		AID:     p.ID,
	//		Content: p.Content,
	//		Title:   p.Title,
	//		Version: n,
	//	}).Error
	//}
	//addJob(cleanArtHis(p.ID))
	return err
}

//func cleanArtHis(id uint) func() {
//	return func() {
//		ars := make([]ArtHis, 0)
//		db.Select("id", "a_id", "version").
//			Where("a_id = ?", id).
//			Order("version desc").
//			Limit(10).Find(&ars)
//		ids := make([]uint, len(ars))
//		for a, i := range ars {
//			ids[a] = i.ID
//		}
//		db.Not(ids).Delete(ArtHis{})
//	}
//}

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
	if len(nt) > 0 {
		err = db.Create(&nt).Error
	}
	if err == nil {
		ta := make([]TagArt, 0)
		for _, n := range name {
			ok, v := hasTag(id, n)
			if !ok {
				ta = append(ta, TagArt{AID: id, Name: n})
				tagsCache[n] = append(v, id)
			}
		}
		if len(ta) > 0 {
			err = db.Create(&ta).Error
		}
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
	db.AutoMigrate(&BlackList{})
	db.AutoMigrate(&Art{})
	//db.AutoMigrate(&ArtHis{})
	db.AutoMigrate(&Tag{})
	db.AutoMigrate(&TagArt{})
	db.AutoMigrate(&Res{})
	db.AutoMigrate(&Notice{})
	db.AutoMigrate(&Author{})
	db.AutoMigrate(&Comment{})
	db.AutoMigrate(&Guest{})
	db.AutoMigrate(&AccessLog{})
	db.AutoMigrate(&FirewallRule{})
}
