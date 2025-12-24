package repo

import (
	"fmt"
	"log"
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

func FindUpdated(portsDir string, lastCommitHashStr string) map[types.PortName]PortChange {
	portsTree, err := git.PlainOpen(portsDir)

	ports := make(map[types.PortName]PortChange)

	if err != nil {
		log.Fatal("Unable to open ports tree: ", err)
	}

	head, err := portsTree.Head()

	if err != nil {
		log.Fatal("Unable to get HEAD: ", err)
	}

	commit, err := portsTree.CommitObject(head.Hash())

	if err != nil {
		log.Fatal("Unable to find commit: ", err)
	}

	tree, err := commit.Tree()

	if err != nil {
		log.Fatal("Error getting tree: ", err)
	}

	fmt.Println("Tree hash:", tree.Hash)

	lastCommitHash := plumbing.NewHash(lastCommitHashStr)

	lastCommit, err := portsTree.CommitObject(lastCommitHash)

	if err != nil {
		log.Fatal("Unable to find commit: ", err)
	}

	lastTree, err := lastCommit.Tree()

	if err != nil {
		log.Fatal("Error getting tree: ", err)
	}

	changes, err := lastTree.Diff(tree)

	if err != nil {
		log.Fatal("Tree diff failed: ", err)
	}

	for _, change := range changes {
		action, _ := change.Action()

		patch, _ := change.Patch()
		from, to := patch.FilePatches()[0].Files()

		if action == merkletrie.Insert {
			if to.Mode() != filemode.Dir {
				continue
			}

			portName, isRoot := getPortName(to.Path())

			if portName == nil {
				continue
			}

			if isRoot {
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
			if from.Mode() != filemode.Dir {
				continue
			}

			portName, isRoot := getPortName(from.Path())

			if portName == nil {
				continue
			}

			if isRoot {
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
			if to.Mode() != filemode.Dir {
				continue
			}

			portName, _ := getPortName(to.Path())

			if portName == nil {
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

	portNames := make([]types.PortName, 0, len(ports))

	for portName := range ports {
		portNames = append(portNames, portName)
	}

	return ports
}
