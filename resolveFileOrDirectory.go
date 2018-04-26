package jack

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Resolve(path string, filetype string) (basename string, filenames []string, err error) {
	path, err = filepath.Abs(path)
	if err != nil {
		return
	}

	fileinfo, err := os.Stat(path)
	if err != nil {
		return
	}

	var dirpath string

	if fileinfo.IsDir() {
		// get all files with specified file type
		dirpath = path
		basename = filepath.Base(path)
		dir, err := os.Open(path)
		if err != nil {
			return
		}
		allnames, err := dir.Readdirnames(0)
		if err != nil {
			return
		}
		for _, fname := range allnames {
			if strings.HasSuffix(fname, filetype) {
				filenames = append(filenames, filepath.Join(dirpath, fname))
			}
		}
	} else {
		if !strings.HasSuffix(path, filetype) {
			err = fmt.Errorf("File %s of wrong type, must be of type %s", path, filetype)
			return
		}
		dirpath = filepath.Dir(path)
		base := filepath.Base(path)
		filenames = []string{filepath.Join(dirpath, base)}
		basename = base[0 : len(base)-len(filetype)]
		basename = filepath.Join(dirpath, basename)
	}
}
