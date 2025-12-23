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
	queryVars := []string{"PORTVERSION"}

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

		results = append(results, types.PortInfo{
			port,
			lines[0],
		})
	}

	return results
}
