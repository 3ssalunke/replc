package term

import (
	"os/exec"
	"path/filepath"
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

func (tm *TerminalManager) CreateTerminal(id uuid.UUID) (*exec.Cmd, error) {
	cmd := exec.Command(SHELL)

	dirPath, err := filepath.Abs(filepath.Join("..", WORKSPACE_DIR))
	if err != nil {
		return nil, err
	}

	cmd.Dir = dirPath
	cmd.Env = append(cmd.Env, "COLUMNS="+string(TERMINAL_COLS))

	if err := cmd.Start(); err != nil {
		return nil, err
	}

	tm.mu.Lock()
	defer tm.mu.Unlock()
	tm.sessions[id] = cmd

	go func() {
		cmd.Wait()
		tm.mu.Lock()
		defer tm.mu.Unlock()
		delete(tm.sessions, id)
	}()

	return cmd, nil
}

func (tm *TerminalManager) WriteToTerminal(id uuid.UUID, data []byte) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if cmd, ok := tm.sessions[id]; ok {
		stdin, err := cmd.StdinPipe()
		if err != nil {
			return err
		}
		defer stdin.Close()

		_, err = stdin.Write(data)
		return err
	}

	return nil
}

func (tm *TerminalManager) CloseTerminal(id uuid.UUID) error {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	if cmd, ok := tm.sessions[id]; ok {
		err := cmd.Process.Kill()
		delete(tm.sessions, id)
		return err
	}

	return nil
}
