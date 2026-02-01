package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Config struct {
	GithubRepoUrl  string `json:"github_repo_url"`
	NotesDirectory string `json:"notes_directory"`
}

func main() {
	// 1. Load Configuration
	config, err := loadConfig("config.json")
	if err != nil {
		fmt.Printf("Configuration Error: %v\n", err)
		return
	}

	// 2. Ensure Notes Directory exists
	if _, err := os.Stat(config.NotesDirectory); os.IsNotExist(err) {
		fmt.Printf("Creating directory: %s\n", config.NotesDirectory)
		os.MkdirAll(config.NotesDirectory, 0755)
	}

	// 3. Open or Initialize Repository
	r, err := git.PlainOpen(config.NotesDirectory)
	if err != nil {
		if err == git.ErrRepositoryNotExists {
			fmt.Println("Initializing new Git repository...")
			r, err = git.PlainInit(config.NotesDirectory, false)
			if err != nil {
				fmt.Printf("Init Failed: %v\n", err)
				return
			}
			setupRemote(r, config.GithubRepoUrl)
		} else {
			fmt.Printf("Failed to open repo: %v\n", err)
			return
		}
	}

	// 4. Ensure we are on the Hostname Branch
	err = ensureCorrectBranch(r)
	if err != nil {
		fmt.Printf("Branch Error: %v\n", err)
		return
	}

	// 5. Commit Changes
	err = autoCommit(r)
	if err != nil {
		fmt.Printf("Commit skipped: %v\n", err)
	} else {
		// 6. Push to GitHub
		err = pushToRemote(r)
		if err != nil {
			fmt.Printf("Push Failed: %v\n", err)
		}
	}
}

func loadConfig(path string) (Config, error) {
	var config Config
	file, err := os.ReadFile(path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}

func setupRemote(r *git.Repository, url string) {
	_, err := r.CreateRemote(&git.Config{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		fmt.Printf("Remote setup: %v\n", err)
	}
}

func ensureCorrectBranch(r *git.Repository) error {
	hostname, _ := os.Hostname()
	branchName := plumbing.NewBranchReferenceName(hostname)
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	err = w.Checkout(&git.CheckoutOptions{
		Branch: branchName,
		Create: false,
	})

	if err != nil {
		fmt.Printf("Creating branch for machine: %s\n", hostname)
		err = w.Checkout(&git.CheckoutOptions{
			Branch: branchName,
			Create: true,
		})
	}
	return err
}

func autoCommit(r *git.Repository) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	status, _ := w.Status()
	if status.IsClean() {
		fmt.Println("Working tree clean. Nothing to commit.")
		return nil
	}

	err = w.AddWithOptions(&git.AddOptions{All: true})
	if err != nil {
		return err
	}

	hostname, _ := os.Hostname()
	msg := fmt.Sprintf("Sync: %s [%s]", hostname, time.Now().Format(time.DateTime))

	_, err = w.Commit(msg, &object.Signature{
		Name:  "Gotes Sync",
		Email: "sync@gotes.local",
		When:  time.Now(),
	})
	
	if err == nil {
		fmt.Println("Changes committed locally.")
	}
	return err
}

func pushToRemote(r *git.Repository) error {
	fmt.Println("Syncing to GitHub...")
	
	err := r.Push(&git.PushOptions{
		RemoteName: "origin",
		Progress:   os.Stdout,
	})

	if err == git.NoErrAlreadyUpToDate {
		fmt.Println("GitHub is already up to date.")
		return nil
	}

	if err == nil {
		fmt.Println("Push successful!")
	}
	return err
}
