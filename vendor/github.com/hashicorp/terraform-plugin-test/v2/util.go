package tftest

import (
	"os"
	"path/filepath"
)

func symlinkFile(src string, dest string) (err error) {
	err = os.Symlink(src, dest)
	if err == nil {
		srcInfo, err := os.Stat(src)
		if err != nil {
			err = os.Chmod(dest, srcInfo.Mode())
		}
	}

	return
}

// symlinkDir is a simplistic function for recursively symlinking all files in a directory to a new path.
// It is intended only for limited internal use and does not cover all edge cases.
func symlinkDir(srcDir string, destDir string) (err error) {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(destDir, srcInfo.Mode())
	if err != nil {
		return err
	}

	directory, _ := os.Open(srcDir)
	defer directory.Close()
	objects, err := directory.Readdir(-1)

	for _, obj := range objects {
		srcPath := filepath.Join(srcDir, obj.Name())
		destPath := filepath.Join(destDir, obj.Name())

		if obj.IsDir() {
			err = symlinkDir(srcPath, destPath)
			if err != nil {
				return err
			}
		} else {
			err = symlinkFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}

	}
	return
}

// symlinkDirectoriesOnly finds only the first-level child directories in srcDir
// and symlinks them into destDir.
// Unlike symlinkDir, this is done non-recursively in order to limit the number
// of file descriptors used.
func symlinkDirectoriesOnly(srcDir string, destDir string) (err error) {
	srcInfo, err := os.Stat(srcDir)
	if err != nil {
		return err
	}

	err = os.MkdirAll(destDir, srcInfo.Mode())
	if err != nil {
		return err
	}

	directory, err := os.Open(srcDir)
	if err != nil {
		return err
	}
	defer directory.Close()
	objects, err := directory.Readdir(-1)
	if err != nil {
		return err
	}

	for _, obj := range objects {
		srcPath := filepath.Join(srcDir, obj.Name())
		destPath := filepath.Join(destDir, obj.Name())

		if obj.IsDir() {
			err = symlinkFile(srcPath, destPath)
			if err != nil {
				return err
			}
		}

	}
	return
}
