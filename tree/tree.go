package tree

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/samott/portscout2/types"
)

type Tree struct {
	makeCmd  string
	portsDir string
	sem      chan struct{}
	maxProc  int
}

func NewTree(makeCmd string, portsDir string, maxProc int) *Tree {
	return &Tree{
		makeCmd:  makeCmd,
		portsDir: portsDir,
		maxProc:  maxProc,
		sem:      make(chan struct{}, maxProc),
	}
}

func (tree *Tree) QueryPorts(ports []types.PortName, callback func(types.PortInfo)) (int32, error) {
	var wg sync.WaitGroup

	var firstErr error = nil
	var errGuard sync.Once

	queryVars := []string{
		"DISTNAME", "DISTVERSION", "DISTFILES", "EXTRACT_SUFX", "MASTER_SITES",
		"MASTER_SITE_SUBDIR", "SLAVE_PORT", "MASTER_PORT", "PORTSCOUT",
		"MAINTAINER", "COMMENT",
	}

	var completedCount int32 = 0

	for _, port := range ports {
		if firstErr != nil {
			break
		}

		wg.Add(1)

		makeFlags := []string{"-C", filepath.Join(tree.portsDir, port.Category, port.Name)}

		flags := make([]string, 0, len(makeFlags)+2*len(queryVars))

		flags = append(flags, makeFlags...)

		for _, v := range queryVars {
			flags = append(flags, "-V")
			flags = append(flags, v)
		}

		go func() {
			defer wg.Done()

			tree.sem <- struct{}{}

			defer func() {
				<-tree.sem
			}()

			cmd := exec.Command(tree.makeCmd, flags...)

			var stderr bytes.Buffer

			cmd.Stderr = &stderr

			output, err := cmd.Output()

			if err != nil {
				errGuard.Do(func() { firstErr = fmt.Errorf("Make call failed: %q: %w", stderr.String(), err) })
				return
			}

			lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")

			files := strings.Split(lines[3], " ")
			sites := strings.Split(lines[4], " ")

			callback(types.PortInfo{
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

			atomic.AddInt32(&completedCount, 1)
		}()
	}

	wg.Wait()

	return completedCount, firstErr
}
