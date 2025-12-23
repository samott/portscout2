package types

type PortName struct {
	Category string
	Name     string
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
}

func (p PortName) String() string {
	return p.Category + "/" + p.Name
}
