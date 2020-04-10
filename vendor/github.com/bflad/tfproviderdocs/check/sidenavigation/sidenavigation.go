package sidenavigation

import (
	"fmt"
	"io/ioutil"
	"log"
	"regexp"
	"strings"

	"golang.org/x/net/html"
)

const erbRegexpPattern = `<%.*%>`

type SideNavigation struct {
	Categories []*Category
	Node       *html.Node

	ExternalLinks []*Link

	// These include all sub-categories
	DataSourceLinks []*Link
	GuideLinks      []*Link
	ResourceLinks   []*Link
}

func Find(node *html.Node) *SideNavigation {
	if node == nil {
		return nil
	}

	if isSideNavigationList(node) {
		return New(node)
	}

	for child := node.FirstChild; child != nil; child = child.NextSibling {
		sideNavigation := Find(child)

		if sideNavigation != nil {
			return sideNavigation
		}
	}

	return nil
}

func FindFile(path string) (*SideNavigation, error) {
	fileContents, err := ioutil.ReadFile(path)

	if err != nil {
		log.Fatalf("error opening file (%s): %w", path, err)
	}

	return FindString(string(fileContents))
}

func FindString(fileContents string) (*SideNavigation, error) {
	strippedFileContents := regexp.MustCompile(erbRegexpPattern).ReplaceAllString(string(fileContents), "")

	doc, err := html.Parse(strings.NewReader(strippedFileContents))

	if err != nil {
		return nil, fmt.Errorf("error HTML parsing file: %w", err)
	}

	return Find(doc), nil
}

func New(node *html.Node) *SideNavigation {
	sn := &SideNavigation{
		Categories:      make([]*Category, 0),
		DataSourceLinks: make([]*Link, 0),
		ExternalLinks:   make([]*Link, 0),
		GuideLinks:      make([]*Link, 0),
		Node:            node,
		ResourceLinks:   make([]*Link, 0),
	}

	sn.processList()

	return sn
}

func (sn *SideNavigation) HasDataSourceLink(name string) bool {
	for _, link := range sn.DataSourceLinks {
		if link.Text == name {
			return true
		}
	}

	return false
}

func (sn *SideNavigation) HasResourceLink(name string) bool {
	for _, link := range sn.ResourceLinks {
		if link.Text == name {
			return true
		}
	}

	return false
}

func (sn *SideNavigation) processList() {
	for child := sn.Node.FirstChild; child != nil; child = child.NextSibling {
		if isListItem(child) {
			sn.processListItem(child)
		}
	}
}

func (sn *SideNavigation) processListItem(node *html.Node) {
	linkNode := childLinkNode(node)

	if linkNode == nil {
		return
	}

	link := NewLink(linkNode)
	listNode := childUnorderedListNode(node)

	if listNode == nil {
		switch link.Type {
		case LinkTypeDataSource:
			sn.DataSourceLinks = append(sn.DataSourceLinks, link)
		case LinkTypeExternal:
			sn.ExternalLinks = append(sn.ExternalLinks, link)
		case LinkTypeGuide:
			sn.GuideLinks = append(sn.GuideLinks, link)
		case LinkTypeResource:
			sn.ResourceLinks = append(sn.ResourceLinks, link)
		}

		return
	}

	category := NewCategory(link.Text, listNode)

	sn.DataSourceLinks = append(sn.DataSourceLinks, category.DataSourceLinks...)
	sn.ExternalLinks = append(sn.ExternalLinks, category.ExternalLinks...)
	sn.GuideLinks = append(sn.GuideLinks, category.GuideLinks...)
	sn.ResourceLinks = append(sn.ResourceLinks, category.ResourceLinks...)
}

func isSideNavigationList(node *html.Node) bool {
	if node.Type != html.ElementNode {
		return false
	}

	if node.Data != "ul" {
		return false
	}

	for _, attr := range node.Attr {
		if attr.Key == "class" && strings.Contains(attr.Val, "docs-sidenav") {
			return true
		}
	}

	return false
}
