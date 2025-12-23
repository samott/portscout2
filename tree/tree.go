package tree

import (
	"log"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/samott/portscout2/types"
)

func QueryPorts(portsDir string, ports []types.PortName) []types.PortInfo {
	results := make([]types.PortInfo, 0, len(ports))
	queryVars := []string{
		"DISTNAME", "DISTVERSION", "DISTFILES", "EXTRACT_SUFX", "MASTER_SITES",
		"MASTER_SITE_SUBDIR", "SLAVE_PORT", "MASTER_PORT", "PORTSCOUT",
		"MAINTAINER", "COMMENT",
	}

	for _, port := range ports {
		makeFlags := []string{"make", "-C", filepath.Join(portsDir, port.Category, port.Name)}

		flags := make([]string, 0, len(makeFlags)+2*len(queryVars))

		flags = append(flags, makeFlags...)

		for _, v := range queryVars {
			flags = append(flags, "-V")
			flags = append(flags, v)
		}

		cmd := exec.Command("make", flags...)

		output, err := cmd.Output()

		if err != nil {
			log.Fatal("Make call failed: ", err)
		}

		lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")

		files := strings.Split(lines[3], " ")
		sites := strings.Split(lines[4], " ")

		results = append(results, types.PortInfo{
			Name:             port,
			DistName:         lines[0],
			DistVersion:      lines[1],
			DistFiles:        files,
			ExtractSuffix:    lines[3],
			MasterSites:      sites,
			MasterSiteSubDir: lines[5],
			SlavePort:        lines[6],
			MasterPort:       lines[7],
			Portscout:        lines[8],
			Maintainer:       lines[9],
			Comment:          lines[10],
		})
	}

	return results
}
