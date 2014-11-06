package beegoHelper

import (
	"encoding/json"
	"fmt"
	"github.com/astaxie/beego/orm"
	"github.com/astaxie/beego/validation"
	"regexp"
	"strconv"
)

func (this *Base) ParseFormAndValidCheckJson(form interface{}, mapping interface{}) {
	err := this.ParseFormAndValid(form)
	this.CheckJson(err, 400, mapping)
}

func (this *Base) ParseFormAndValid(form interface{}) (err error) {
	err = this.ParseForm(form)
	if err != nil {
		return err
	}

	valid := validation.Validation{}
	b, err := valid.Valid(form)
	if err != nil {
		return err
	}
	if !b {
		errmsg := ``
		for _, err := range valid.Errors {
			errmsg += fmt.Sprint(err.Key, " ", err.Message, "\n")
		}
		return fmt.Errorf(errmsg)
	}
	return nil
}
func (this *Base) ParseJson(obj interface{}) (err error) {
	dec := json.NewDecoder(this.Ctx.Input.Request.Body)
	err = dec.Decode(obj)
	return
}
func (this *Base) GetParamInt(field string) (int, error) {
	param := this.Ctx.Input.Param(field)
	ret, err := strconv.Atoi(param)
	return ret, err
}
func (this *Base) GetParamString(field string) string {
	return this.Ctx.Input.Param(field)
}
func (this *Base) GetParamBool(field string) (bool, error) {
	param := this.Ctx.Input.Param(field)
	ret, err := strconv.ParseBool(param)
	return ret, err
}
func (this *Base) GetInts(field string) []int {
	ss := this.GetStrings(field)
	retv := []int{}
	for _, v := range ss {
		r, _ := strconv.Atoi(v)
		retv = append(retv, r)
	}
	return retv
}
func (this *Base) ParseQuery(fields ...string) *orm.Condition {
	cond := orm.NewCondition()
	extractor := regexp.MustCompile(`^(.+?)\.(.+?)\.(.+?)$`)
	for _, v := range fields {
		entity := this.GetString("q." + v)
		qs := extractor.FindStringSubmatch(entity)
		if len(qs) != 4 {
			continue
		}
		switch qs[2] {
		case "contains", "icontains", "exact", "iexact":
			{
				v += "__" + qs[2]
			}
		default:
			{
				continue
			}
		}
		switch qs[1] {
		case "and":
			{
				cond = cond.And(v, qs[3])
			}
		case "or":
			{
				cond = cond.Or(v, qs[3])
			}
		case "andnot":
			{
				cond = cond.AndNot(v, qs[3])
			}
		case "ornot":
			{
				cond = cond.OrNot(v, qs[3])
			}
		default:
			{
				continue
			}
		}

	}

	return cond
}
