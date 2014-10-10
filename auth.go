package beegoBaseController

import (
	"encoding/base64"
	"strings"
)

func (this *Base) CheckBaseAuth(cred string) bool {
	this.Ctx.Output.Header("WWW-Authenticate", `Basic realm="Authenticate"`)
	s := strings.SplitN(this.Ctx.Input.Header("Authorization"), " ", 2)
	if len(s) != 2 {
		return false
	}

	b, err := base64.StdEncoding.DecodeString(s[1])
	if err != nil {
		return false
	}

	pair := strings.SplitN(string(b), ":", 2)
	if len(pair) != 2 {
		return false
	}

	user := strings.Split(cred, ":")
	return pair[0] == user[0] && pair[1] == user[1]
}
