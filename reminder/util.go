package reminder

import "os"

func IsExist(path string) bool {
	_, err := os.Stat(path)
	return err == nil
}

func IsDir(path string) bool {
	if fileinfo, err := os.Stat(path); err != nil {
		return false
	} else {
		return fileinfo.Mode().IsDir()
	}
}

func IsFile(path string) bool {
	if fileinfo, err := os.Stat(path); err != nil {
		return false
	} else {
		return fileinfo.Mode().IsRegular()
	}
}
