package tree

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/samott/portscout2/types"
)

type Tree struct {
	makeCmd  string
	portsDir string
	sem      chan struct{}
	maxProc  int
	in       chan QueryJob
	out      chan QueryResult
}

type QueryJob struct {
	Port types.PortName
}

type QueryResult struct {
	Info types.PortInfo
	Err  error
}

func NewTree(makeCmd string, portsDir string, maxProc int) *Tree {
	return &Tree{
		makeCmd:  makeCmd,
		portsDir: portsDir,
		maxProc:  maxProc,
		sem:      make(chan struct{}, maxProc),
		in:       make(chan QueryJob, maxProc),
		out:      make(chan QueryResult, maxProc),
	}
}

func (c *Tree) In() chan<- QueryJob {
	return c.in
}

func (c *Tree) Out() <-chan QueryResult {
	return c.out
}

func (tree *Tree) QueryPorts() {
	var wg sync.WaitGroup

	queryVars := []string{
		"DISTNAME", "DISTVERSION", "DISTFILES", "EXTRACT_SUFX", "MASTER_SITES",
		"MASTER_SITE_SUBDIR", "SLAVE_PORT", "MASTER_PORT", "PORTSCOUT",
		"MAINTAINER", "COMMENT", "USE_GITHUB", "GH_ACCOUNT", "GH_PROJECT",
		"GH_TAGNAME", "GH_SUBDIR",
	}

	for job := range tree.in {
		port := job.Port

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
				tree.out <- QueryResult{
					Info: types.PortInfo{
						Name: port,
					},
					Err: fmt.Errorf("Make call failed: %q: %w", stderr.String(), err),
				}
				return
			}

			lines := strings.Split(strings.TrimSuffix(string(output), "\n"), "\n")

			files := strings.Split(lines[3], " ")
			sites := strings.Split(lines[4], " ")

			var github *types.GitHubInfo

			if lines[11] != "" {
				github = &types.GitHubInfo{
					Account: lines[12],
					Project: lines[13],
					TagName: lines[14],
					SubDir:  lines[15],
				}
			} else {
				github = nil
			}

			tree.out <- QueryResult{
				Info: types.PortInfo{
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
					GitHub:           github,
				},
				Err: nil,
			}
		}()
	}

	wg.Wait()
	close(tree.out)
}
