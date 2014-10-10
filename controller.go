package beegoBaseController

import (
	"fmt"
	"github.com/astaxie/beego"
)

type more []string

func (m *more) Add(s string) string {
	*m = append(*m, s)
	return ``
}

type Base struct {
	beego.Controller
}

type ErrMap map[error]string

func InfoPrepend(s string) func(err error) string {
	return func(err error) string { return fmt.Sprint(s, err) }
}

func (this *Base) NewPaginator(per, nums int) *Paginator {
	paginator := NewPaginator(this.Ctx.Input.Request, per, nums)
	this.Data[`paginator`] = paginator
	return paginator
}
func (b *Base) Prepare() {
	b.Data[`moreStyles`] = &more{}
	b.Data[`beforeScripts`] = &more{}
	b.Data[`laterScripts`] = &more{}
	b.Data["position"] = ``
	b.Data[`title`] = ""
	beego.ReadFromRequest(&b.Controller)

}
