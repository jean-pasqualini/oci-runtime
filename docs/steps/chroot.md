```
	pwd, _ := os.Getwd()
	pwd += "/root"
	
		containerCommand.SysProcAttr = &syscall.SysProcAttr{
		Chroot: pwd,
	}
```