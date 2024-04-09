package fs

import (
	"os"
	"path/filepath"
)

type File struct {
	Type string `json:"type"`
	Name string `json:"name"`
	Path string `json:"path"`
}

func FetchDir(dir, baseDir string) ([]File, error) {
	files, err := os.ReadDir(dir)
	if err != nil {
		return nil, err
	}

	var fileList []File
	for _, file := range files {
		fileType := "file"
		if file.IsDir() {
			fileType = "dir"
		}
		fileList = append(fileList, File{
			Type: fileType,
			Name: file.Name(),
			Path: filepath.Join(baseDir, file.Name()),
		})
	}

	return fileList, nil
}
