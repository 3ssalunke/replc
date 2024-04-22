package term

import (
	"bytes"
	"fmt"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type TerminalManager struct {
	sessions map[uuid.UUID]*exec.Cmd
	mu       sync.Mutex
}

const (
	SHELL              = "bash"
	TERMINAL_COLS int8 = 100
	WORKSPACE_DIR      = "/workspace"
)

func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		sessions: make(map[uuid.UUID]*exec.Cmd),
	}
}

func (tm *TerminalManager) CreateTerminal(id uuid.UUID) error {
	cmd := exec.Command(SHELL)

	dirPath, err := filepath.Abs(filepath.Join("..", WORKSPACE_DIR))
	if err != nil {
		return err
	}

	cmd.Dir = dirPath
	cmd.Env = append(cmd.Env, fmt.Sprintf("COLUMNS=%d", TERMINAL_COLS))

	tm.mu.Lock()
	tm.sessions[id] = cmd
	tm.mu.Unlock()

	return nil
}

func (tm *TerminalManager) WriteToTerminal(id uuid.UUID, command string) (string, error) {
	cmd, ok := tm.sessions[id]
	if !ok {
		return "", fmt.Errorf("terminal session with ID %s not found", id)
	}

	// Obtain a pipe to the standard input of the command
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}
	defer stdout.Close()

	// Obtain a pipe to the standard input of the command
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}
	defer stdin.Close()

	// Start the command
	if err := cmd.Start(); err != nil {
		return "", err
	}

	// Write a command to the standard input
	command = command + "\npwd\n"
	if _, err := stdin.Write([]byte(command)); err != nil {
		return "", err
	}

	// Close the standard input to signal the end of input
	if err := stdin.Close(); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, err = buf.ReadFrom(stdout)
	if err != nil {
		return "", err
	}
	stdoutArr := strings.Split(buf.String(), "\n")
	cmd.Dir = stdoutArr[len(stdoutArr)-2]
	stdoutArr = stdoutArr[:len(stdoutArr)-2]

	modifiedStdoutArr := make([]string, len(stdoutArr))
	for i, element := range modifiedStdoutArr {
		modifiedStdoutArr[i] = cmd.Dir + ": " + element
	}

	// Wait for the command to finish executing
	if err := cmd.Wait(); err != nil {
		return "", err
	}
	fmt.Println("Command executed successfully")

	stdoutString := strings.Join(modifiedStdoutArr, "\n")
	return stdoutString, nil
}

func (tm *TerminalManager) CloseTerminal(id uuid.UUID) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	cmd, ok := tm.sessions[id]
	if !ok {
		return fmt.Errorf("terminal session with ID %s not found", id)
	}
	err := cmd.Process.Kill()
	if err != nil {
		return err
	}
	delete(tm.sessions, id)
	return nil
}
