package repo

import (
	"fmt"
	"log"
	"log/slog"
	"strings"

	git "github.com/go-git/go-git/v6"
	"github.com/go-git/go-git/v6/plumbing"
	"github.com/go-git/go-git/v6/utils/merkletrie"
)

type PortName struct {
	category string
	name     string
}

func getPortName(path string) (*PortName, bool) {
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

	portName := PortName{frags[0], frags[1]}

	return &portName, isRoot
}

func FindUpdated(portsDir string, lastCommitHashStr string) []PortName {
	portsTree, err := git.PlainOpen(portsDir)

	addedPorts := make(map[PortName]bool)
	deletedPorts := make(map[PortName]bool)
	changedPorts := make(map[PortName]bool)

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
			portName, isRoot := getPortName(to.Path())

			if portName == nil {
				continue
			}

			if isRoot {
				_, exists := changedPorts[*portName]

				if exists {
					// Ignore changes to an added port in
					// the same diff.
					delete(changedPorts, *portName)
				}

				addedPorts[*portName] = true
			} else {
				_, exists := addedPorts[*portName]

				if !exists {
					// Don't mark a newly-added port as
					// changed (e.g. if we see inserts for
					// files in the port directory)
					changedPorts[*portName] = true
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
				_, exists := changedPorts[*portName]

				if exists {
					// Ignore changes to a deleted port in
					// the same diff.
					delete(changedPorts, *portName)
				}

				deletedPorts[*portName] = true
			} else {
				_, exists := addedPorts[*portName]

				if !exists {
					// Don't mark a newly-added port as
					// changed (e.g. if we see inserts for
					// files in the port directory)
					changedPorts[*portName] = true
				}
			}

			continue
		}

		if action == merkletrie.Modify {
			portName, _ := getPortName(to.Path())

			if portName == nil {
				continue
			}

			_, existsAdded := addedPorts[*portName]
			_, existsDeleted := deletedPorts[*portName]

			if existsAdded || existsDeleted {
				// Ignore changes if we are adding or
				// deleting a port
				continue
			}

			changedPorts[*portName] = true

			continue
		}
	}

	slog.Info("Changed", changedPorts)

	return []PortName{}
}
