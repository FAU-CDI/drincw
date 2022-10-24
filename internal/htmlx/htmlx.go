package htmlx

import (
	"strings"

	"golang.org/x/net/html"
)

func ReplaceLinks(source string, replace func(string) string) string {
	var builder strings.Builder
	builder.Grow(len(source))

	nodes, err := html.ParseFragment(strings.NewReader(source), nil)
	if err != nil {
		panic(err)
	}
	for _, node := range nodes {
		replaceNode(node, replace)
		html.Render(&builder, node)
	}
	return builder.String()
}

func replaceNode(node *html.Node, replace func(string) string) {
	if node.Type == html.ElementNode && node.Data == "a" {
		replaceAttr(node.Attr, "href", replace)
	}
	if node.Type == html.ElementNode && node.Data == "img" {
		replaceAttr(node.Attr, "src", replace)
	}

	if node.FirstChild == nil {
		return
	}

	child := node.FirstChild
	replaceNode(child, replace)

	for child.NextSibling != nil {
		child = child.NextSibling
		replaceNode(child, replace)
	}

}

func replaceAttr(attr []html.Attribute, key string, replace func(string) string) {
	for i, a := range attr {
		if a.Key == key {
			attr[i].Val = replace(a.Val)
			break
		}
	}
}
