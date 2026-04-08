package tree

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net/url"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
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

func parsePortConfig(portscoutStr string) (types.PortConfig, error) {
	vars := strings.Fields(portscoutStr)

	vmap := make(map[string]string)

	for _, pair := range vars {
		vals := strings.SplitN(pair, "=", 2)

		if len(vals) != 2 {
			return cfg, errors.New("Invalid tuple in PORTSCOUT variable")
		}

		vmap[vals[0]] = vals[1]
	}

	cfg := types.PortConfig{
		IndexSite:    nil,
		LimitVer:     nil,
		LimitEven:    false,
		LimitWhich:   0,
		SkipBeta:     true,
		SkipVersions: make([]string, 0),
		Ignore:       false,
	}

	if val, ok := vmap["site"]; ok {
		if u, err := url.ParseRequestURI(val); err != nil {
			cfg.IndexSite = u
		} else {
			slog.Warn("Invalid site value in PORTSCOUT variable; ignoring", "site", val)
		}
	}

	if val, ok := vmap["limit"]; ok {
		if re, err := regexp.Compile(val); err != nil {
			cfg.LimitVer = re
		} else {
			slog.Warn("Invalid limit value in PORTSCOUT variable; ignoring", "limit", val)
		}
	}

	if val, ok := vmap["limitw"]; ok {
		vals := strings.SplitN(val, ",", 2)

		which, even, err := (func() (int, bool, error) {
			even := true

			if len(vals) != 2 {
				return 0, even, errors.New("Invalid limitw tuple")
			}

			which, err := strconv.Atoi(vals[0])

			if err != nil {
				return which, even, errors.New("Invalid limitw index")
			}

			evenOdd := strings.ToLower(vals[1])

			if evenOdd == "even" {
				even = true
			} else if evenOdd == "odd" {
				even = false
			} else {
				return which, even, errors.New("Invalid limitw parity")
			}

			return which, even, nil
		})()

		if err == nil {
			slog.Warn("Invalid limitw value in PORTSCOUT variable; ignoring", "limitw", val)
		} else {
			cfg.LimitWhich = which
			cfg.LimitEven = even
		}
	}

	if val, ok := vmap["ignore"]; ok {
		if val == "1" || val == "true" || val == "yes" {
			cfg.Ignore = true
		} else {
			cfg.Ignore = false
		}
	}

	if val, ok := vmap["skipb"]; ok {
		if val == "1" || val == "true" || val == "yes" {
			cfg.SkipBeta = true
		} else {
			cfg.SkipBeta = false
		}
	}

	if val, ok := vmap["skipv"]; ok {
		vers := strings.Split(val, ",")

		for _, ver := range vers {
			if trimmed := strings.TrimSpace(ver); trimmed != "" {
				cfg.SkipVersions = append(cfg.SkipVersions, trimmed)
			}
		}
	}

	return cfg, nil
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

func (tree *Tree) QueryPorts(ctx context.Context) {
	var wg sync.WaitGroup

	queryVars := []string{
		"DISTNAME", "DISTVERSION", "DISTFILES", "EXTRACT_SUFX", "MASTER_SITES",
		"MASTER_SITE_SUBDIR", "SLAVE_PORT", "MASTER_PORT", "PORTSCOUT",
		"MAINTAINER", "COMMENT", "USE_GITHUB", "GH_ACCOUNT", "GH_PROJECT",
		"GH_TAGNAME", "GH_SUBDIR",
	}

	for job := range tree.in {
		if ctx.Err() != nil {
			wg.Wait()
			close(tree.out)
			return
		}

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

			ms_subdir := lines[5]
			files := types.UnmarshalTaggedLists(lines[2])
			sites := types.UnmarshalTaggedLists(strings.ReplaceAll(lines[4], "%SUBDIR%", ms_subdir))

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

			_, err = parsePortConfig(lines[8])

			if err != nil {
				tree.out <- QueryResult{
					Info: types.PortInfo{
						Name: port,
					},
					Err: fmt.Errorf("PORTSCOUT value couldn't be parsed: %w", err),
				}
				return
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
