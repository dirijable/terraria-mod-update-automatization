package git

import (
	"fmt"
	"os/exec"
)

func SyncRepo(repoPath string) error {
	if err := runGitCmd(repoPath, "fetch", "origin"); err != nil {
		return fmt.Errorf("git fetch failed: %w", err)
	}
	if err := runGitCmd(repoPath, "reset", "--hard", "origin/main"); err != nil {
		return fmt.Errorf("git reset failed: %w", err)
	}
	return nil
}

func PushUpdates(repoPath, commitMsg string) error {
	if err := runGitCmd(repoPath, "add", "."); err != nil {
		return fmt.Errorf("git add failed: %w", err)
	}

	if err := runGitCmd(repoPath, "commit", "-m", commitMsg); err != nil {
		return fmt.Errorf("git commit failed: %w", err)
	}

	if err := runGitCmd(repoPath, "push", "origin", "main"); err != nil {
		return fmt.Errorf("git push failed: %w", err)
	}

	return nil
}

func ResetHard(repoPath string) error {
	if err := runGitCmd(repoPath, "reset", "--hard", "origin/main"); err != nil {
		return fmt.Errorf("git reset: %w", err)
	}

	if err := runGitCmd(repoPath, "clean", "-fd"); err != nil {
		return fmt.Errorf("git clean: %w", err)
	}
	
	return nil
}

func runGitCmd(dir string, args ...string) error {
	cmd := exec.Command("git", args...)
	cmd.Dir = dir
	output, err := cmd.CombinedOutput()
	if err != nil {
		return fmt.Errorf("%w: %s", err, string(output))
	}
	return nil
}
