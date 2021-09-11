package errSrv

import (
	"crypto/md5"
	b64 "encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/kataras/iris/v12"
	"strings"
	"time"
)

func authFail(ctx iris.Context) {
	ctx.StatusCode(403)
	d := "auth fail"
	if sys.Token == "" {
		d = "session expired"
	}
	ctx.WriteString(d)
}

func upk(s string) (string, string, string, string, error) {
	s = s[1:]
	st, er := dec(s)
	if er == nil {
		v := strings.Split(st, "\u0001")
		if len(v) == 4 && len(v[0]) > 2 && len(v[1]) > 3 {
			v0, e0 := dec(v[0])
			v1, e1 := dec(v[1])
			v2, e2 := dec(v[2])
			v3, e3 := dec(v[3])
			if e0 == nil && e1 == nil && e2 == nil && e3 == nil {
				return v0, v1, v2, v3, nil
			}
		} else {
			er = errors.New("wrong text")
		}
	}
	return "", "", "", "", er
}

func enc(str string) string {
	return string(b64.StdEncoding.EncodeToString([]byte(str)))
}

func dec(str string) (string, error) {
	des, err := b64.StdEncoding.DecodeString(str)
	return string(des), err
}

func newTk() string {
	tk := enc(uuid.New().String())
	sys.Token = tk
	return tk
}

func handleErr(ctx iris.Context, err error) {
	if err == nil {
		ctx.StatusCode(200)
	} else {
		_ = fmt.Errorf("%v", err)
		ctx.StatusCode(500)
		_, _ = ctx.WriteString(err.Error())
	}
}

func cleanToken() {
	n := now()
	if nextToken < n {
		nextToken = n + day*2
		sys.Token = ""
	}
}

var jobs []func()

func doJobs() {
	for {
		l := len(jobs)
		if l > 0 {
			j := jobs[0]
			jobs = jobs[1:]
			j()
		} else {
			time.Sleep(time.Millisecond * 300)
		}
	}
}

var nextToken = int64(0)
var day = int64(60 * 60 * 24)

func addJob(fn func()) {
	if jobs == nil {
		jobs = make([]func(), 0)
	}
	jobs = append(jobs, fn)
}

func md5Enc(str string) string {
	c := md5.Sum([]byte(str + "err#*&@#1"))
	return hex.EncodeToString(c[0:len(c)])
}
