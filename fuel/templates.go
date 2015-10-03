package main

import (
	"bytes"
	"text/template"
)

type (
	importTD struct {
		Path string
		Name string
	}

	preludeTD struct {
		Pkg     string
		Imports []importTD
	}

	stateFieldTD struct {
		Name, Type, Path string
	}

	comMethodsTD struct {
		Receiver    string
		StateFields []stateFieldTD
	}

	refFieldTD struct {
		Name, Type string
	}

	childCode struct {
		Code    *bytes.Buffer
		RefName string
		ElTag   string
	}

	refsTD struct {
		ComName string
		Refs    map[string]string
	}

	fieldAssTD struct {
		Name, Value string
	}

	comCreateTD struct {
		Decls            *bytes.Buffer
		ComName, ComType string
		FieldsAss        []fieldAssTD

		ChildrenCode []childCode
	}

	comDefTD struct {
		ComName string
	}

	renderFuncTD struct {
		ComName string
		Return  *bytes.Buffer
		Decls   *bytes.Buffer
		HasRefs bool
	}

	elementVDOMTD struct {
		Tag      string
		Key      string
		Attrs    map[string]string
		Children []childCode
	}

	textNodeVDOMTD struct {
		Text string
	}
)

const (
	childrenVDOMCode = `[[if .]]wade.NewVNodeList([[$last := lastIdx .]]` +
		`[[range $i, $c := .]]` +

		`[[if $c.RefName]](func() vdom.VNode {
		__ret := [[$c.Code]]
		__ret.OnRendered(func(domNode dom.Node) {
			__refs.[[$c.RefName]] = domNode.(dom.[[elDOMType $c.ElTag]])
		})
		return __ret
		})()[[else]][[$c.Code]][[end]]` +
		`[[if lt $i $last]],[[end]]` +
		`[[end]])` +

		`[[else]]nil[[end]]`

	textNodeVDOMCode = `vdom.VText([[.Text]])`

	elementVDOMCode = `` +
		`[[define "attrs"]]` +
		`[[if .Attrs]]` +
		`vdom.Properties{` +
		`[[range $key, $value := .Attrs]]
				"[[$key]]": [[$value]],
			[[end]]` +
		`}[[else]]nil[[end]]` +
		`[[end]]` +

		`vdom.NewElement("[[.Tag]]", wade.Str([[.Key]]), [[template "attrs" .]],` +
		`[[template "children" .Children]])`

	renderFuncCode = `
func [[if .ComName]](this *[[.ComName]])[[end]] VDOMRender() *vdom.VElement {
	[[if .HasRefs]]__refs := vdom.GetComponentData(this).Refs.(*[[.ComName]]Refs)[[end]]
	[[.Decls]]
	return [[.Return]]
}
`

	preludeCode = `package [[.Pkg]]

// THIS FILE IS AUTOGENERATED BY WADE.GO FUEL
// CHANGES WILL BE OVERWRITTEN
import (
[[range .Imports]]
	[[.Name]] "[[.Path]]"
[[end]]
)

func init() {
	_, _, _, _ = fmt.Printf, vdom.NewElement,  wade.Str, dom.GetDocument
}
`

	comMethodsCode = `
[[ $receiver := .Receiver ]]

[[range .StateFields]]
	func (this [[$receiver]]) set[[.Name]](v [[.Type]]) {
		this.[[.Path]] = v
		this.rerender()
	}

	[[if eq .Type "bool"]]
	func (this [[$receiver]]) toggle[[.Name]]() {
		this.[[.Path]] = !this.[[.Path]]
	}
	[[end]]
[[end]]

func (this [[$receiver]]) VDOMChildren() []vdom.VNode {
	return vdom.GetComponentData(this).Children
}

func (this [[$receiver]]) rerender() {
	vdom.RerenderComponent(this)
}
`

	refsCode = `
type [[.ComName]]Refs struct {
[[range $refField, $elTag := .Refs]]
	[[$refField]] dom.[[elDOMType $elTag]]
[[end]]
}

func (this *[[.ComName]]) VDOMCreateRefs() interface{} {
	return &[[.ComName]]Refs{}
}

func (this *[[.ComName]]) Refs() (ret *[[.ComName]]Refs) {
	return vdom.GetComponentData(this).Refs.(*[[.ComName]]Refs)
}`

	comCreateCode = `&vdom.VElement{
	RenderComponent: func(old vdom.Component) *vdom.VElement {
		var com *[[.ComType]]
		var ok bool
		if old != nil {
			com, ok = old.(*[[.ComType]])
		}
		if old == nil || !ok {
			com = &[[.ComType]]{}
		}

		[[range .FieldsAss]]
			com.[[.Name]] = [[.Value]]
		[[end]]
		[[.Decls]]

		return vdom.RenderComponent(com, [[template "children" .ChildrenCode]])
	},
}`

	comDefCode = `
	type [[.ComName]] struct {}
	`
)

func newTpl(name string, code string) *template.Template {
	return template.Must(gTpl.New(name).Parse(code))
}

var (
	funcMap = template.FuncMap{
		"lastIdx": func(l []childCode) int {
			return len(l) - 1
		},
		"elDOMType": func(elTag string) string {
			switch elTag {
			case "input":
				return "InputEl"
			case "form":
				return "FormEl"
			}

			return "Node"
		},
	}

	gTpl            = template.New("root").Delims("[[", "]]").Funcs(funcMap)
	childrenVDOMTpl = newTpl("children", childrenVDOMCode)
	textNodeVDOMTpl = newTpl("txvdom", textNodeVDOMCode)
	elementVDOMTpl  = newTpl("elvdom", elementVDOMCode)
	renderFuncTpl   = newTpl("renderFunc", renderFuncCode)
	preludeTpl      = newTpl("prelude", preludeCode)
	comMethodsTpl   = newTpl("comMethods", comMethodsCode)
	refsTpl         = newTpl("refs", refsCode)
	comCreateTpl    = newTpl("comCreate", comCreateCode)
	comDefTpl       = newTpl("comDef", comDefCode)
)

func newChildTpl(parent *template.Template, name, code string) *template.Template {
	return template.Must(parent.New(name).Parse(code))
}

func addChildTpl(parent *template.Template, name string, child *template.Template) *template.Template {
	return template.Must(parent.AddParseTree(name, child.Tree))
}
