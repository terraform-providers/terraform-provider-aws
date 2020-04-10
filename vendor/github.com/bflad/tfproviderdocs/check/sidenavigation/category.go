package sidenavigation

import (
	"golang.org/x/net/html"
)

type Category struct {
	Name string
	Node *html.Node

	DataSourceLinks []*Link
	ExternalLinks   []*Link
	GuideLinks      []*Link
	ResourceLinks   []*Link
}

func NewCategory(name string, listNode *html.Node) *Category {
	category := &Category{
		DataSourceLinks: make([]*Link, 0),
		ExternalLinks:   make([]*Link, 0),
		GuideLinks:      make([]*Link, 0),
		Name:            name,
		Node:            listNode,
		ResourceLinks:   make([]*Link, 0),
	}

	category.processList()

	return category
}

func (category *Category) processList() {
	for child := category.Node.FirstChild; child != nil; child = child.NextSibling {
		if isListItem(child) {
			category.processListItem(child)
		}
	}
}

func (category *Category) processListItem(node *html.Node) {
	linkNode := childLinkNode(node)

	if linkNode == nil {
		return
	}

	link := NewLink(linkNode)
	listNode := childUnorderedListNode(node)

	if listNode == nil {
		switch link.Type {
		case LinkTypeDataSource:
			category.DataSourceLinks = append(category.DataSourceLinks, link)
		case LinkTypeExternal:
			category.ExternalLinks = append(category.ExternalLinks, link)
		case LinkTypeGuide:
			category.GuideLinks = append(category.GuideLinks, link)
		case LinkTypeResource:
			category.ResourceLinks = append(category.ResourceLinks, link)
		}

		return
	}

	// Categories can contain one single subcategory (e.g. Service Name > Resources)
	subCategory := NewCategory(link.Text, listNode)

	category.DataSourceLinks = append(category.DataSourceLinks, subCategory.DataSourceLinks...)
	category.ExternalLinks = append(category.ExternalLinks, subCategory.ExternalLinks...)
	category.GuideLinks = append(category.GuideLinks, subCategory.GuideLinks...)
	category.ResourceLinks = append(category.ResourceLinks, subCategory.ResourceLinks...)
}
