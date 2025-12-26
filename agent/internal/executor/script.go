package executor

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"time"
)

type ScriptResult struct {
	ExitCode int    `json:"exit_code"`
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	Error    string `json:"error,omitempty"`
	Duration int64  `json:"duration"`
}

type ScriptExecutor struct {
	coreURL   string
	scriptDir string
	timeout   time.Duration
	client    *http.Client
}

func NewScriptExecutor(coreURL, scriptDir string, timeout time.Duration) *ScriptExecutor {
	if timeout <= 0 {
		timeout = 60 * time.Second
	}

	// Create script directory
	os.MkdirAll(scriptDir, 0755)

	return &ScriptExecutor{
		coreURL:   coreURL,
		scriptDir: scriptDir,
		timeout:   timeout,
		client: &http.Client{
			Timeout: 30 * time.Second,
		},
	}
}

func (e *ScriptExecutor) Execute(ctx context.Context, scriptID string, params map[string]string, expectedChecksum string) (*ScriptResult, error) {
	result := &ScriptResult{}
	start := time.Now()

	// Download script
	scriptPath, err := e.downloadScript(scriptID)
	if err != nil {
		result.Error = fmt.Sprintf("download failed: %v", err)
		result.Duration = time.Since(start).Milliseconds()
		return result, err
	}

	// Verify checksum
	if expectedChecksum != "" {
		actualChecksum, err := e.computeChecksum(scriptPath)
		if err != nil {
			result.Error = fmt.Sprintf("checksum computation failed: %v", err)
			result.Duration = time.Since(start).Milliseconds()
			return result, err
		}

		if actualChecksum != expectedChecksum {
			result.Error = "checksum mismatch"
			result.Duration = time.Since(start).Milliseconds()
			os.Remove(scriptPath)
			return result, fmt.Errorf("checksum mismatch: expected %s, got %s", expectedChecksum, actualChecksum)
		}
	}

	// Make executable
	os.Chmod(scriptPath, 0755)

	// Create context with timeout
	execCtx, cancel := context.WithTimeout(ctx, e.timeout)
	defer cancel()

	// Build command
	var cmd *exec.Cmd
	if runtime.GOOS == "windows" {
		cmd = exec.CommandContext(execCtx, "cmd", "/c", scriptPath)
	} else {
		cmd = exec.CommandContext(execCtx, "/bin/sh", scriptPath)
	}

	// Set environment from params
	cmd.Env = os.Environ()
	for k, v := range params {
		cmd.Env = append(cmd.Env, fmt.Sprintf("%s=%s", k, v))
	}

	// Capture output
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr

	// Run
	err = cmd.Run()
	result.Duration = time.Since(start).Milliseconds()
	result.Stdout = stdout.String()
	result.Stderr = stderr.String()

	if execCtx.Err() == context.DeadlineExceeded {
		result.Error = "execution timeout"
		result.ExitCode = -1
		return result, fmt.Errorf("script execution timeout after %v", e.timeout)
	}

	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			result.ExitCode = exitErr.ExitCode()
		} else {
			result.ExitCode = -1
			result.Error = err.Error()
		}
	}

	// Cleanup
	os.Remove(scriptPath)

	return result, nil
}

func (e *ScriptExecutor) downloadScript(scriptID string) (string, error) {
	url := fmt.Sprintf("%s/api/scripts/%s/content", e.coreURL, scriptID)

	resp, err := e.client.Get(url)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("download failed with status %d", resp.StatusCode)
	}

	// Save to temp file
	scriptPath := filepath.Join(e.scriptDir, fmt.Sprintf("script_%s_%d", scriptID, time.Now().UnixNano()))

	file, err := os.Create(scriptPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = io.Copy(file, resp.Body)
	if err != nil {
		os.Remove(scriptPath)
		return "", err
	}

	return scriptPath, nil
}

func (e *ScriptExecutor) computeChecksum(path string) (string, error) {
	file, err := os.Open(path)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := sha256.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return hex.EncodeToString(hash.Sum(nil)), nil
}

func (e *ScriptExecutor) SetTimeout(timeout time.Duration) {
	e.timeout = timeout
}
