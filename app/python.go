package app

import (
	"fmt"
	"io"
	"os/exec"
)

// StartPythonScript starts the Python script for speech recognition.

func StartPythonScript() (*exec.Cmd, io.WriteCloser, io.Reader, io.Reader, error) {
	cmd := exec.Command("python", "C:\\Users\\NotMyPc\\Documents\\Projects\\python\\discordBot-NLP\\voice_recognition.py")

	// Create stdin pipe before starting the process
	stdin, err := cmd.StdinPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get stdin pipe for Python: %v", err)
	}

	// Create stdout and stderr pipes
	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get stdout pipe for Python: %v", err)
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to get stderr pipe for Python: %v", err)
	}

	// Start the Python script
	err = cmd.Start()
	if err != nil {
		return nil, nil, nil, nil, fmt.Errorf("failed to start Python script: %v", err)
	}

	return cmd, stdin, stdout, stderr, nil
}
