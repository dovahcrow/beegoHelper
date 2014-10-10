beegoHelper
===========

创建beego控制器时候的一个辅助控制器，方便开发.

使用方法： 直接嵌入即可，即

	import (
		"github.com/doomsplayer/beegoHelper"
	)
	
	type MainController struct {
		beegoHelper.Base
	}
	
	func (this *MainController) Get() {
		.........
	}

内容包括：
Parse 系列:

> `this.GetInts` 函数, 返回一个[]int

> `this.ParseJson` 直接将body当成一个json，parse到对象中。方便api交互（尤其是angular）

> `this.GetParamString`, `this.GetParamInt` 是 `this.Ctx.Input.Param()` 的缩短版

> `this.ParseFormAndValid`, 直接解析内容到结构体,并且按照valid的tag来验证struct,详细请查看beego文档`form验证`一节

> `this.ParseQuery`, 实验函数,解析query直接返回一个orm的Cond

Paginator系列:分页器
使用以后会在url最后加一个?p=xxx表示页码

> `this.Paginator(每页几个元素,一共有几个元素)` 生成一个分页器,方便数据库查询以及分页,有以下方法:

>> `Pages() []int` 返回页码数组

>> `PageLink(page int) string` 根据页码生成某一页的链接

>> `PageLinkPrev() (link string)`

>> `PageLinkLast() (link string)`

>> `HasPrev() bool`

>> `HasNext() bool`

>> `IsActive(page int) bool`

>> `Offset() int`

>> `HasPages() bool`

>> `End() int`

Auth系列:HTTPBaseAuth

> `CheckBaseAuth(cred string) bool` cred是验证字符串,格式 `用户名:密码` 只要使用了这个函数会自动加载authorization的头.

Check,Ok 系列
to be continue
