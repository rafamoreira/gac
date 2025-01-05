package main

import (
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"

	"github.com/BurntSushi/toml"
)

type Config struct {
	CheckRemote bool
	commitMsg   string
	dryRun      bool
}

func main() {
	dryRun, commitMsg := extractArgs()

	// config := Config{
	// 	checkRemote: false,
	// 	commitMsg:   commitMsg,
	// 	dryRun:      *dryRun,
	// }

	config, err := LoadConfig("config.toml")
	if err != nil {
		log.Fatal(err)
	}
	config.commitMsg = commitMsg
	config.dryRun = *dryRun

	err = run(config)
	if err != nil {
		_, _ = fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		os.Exit(1)
	}
}

func extractArgs() (*bool, string) {
	dryRun := flag.Bool("dry-run", false, "Run without committing")
	flag.Parse()
	args := flag.Args()

	var commitMsg string
	if len(args) >= 1 {
		commitMsg = args[0]
	} else {
		commitMsg = ""
	}
	return dryRun, commitMsg
}

func LoadConfig(configPath string) (Config, error) {
	var config Config

	_, err := toml.DecodeFile(configPath, &config)

	return config, err
}

func run(config Config) error {

	// Check if we're in a git repository
	if err := checkGitRepo(); err != nil {
		return fmt.Errorf("not in a git repository: %w", err)
	}

	// Check if this is the first commit
	isFirstCommit, err := isRepositoryEmpty()
	if err != nil {
		return fmt.Errorf("failed to check if repository is empty: %w", err)
	}

	if isFirstCommit {
		return createFirstCommit()
	}

	// Determine if there are any changes
	dirty, err := checkForChanges()
	if err != nil {
		return fmt.Errorf("failed to check for changes: %w", err)
	}
	if !dirty {
		fmt.Println("No changes to commit")
		return nil
	}

	if config.CheckRemote {
		// Try to fetch from remote if there's a remote origin
		origin, err := checkOriginRemote()
		if err != nil {
			return fmt.Errorf("failed to check origin remote: %w", err)
		}
		if origin {
			if err := fetchFromRemote(); err != nil {
				fmt.Println("No remote found or fetch failed, continuing with local operations")
			} else {
				// If fetch succeeded, try to pull
				if err := pullChanges(); err != nil {
					return fmt.Errorf("failed to pull changes: %w", err)
				}
			}
		}
	}

	commitMsg, err := getCommitMessage(config.commitMsg)
	if err != nil {
		return fmt.Errorf("failed to get commit message: %w", err)
	}

	if !config.dryRun {
		// Create the commit
		if err := createCommit(commitMsg); err != nil {
			return fmt.Errorf("failed to create commit: %w", err)
		}
		fmt.Printf("Successfully created commit %s\n", commitMsg)
	} else {
		fmt.Printf("Dry-run: Would create commit with message: %s\n", commitMsg)
	}

	return nil
}

func getCommitMessage(msg string) (string, error) {
	number, err := determineNextCommitNumber()
	if err != nil {
		return "", fmt.Errorf("failed to determine next commit number: %w", err)
	}

	if msg == "" {
		return fmt.Sprintf("%d", number), nil
	} else {
		return fmt.Sprintf("%d: %s", number, msg), nil
	}
}

func checkForChanges() (bool, error) {
	cmd := exec.Command("git", "status", "--porcelain")
	out, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get git status: %w", err)
	}
	outStr := strings.TrimSpace(string(out))
	if outStr == "" {
		return false, nil
	}
	return true, nil
}

func checkGitRepo() error {
	cmd := exec.Command("git", "rev-parse", "--git-dir")
	return cmd.Run()
}

func isRepositoryEmpty() (bool, error) {
	cmd := exec.Command("git", "rev-parse", "HEAD")
	if err := cmd.Run(); err != nil {
		// If HEAD doesn't exist, this is an empty repository
		return true, nil
	}
	return false, nil
}

func createFirstCommit() error {
	// Stage all files
	cmd := exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage files: %w", err)
	}

	// Create first commit
	cmd = exec.Command("git", "commit", "-m", "1")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create first commit: %s", out)
	}
	return nil
}

func checkOriginRemote() (bool, error) {
	cmd := exec.Command("git", "remote")
	remotes, err := cmd.Output()
	if err != nil {
		return false, fmt.Errorf("failed to get remotes: %w", err)
	}
	if len(remotes) == 0 {
		fmt.Println("no remotes found")
		return false, nil
	}

	remoteList := strings.Split(strings.TrimSpace(string(remotes)), "\n")
	for _, remote := range remoteList {
		if remote == "origin" {
			return true, nil
		}
	}
	fmt.Println("no origin remote found")
	return false, nil
}

func fetchFromRemote() error {
	cmd := exec.Command("git", "fetch")
	return cmd.Run()
}

func pullChanges() error {
	cmd := exec.Command("git", "pull")
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("pull failed: %s", out)
	}
	return nil
}

func determineNextCommitNumber() (int, error) {
	// Try to get the last commit message
	lastMsg, err := getLastCommitMessage()
	if err != nil {
		return 0, fmt.Errorf("failed to get last commit message: %w", err)
	}

	// Try to parse it as a number
	if num, err := strconv.Atoi(strings.TrimSpace(lastMsg)); err == nil {
		return num + 1, nil
	}

	// If parsing failed, count all non-merge commits
	return countNonMergeCommits()
}

func getLastCommitMessage() (string, error) {
	cmd := exec.Command("git", "log", "-1", "--pretty=%B")
	out, err := cmd.Output()
	if err != nil {
		return "", err
	}
	return strings.TrimSpace(string(out)), nil
}

func countNonMergeCommits() (int, error) {
	// Get all commits but exclude merges
	cmd := exec.Command("git", "log", "--no-merges", "--oneline")
	out, err := cmd.Output()
	if err != nil {
		return 0, fmt.Errorf("failed to get git log: %w", err)
	}

	// Count lines, each line is one commit
	commits := strings.Split(strings.TrimSpace(string(out)), "\n")
	count := len(commits)
	if commits[0] == "" {
		count = 0
	}
	return count + 1, nil
}

func createCommit(commitMsg string) error {
	// Stage all changes
	cmd := exec.Command("git", "add", ".")
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to stage changes: %w", err)
	}

	// Create commit
	cmd = exec.Command("git", "commit", "-m", commitMsg)
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("failed to create commit: %s", out)
	}
	return nil
}
