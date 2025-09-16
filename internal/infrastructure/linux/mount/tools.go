package mount

import "os"

func isDir(path string) (bool, error) {
	stat, err := os.Stat(path)
	if err != nil {
		return false, err
	}
	return stat.IsDir(), nil
}

func touch(file string) error {
	f, err := os.Create(file)
	if err != nil && !os.IsExist(err) {
		return err
	}
	f.Close()
	return nil
}
