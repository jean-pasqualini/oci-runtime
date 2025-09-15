package mount

import "os"

func Touch(file string) error {
	f, err := os.Create(file)
	if err != nil && !os.IsExist(err) {
		return err
	}
	f.Close()
	return nil
}
