package main

import (
	"context"
	"fmt"
	"golang.org/x/sys/unix"
	"os"
	"os/exec"
	"syscall"
)

func main() {
	fmt.Println("start container-shim")
	ctx := context.Background()

	if err := do(ctx); err != nil {
		fmt.Println(err)
		os.Exit(2)
	}
}

func do(ctx context.Context) error {
	// prctl(PR_SET_CHILD_SUBREAPER, 1);
	if err := unix.Prctl(unix.PR_SET_CHILD_SUBREAPER, 1, 0, 0, 0); err != nil {
		return err
	}
	cmd := exec.CommandContext(ctx,
		"/tmp/oci-runtime",
		[]string{"run", "--root", "/tmp/state", "--bundle", "/app/bundle", "cid"}...,
	)
	cmd.SysProcAttr = &syscall.SysProcAttr{
		Pdeathsig: syscall.SIGKILL,
		Setpgid:   true,
	}
	// Optional
	cmd.Stdin, cmd.Stdout, cmd.Stderr = os.Stdin, os.Stdout, os.Stderr

	if err := cmd.Run(); err != nil {
		return err
	}

	// Wait init
	var waitInit syscall.WaitStatus
	_, err := syscall.Wait4(-1, &waitInit, 0, nil)
	if err != nil {
		return err
	}

	if waitInit.Exited() {
		os.Exit(waitInit.ExitStatus())
	}
	if waitInit.Signaled() {
		os.Exit(128 + int(waitInit.Signal()))
	}

	return nil
}
