package repo

import (
	"fmt"
	"strings"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/plumbing/filemode"
	"github.com/go-git/go-git/v6/utils/merkletrie"

	"github.com/samott/portscout2/types"
)

type PortChange int

const (
	PortAdded = iota
	PortRemoved
	PortChanged
)

func getPortName(path string) (*types.PortName, bool) {
	frags := strings.Split(path, "/")

	if len(frags) < 2 || (frags[0][0] >= 'A' && frags[0][0] <= 'Z') {
		// Not under a port directory
		return nil, false
	}

	if len(frags[0]) == 0 || len(frags[1]) == 0 {
		// Empty name or category
		return nil, false
	}

	// category/port or category/port/ (unsure if the second
	// case actually happens...)
	isRoot := (len(frags) == 2) || (len(frags) == 3 && len(frags[2]) == 0)

	portName := types.PortName{frags[0], frags[1]}

	return &portName, isRoot
}

func FindUpdated(portsDir string, lastCommitHashStr string) (string, map[types.PortName]PortChange, error) {
	portsTree, err := git.PlainOpen(portsDir)

	ports := make(map[types.PortName]PortChange)

	if err != nil {
		return "", nil, fmt.Errorf("Unable to open ports tree: %w", err)
	}

	head, err := portsTree.Head()

	if err != nil {
		return "", nil, fmt.Errorf("Unable to get HEAD: %w", err)
	}

	commit, err := portsTree.CommitObject(head.Hash())

	if err != nil {
		return "", nil, fmt.Errorf("Unable to find commit: %w", err)
	}

	tree, err := commit.Tree()

	if err != nil {
		return "", nil, fmt.Errorf("Error getting tree: %w", err)
	}

	fmt.Println("Tree hash:", tree.Hash)

	lastCommitHash := plumbing.NewHash(lastCommitHashStr)

	lastCommit, err := portsTree.CommitObject(lastCommitHash)

	if err != nil {
		return "", nil, fmt.Errorf("Unable to find commit: %w", err)
	}

	lastTree, err := lastCommit.Tree()

	if err != nil {
		return "", nil, fmt.Errorf("Error getting tree: %w", err)
	}

	changes, err := lastTree.Diff(tree)

	if err != nil {
		return "", nil, fmt.Errorf("Tree diff failed: %w", err)
	}

	for _, change := range changes {
		action, _ := change.Action()

		patch, _ := change.Patch()
		from, to := patch.FilePatches()[0].Files()

		if action == merkletrie.Insert {
			portName, isRoot := getPortName(to.Path())

			if portName == nil {
				continue
			}

			if isRoot {
				if to.Mode() != filemode.Dir {
					continue
				}

				ports[*portName] = PortAdded
			} else {
				_, exists := ports[*portName]

				if !exists {
					// Don't mark a newly-added port as
					// changed (e.g. if we see inserts for
					// files in the port directory)
					ports[*portName] = PortChanged
				}
			}

			continue
		}

		if action == merkletrie.Delete {
			portName, isRoot := getPortName(from.Path())

			if portName == nil {
				continue
			}

			if isRoot {
				if from.Mode() != filemode.Dir {
					continue
				}

				ports[*portName] = PortRemoved
			} else {
				_, exists := ports[*portName]

				if !exists {
					// Don't mark a newly-added port as
					// changed (e.g. if we see inserts for
					// files in the port directory)
					ports[*portName] = PortChanged
				}
			}

			continue
		}

		if action == merkletrie.Modify {
			portName, isRoot := getPortName(to.Path())

			if portName == nil {
				continue
			}

			if isRoot && to.Mode() != filemode.Dir {
				continue
			}

			_, exists := ports[*portName]

			if !exists {
				// Don't mark a newly-added port as
				// changed (e.g. if we see inserts for
				// files in the port directory)
				ports[*portName] = PortChanged
			}

			continue
		}
	}

	return tree.Hash.String(), ports, nil
}

func FindAllPorts(portsDir string) (string, map[types.PortName]PortChange, error) {
	portsTree, err := git.PlainOpen(portsDir)

	ports := make(map[types.PortName]PortChange)

	if err != nil {
		return "", nil, fmt.Errorf("Unable to open ports tree: %w", err)
	}

	head, err := portsTree.Head()

	if err != nil {
		return "", nil, fmt.Errorf("Unable to get HEAD: %w", err)
	}

	commit, err := portsTree.CommitObject(head.Hash())

	if err != nil {
		return "", nil, fmt.Errorf("Unable to find commit: %w", err)
	}

	tree, err := commit.Tree()

	if err != nil {
		return "", nil, fmt.Errorf("Error getting tree: %w", err)
	}

	for _, entry := range tree.Entries {
		if entry.Mode != filemode.Dir {
			continue
		}

		if entry.Name[0] >= 'A' && entry.Name[0] <= 'Z' {
			// Not a category dir (rather a ports system dir)
			continue
		}

		if entry.Name[0] == '.' {
			// Not a category dir
			continue
		}

		category := entry.Name

		subTree, err := tree.Tree(category)

		if err != nil {
			return "", nil, fmt.Errorf("Unable to get subtree: %w", err)
		}

		for _, subDir := range subTree.Entries {
			if subDir.Mode != filemode.Dir {
				continue
			}

			if subDir.Name[0] == '.' {
				// Not a port dir
				continue
			}

			port := subDir.Name

			portName := types.PortName{
				Category: category,
				Name:     port,
			}

			ports[portName] = PortAdded
		}
	}

	return commit.Hash.String(), ports, nil
}
