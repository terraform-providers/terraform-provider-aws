package sidenavigation

import (
	"fmt"
	"strings"

	"golang.org/x/net/html"
)

type Link struct {
	Text string
	Type LinkType
	Url  string
}

type LinkType int

const (
	LinkTypeAnchor LinkType = iota
	LinkTypeExternal
	LinkTypeDataSource
	LinkTypeGuide
	LinkTypeResource
)

func (typ LinkType) String() string {
	return [...]string{"Anchor", "External", "Data Source", "Guide", "Resource"}[typ]
}

func NewLink(node *html.Node) *Link {
	link := &Link{
		Text: childText(node),
	}

	for _, attr := range node.Attr {
		if attr.Key == "href" {
			link.Url = attr.Val

			if strings.HasPrefix(link.Url, "#") {
				link.Type = LinkTypeAnchor
			} else if strings.Contains(link.Url, "/d/") {
				link.Type = LinkTypeDataSource
			} else if strings.Contains(link.Url, "/guides/") {
				link.Type = LinkTypeGuide
			} else if strings.Contains(link.Url, "/r/") {
				link.Type = LinkTypeResource
			} else {
				link.Type = LinkTypeExternal
			}

			break
		}
	}

	return link
}

func (link *Link) Validate(expectedProviderName string) error {
	if link.Type != LinkTypeDataSource && link.Type != LinkTypeResource {
		return nil
	}

	var linkTypeUrlPart string

	switch link.Type {
	case LinkTypeDataSource:
		linkTypeUrlPart = "d"
	case LinkTypeResource:
		linkTypeUrlPart = "r"
	}

	if !strings.Contains(link.Url, "/") {
		return fmt.Errorf("link URL (%s) is missing / separators, should be in form: /docs/providers/PROVIDER/%s/NAME.html", link.Url, linkTypeUrlPart)
	}

	urlParts := strings.Split(link.Url, "/")

	if len(urlParts) < 6 {
		return fmt.Errorf("link URL (%s) is missing path parts, e.g. /docs/providers/PROVIDER/%s/NAME.html", link.Url, linkTypeUrlPart)
	} else if len(urlParts) > 6 {
		return fmt.Errorf("link URL (%s) has too many path parts, should be in form: /docs/providers/PROVIDER/%s/NAME.html", link.Url, linkTypeUrlPart)
	}

	urlProviderName := urlParts[3]

	if expectedProviderName != "" && urlProviderName != expectedProviderName {
		return fmt.Errorf("link URL (%s) has incorrect provider name (%s), expected: %s", link.Url, urlProviderName, expectedProviderName)
	}

	urlResourceName := urlParts[len(urlParts)-1]
	urlResourceName = strings.TrimSuffix(urlResourceName, ".html")

	if expectedProviderName != "" {
		expectedText := fmt.Sprintf("%s_%s", expectedProviderName, urlResourceName)
		if link.Text != expectedText {
			return fmt.Errorf("link URL (%s) has incorrect text (%s), expected: %s", link.Url, link.Text, expectedText)
		}
	} else {
		expectedSuffix := fmt.Sprintf("_%s", urlResourceName)
		if !strings.HasSuffix(link.Text, expectedSuffix) {
			return fmt.Errorf("link URL (%s) has incorrect text (%s), expected: PROVIDER%s", link.Url, link.Text, expectedSuffix)
		}
	}

	return nil
}
