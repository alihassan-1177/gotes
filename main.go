package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"

	"github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/config"
	"github.com/go-git/go-git/v5/plumbing"
	"github.com/go-git/go-git/v5/plumbing/object"
)

type Config struct {
	GithubRepoUrl  string `json:"github_repo_url"`
	NotesDirectory string `json:"notes_directory"`
	BranchName string `json:"branch_name"`
}

func main() {
	config, err := loadConfig("gotes-config.json")
	if err != nil {
		fmt.Printf("Configuration Error: %v\n", err)
		return
	}

	if _, err := os.Stat(config.NotesDirectory); os.IsNotExist(err) {
		fmt.Printf("Creating directory: %s\n", config.NotesDirectory)
		os.MkdirAll(config.NotesDirectory, 0755)
	}

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

	pullLatest(r)
	if err != nil {
		fmt.Printf("Pull Warning: %v (Proceeding anyway...)\n", err)
	}
	
	err = autoCommit(r)
	if err != nil {
		fmt.Printf("Commit skipped: %v\n", err)
	} else {
		err = pushToRemote(r)
		if err != nil {
			fmt.Printf("Push Failed: %v\n", err)
		}
	}

	err = ensureCorrectBranch(r, config.BranchName)
	if err != nil {
		fmt.Printf("Branch Error: %v\n", err)
		return
	}

}

func loadConfig(path string) (Config, error) {
	var config Config
	home, _ := os.UserHomeDir()
	var full_path = filepath.Join(home, path)
	
	file, err := os.ReadFile(full_path)
	if err != nil {
		return config, err
	}
	err = json.Unmarshal(file, &config)
	return config, err
}

func setupRemote(r *git.Repository, url string) {

	_, err := r.CreateRemote(&config.RemoteConfig{
		Name: "origin",
		URLs: []string{url},
	})
	if err != nil {
		fmt.Printf("Remote setup: %v\n", err)
	}
}

func ensureCorrectBranch(r *git.Repository, hostname string) error {
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

	_, err = w.Commit(msg, &git.CommitOptions{
		Author: &object.Signature{
			Name:  "Gotes Sync",
			Email: "sync@gotes.local",
			When:  time.Now(),
		},
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
		Force: true,
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

func pullLatest(r *git.Repository) error {
	w, err := r.Worktree()
	if err != nil {
		return err
	}

	fmt.Println("Pulling latest changes from remote...")
	err = w.Pull(&git.PullOptions{
		RemoteName: "origin",
	})

	if err == git.NoErrAlreadyUpToDate {
		fmt.Println("Local is already up to date with remote.")
		return nil
	}

	if err != nil && err.Error() != "remote repository is empty" {
		return err
	}

	return nil
}
