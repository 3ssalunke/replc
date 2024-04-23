package term

import (
	"bytes"
	"fmt"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type TerminalManager struct {
	sessions map[uuid.UUID]string
	mu       sync.RWMutex
}

const (
	SHELL              = "bash"
	TERMINAL_COLS int8 = 100
	WORKSPACE_DIR      = "/workspace"
)

func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		sessions: make(map[uuid.UUID]string),
	}
}

func (tm *TerminalManager) CreateTerminal(id uuid.UUID) (string, error) {
	// cmd := exec.Command(SHELL)
	tm.mu.Lock()
	defer tm.mu.Unlock()

	dirPath, err := filepath.Abs(filepath.Join("..", WORKSPACE_DIR))
	if err != nil {
		return "", err
	}

	tm.sessions[id] = dirPath
	return dirPath + ": ", nil
}

func (tm *TerminalManager) WriteToTerminal(id uuid.UUID, command string) (string, error) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	dirPath, ok := tm.sessions[id]
	if !ok {
		return "", fmt.Errorf("terminal session with ID %s not found", id)
	}
	cmd := exec.Command(SHELL)
	cmd.Dir = dirPath
	cmd.Env = append(cmd.Env, fmt.Sprintf("COLUMNS=%d", TERMINAL_COLS))

	stdin, err := cmd.StdinPipe()
	if err != nil {
		return "", err
	}

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return "", err
	}

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
	if _, err = buf.ReadFrom(stdout); err != nil {
		return "", err
	}

	stdoutArr := strings.Split(buf.String(), "\n")
	updatedDirPath := stdoutArr[len(stdoutArr)-2]
	stdoutArr = stdoutArr[:len(stdoutArr)-2]
	stdoutString := updatedDirPath + ": " + strings.Join(stdoutArr, "\n")

	tm.sessions[id] = updatedDirPath

	defer func() {
		go func() {
			err := cmd.Wait()
			if err != nil {
				if exitErr, ok := err.(*exec.ExitError); ok {
					// Process exited with a non-zero status
					log.Printf("process exited with non-zero status: %v", exitErr)
				} else {
					// Process terminated unexpectedly
					log.Printf("process terminated unexpectedly: %v", err)
				}
			}

			// Check if the process is still running before killing it
			if cmd.ProcessState == nil || !cmd.ProcessState.Exited() {
				// Kill the process
				if err := cmd.Process.Kill(); err != nil {
					log.Printf("error killing process: %v", err)
				}
			} else {
				log.Println("process has already finished")
			}
		}()
	}()

	fmt.Println("Command executed successfully")
	return stdoutString, nil
}

func (tm *TerminalManager) CloseTerminal(id uuid.UUID) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	_, ok := tm.sessions[id]
	if !ok {
		return fmt.Errorf("terminal session with ID %s not found", id)
	}
	delete(tm.sessions, id)
	return nil
}
