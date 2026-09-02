package sync

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"os/exec"
	"path/filepath"
)

func cloneOrPull(ctx context.Context, repoURL, localPath string) error {
	if _, err := os.Stat(filepath.Join(localPath, ".git")); err == nil {
		log.Printf("sync: pulling %s", localPath)
		cmd := exec.CommandContext(ctx, "git", "-C", localPath, "pull", "--ff-only")
		cmd.Stdout = io.Discard
		cmd.Stderr = io.Discard
		if runErr := cmd.Run(); runErr != nil {
			return fmt.Errorf("git pull %s: %w", localPath, runErr)
		}
		return nil
	}

	if err := os.MkdirAll(filepath.Dir(localPath), 0o755); err != nil {
		return fmt.Errorf("create parent dir for %s: %w", localPath, err)
	}

	log.Printf("sync: cloning %s into %s", repoURL, localPath)
	cmd := exec.CommandContext(ctx, "git", "clone", repoURL, localPath)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if runErr := cmd.Run(); runErr != nil {
		return fmt.Errorf("git clone %s: %w", repoURL, runErr)
	}
	return nil
}
