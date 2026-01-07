package types

import (
	"maps"
	"regexp"
	"slices"
	"strings"
)

var siteGroupSuffix = regexp.MustCompile(`:([A-Za-z0-9_][A-Za-z0-9_,]*)$`)

type PortName struct {
	Category string
	Name     string
}

type GitHubInfo struct {
	Account string `json:"account"`
	Project string `json:"project"`
	TagName string `json:"tagName"`
	SubDir  string `json:"account"`
}

type TaggedList struct {
	Items []string
}

type PortInfo struct {
	Name             PortName
	DistName         string
	DistVersion      string
	DistFiles        map[string]*TaggedList
	ExtractSuffix    string
	MasterSites      map[string]*TaggedList
	MasterSiteSubDir string
	SlavePort        string
	MasterPort       string
	Portscout        string
	Maintainer       string
	Comment          string
	GitHub           *GitHubInfo
}

func (p PortName) String() string {
	return p.Category + "/" + p.Name
}

/**
 * Unmarshals a string-encoded list of sites into a map
 * grouped by their tags.
 *
 * Example:
 *    "http://x.com/:tag1,tag2 http://y.com/:tag1 http://z.com/:tag3"
 *
 * Yields:
 *   {
 *     'tag1': [ 'http://x.com', 'http://y.com' ],
 *     'tag2': [ 'http://x.com' ],
 *     'tag3': [ 'http://z.com' ],
 *   }
 */
func UnmarshalTaggedLists(str string) map[string]*TaggedList {
	items := strings.Fields(str)

	listsByTag := make(map[string]*TaggedList)

	for _, item := range items {
		matches := siteGroupSuffix.FindStringSubmatch(item)

		var tags []string
		var url string

		if len(matches) == 2 {
			tags = strings.Split(matches[1], ",")
			url = siteGroupSuffix.ReplaceAllString(item, "")
		} else {
			tags = []string{""}
			url = item
		}

		for _, tag := range tags {
			_, exists := listsByTag[tag]

			// Skip special (e.g. cargo) tags
			if len(tag) > 0 && tag[0] == '_' {
				continue
			}

			if !exists {
				listsByTag[tag] = &TaggedList{
					Items: []string{url},
				}
			} else {
				listsByTag[tag].Items = append(listsByTag[tag].Items, url)
			}
		}
	}

	return listsByTag
}

/**
 * Opposite of the above, except grouped tags (x:tag1,tag2) are
 * broken out into multiple entries (x:tag1 x:tag2).
 */
func MarshalTaggedLists(list map[string]*TaggedList) string {
	arr := make([]string, 0)

	// Tags are sorted for deterministic serialisation
	tags := slices.Sorted(maps.Keys(list))

	for _, tag := range tags {
		for _, item := range list[tag].Items {
			// Skip special (e.g. cargo) tags
			if len(tag) > 0 && tag[0] == '_' {
				continue
			}

			if tag == "" {
				arr = append(arr, item)
			} else {
				arr = append(arr, item+":"+tag)
			}
		}
	}
	return strings.Join(arr, " ")
}
