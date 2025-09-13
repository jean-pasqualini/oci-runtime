Short answer: runc creates a FIFO named exec.fifo in the container’s state dir (e.g. /run/runc/<id>/exec.fifo) during create, waits on it in start, and the in-container runc init writes a byte to that FIFO to “kick” the exec of the user process.

The exact places in runc’s code
1.	Create the FIFO (during runc create, only for the init process):

// libcontainer/container_linux.go
func (c *linuxContainer) createExecFifo() error {
rootuid, _ := c.Config().HostRootUID()
rootgid, _ := c.Config().HostRootGID()

    fifoName := filepath.Join(c.root, execFifoFilename) // execFifoFilename == "exec.fifo"
    if _, err := os.Stat(fifoName); err == nil {
        return fmt.Errorf("exec fifo %s already exists", fifoName)
    }
    oldMask := unix.Umask(0000)
    if err := unix.Mkfifo(fifoName, 0622); err != nil {
        unix.Umask(oldMask)
        return err
    }
    unix.Umask(oldMask)
    return os.Chown(fifoName, rootuid, rootgid)
}

This is called from Start when process.Init is true.  ￼
2.	runc start waits on the FIFO:

// libcontainer/container_linux.go
func (c *linuxContainer) exec() error {
path := filepath.Join(c.root, execFifoFilename)

    fifoOpen := make(chan struct{})
    select {
    case <-awaitProcessExit(c.initProcess.pid(), fifoOpen):
        return errors.New("container process is already dead")
    case result := <-awaitFifoOpen(path):
        close(fifoOpen)
        f := result.file
        defer f.Close()
        if err := readFromExecFifo(f); err != nil { return err }
        return os.Remove(path) // cleanup exec.fifo
    }
}

awaitFifoOpen opens exec.fifo for reading and blocks until the writer appears.  ￼
3.	runc init (inside the container) writes the “go” byte:

// libcontainer/standard_init_linux.go
// open /proc/self/fd/<fifoFd> (the O_PATH fd to exec.fifo passed in env/fds)
fd, err := unix.Open(fmt.Sprintf("/proc/self/fd/%d", l.fifoFd),
unix.O_WRONLY|unix.O_CLOEXEC, 0)
if err != nil {
return newSystemErrorWithCause(err, "open exec fifo")
}
if _, err := unix.Write(fd, []byte("0")); err != nil {
return newSystemErrorWithCause(err, "write 0 exec fifo")
}
unix.Close(l.fifoFd)

When this byte is written, the start side unblocks and proceeds to exec your container’s entrypoint.  ￼

Pointers / background
•	The FIFO lives in the container state dir (for example /run/runc/<container-id>/exec.fifo). This mechanism replaced older signal-based startup in runc around v1.0.0-rc2.  ￼
•	Nice write-ups that describe this flow and the exec.fifo handshake:  ￼

If you’re re-implementing the same pattern in your own runtime, the key details are:
•	mkfifo(path, 0622) + chown to the container’s root uid/gid on the host.
•	Parent (start) opens FIFO for read and blocks.
•	Child (init) opens the same FIFO for write (via a safe fd path like /proc/self/fd/<fd>) and writes a single byte.
•	Clean up (remove FIFO) once the handshake is done.  ￼