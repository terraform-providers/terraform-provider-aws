package sidenavigation

import (
	"golang.org/x/net/html"
)

func childLinkNode(node *html.Node) *html.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if isLink(child) {
			return child
		}
	}

	return nil
}

func childText(node *html.Node) string {
	if node.Type == html.TextNode {
		return node.Data
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		return childText(child)
	}

	return ""
}

func childUnorderedListNode(node *html.Node) *html.Node {
	for child := node.FirstChild; child != nil; child = child.NextSibling {
		if isUnorderedList(child) {
			return child
		}
	}

	return nil
}

func isLink(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}

	return node.Data == "a"
}

func isListItem(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}

	return node.Data == "li"
}

func isUnorderedList(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}

	return node.Data == "ul"
}
