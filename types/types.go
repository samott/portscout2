package types

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

type PortInfo struct {
	Name             PortName
	DistName         string
	DistVersion      string
	DistFiles        []string
	ExtractSuffix    string
	MasterSites      []string
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
