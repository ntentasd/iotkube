package ssh

import (
	"bytes"
	"fmt"
	"syscall"

	"golang.org/x/term"
)

func promptSudoPassword() (string, error) {
	fmt.Print("Enter your sudo password: ")
	pw, err := term.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	if err != nil {
		return "", err
	}
	return string(pw), nil
}

func (c *Client) remoteWithSudo(password, cmd string, input ...string) (string, error) {
	session, err := c.client.NewSession()
	if err != nil {
		return "", err
	}
	defer session.Close()

	// Prepare pipes
	stdin, err := session.StdinPipe()
	if err != nil {
		return "", err
	}

	var stdout, stderr bytes.Buffer
	session.Stdout = &stdout
	session.Stderr = &stderr

	go func() {
		fmt.Fprintln(stdin, password)
		for _, s := range input {
			fmt.Fprint(stdin, s)
		}
		stdin.Close()
	}()

	err = session.Run(cmd)
	if err != nil {
		return "", fmt.Errorf("failed to run remote with sudo: %v\nstderr: %s", err, stderr.String())
	}
	return stdout.String(), nil
}
