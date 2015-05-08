package browser

import (
	"fmt"
	"strings"

	"github.com/gopherjs/gopherjs/js"

	"github.com/gowade/wade"
	"github.com/gowade/wade/vdom"
)

var (
	document *js.Object
)

type DOMInputEl struct{ *js.Object }

func (e DOMInputEl) Value() string {
	return e.Get("value").String()
}

func (e DOMInputEl) SetValue(value string) {
	e.Set("value", value)
}

func init() {
	if js.Global == nil || js.Global.Get("document") == js.Undefined {
		panic("This package is only available in browser environment.")
	}

	wade.SetDOMDriver(driver{})
	document = js.Global.Get("document")
}

func ElementById(id string) vdom.DOMNode {
	elem := document.Call("getElementById", id)
	if elem == js.Undefined || elem == nil {
		panic(fmt.Sprintf("No element with id %v found", id))
	}

	return DOMNode{elem}
}

type driver struct{}

func (d driver) PerformDiff(a, b vdom.Node, dNode vdom.DOMNode) {
	vdom.PerformDiff(a, b, dNode.(DOMNode))
}

func (d driver) ToInputEl(el vdom.DOMNode) vdom.DOMInputEl {
	return DOMInputEl{el.(DOMNode).Object}
}

func createElement(tag string) *js.Object {
	return document.Call("createElement", tag)
}

func createTextNode(data string) *js.Object {
	return document.Call("createTextNode", data)
}

type DOMNode struct {
	*js.Object
}

func (d DOMNode) Child(i int) vdom.DOMNode {
	return DOMNode{d.Get("childNodes").Index(i)}
}

func renderNew(node vdom.Node) *js.Object {
	if !node.IsElement() {
		return createTextNode(node.NodeData())
	}

	return renderTo(node, createElement(node.NodeData()))
}

func renderTo(node vdom.Node, d *js.Object) *js.Object {
	oe := node.(*vdom.Element)
	e := oe.Render().(*vdom.Element)
	for attr, v := range e.Attrs {
		if vdom.IsEvent(attr) {
			d.Set(strings.ToLower(attr), v)
			continue
		}

		switch v := v.(type) {
		case bool:
			if v {
				d.Call("setAttribute", attr, attr)
			}
		case string:
			d.Call("setAttribute", attr, v)
		default:
			d.Call("setAttribute", attr, fmt.Sprint(v))
		}
	}

	for _, c := range e.Children {
		if c != nil {
			d.Call("appendChild", renderNew(c))
		}
	}

	e.SetRenderedDOMNode(DOMNode{d})
	if oe != e {
		oe.SetRenderedDOMNode(DOMNode{d})
	}

	return d
}

func (d DOMNode) Render(content vdom.Node, root bool) {
	if !root {
		d.Get("parentNode").Call("replaceChild", renderNew(content), d.Object)
	} else {
		renderTo(content, d.Object)
	}
}

func (dNode DOMNode) Do(action vdom.Action) {
	d := dNode.Object

	switch action.Type {
	case vdom.Deletion:
		d.Call("removeChild", action.Element.(DOMNode).Object)
	case vdom.Insertion:
		insertee := renderNew(action.Content)
		if action.Index == -1 {
			d.Call("appendChild", insertee)
		} else {
			d.Call("insertBefore", insertee, d.Get("childNodes").Index(action.Index))
		}
	case vdom.Move:
		d.Call("insertBefore", action.Element.(DOMNode).Object, d.Get("childNodes").Index(action.Index))
	}
}

func (dNode DOMNode) RemoveAttr(attr string) {
	dNode.Object.Call("removeAttribute", attr)
}

func (dNode DOMNode) SetProp(prop string, value interface{}) {
	dNode.Object.Set(prop, value)
}

func (dNode DOMNode) SetAttr(attr string, value interface{}) {
	d := dNode.Object

	var vstr string
	switch v := value.(type) {
	case bool:
		if !v {
			if d.Call("hasAttribute", attr).Bool() {
				d.Call("removeAttribute", attr)
			}

			return
		} else {
			vstr = attr
		}

	case string:
		vstr = v
	default:
		vstr = fmt.Sprint(v)
	}

	d.Call("setAttribute", attr, vstr)
}
