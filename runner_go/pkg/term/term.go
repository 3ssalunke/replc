package term

import (
	"bytes"
	"fmt"
	"io"
	"log"
	"os/exec"
	"path/filepath"
	"strings"
	"sync"

	"github.com/google/uuid"
)

type CmdContainer struct {
	Cmd    *exec.Cmd
	StdIn  io.WriteCloser
	StdOut io.ReadCloser
}

type TerminalManager struct {
	sessions map[uuid.UUID]CmdContainer
	mu       sync.RWMutex
}

const (
	SHELL              = "bash"
	TERMINAL_COLS int8 = 100
	WORKSPACE_DIR      = "/workspace"
)

func NewTerminalManager() *TerminalManager {
	return &TerminalManager{
		sessions: make(map[uuid.UUID]CmdContainer),
	}
}

func (tm *TerminalManager) CreateTerminal(id uuid.UUID) (string, error) {
	cmd := exec.Command(SHELL)

	dirPath, err := filepath.Abs(filepath.Join("..", WORKSPACE_DIR))
	if err != nil {
		return "", err
	}

	cmd.Dir = dirPath
	cmd.Env = append(cmd.Env, fmt.Sprintf("COLUMNS=%d", TERMINAL_COLS))

	tm.mu.Lock()
	defer tm.mu.Unlock()

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

	tm.sessions[id] = CmdContainer{
		Cmd:    cmd,
		StdIn:  stdin,
		StdOut: stdout,
	}

	return cmd.Dir + ": ", nil
}

func (tm *TerminalManager) WriteToTerminal(id uuid.UUID, command string) (string, error) {
	tm.mu.RLock()
	defer tm.mu.RUnlock()

	cmd, ok := tm.sessions[id]
	if !ok {
		return "", fmt.Errorf("terminal session with ID %s not found", id)
	}

	// Write a command to the standard input
	command = command + "\npwd\n"
	if _, err := cmd.StdIn.Write([]byte(command)); err != nil {
		log.Println("in", err)
		return "", err
	}

	// Close the standard input to signal the end of input
	if err := cmd.StdIn.Close(); err != nil {
		return "", err
	}

	var buf bytes.Buffer
	_, err := buf.ReadFrom(cmd.StdOut)
	if err != nil {
		return "", err
	}
	stdoutArr := strings.Split(buf.String(), "\n")
	cmd.Cmd.Dir = stdoutArr[len(stdoutArr)-2]
	stdoutArr = stdoutArr[:len(stdoutArr)-2]
	stdoutString := cmd.Cmd.Dir + ": " + strings.Join(stdoutArr, "\n")

	// Wait for the command to finish executing
	if err := cmd.Cmd.Wait(); err != nil {
		return "", err
	}
	fmt.Println("Command executed successfully")
	return stdoutString, nil
}

func (tm *TerminalManager) CloseTerminal(id uuid.UUID) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	cmd, ok := tm.sessions[id]
	if !ok {
		return fmt.Errorf("terminal session with ID %s not found", id)
	}
	err := cmd.Cmd.Process.Kill()
	if err != nil {
		return err
	}
	delete(tm.sessions, id)
	return nil
}
