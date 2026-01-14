package helpers

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
)

type File struct {
	Name string `json:"name"`
	Size string `json:"size"`
	Path string `json:"-"`
}

func GetFileList(fileDirectory string) []File {
	entries, err := os.ReadDir(fileDirectory)
	out := []File{}
	if err != nil {
		log.Println(err)
		return out
	}
	for _, folder := range entries {
		if folder.Name() == "compiled" {
			continue
		}
		volumes, err := os.ReadDir(fileDirectory + folder.Name())
		if err != nil {
			log.Println(err)
			return out
		}

		for _, volume := range volumes {

			zippedFiles, err := os.ReadDir(fileDirectory + folder.Name() + "/" + volume.Name())
			if err != nil {
				log.Println(err)
				return out
			}

			for _, zippedFile := range zippedFiles {
				info, err := zippedFile.Info()
				if err != nil {
					log.Println(err)
					continue
				}
				out = append(out, File{
					Path: fileDirectory + folder.Name() + "/" + volume.Name() + "/" + info.Name(),
					Name: info.Name(),
					Size: parseFileSize(uint64(info.Size()))})
			}
		}
	}
	return out
}

func RenameAllCBFiles(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, file := range entries {
		if file.IsDir() {
			RenameAllCBFiles(filepath.Join(path, file.Name()))
		} else {
			ext := filepath.Ext(file.Name())
			if ext == ".cbr" || ext == ".cbz" {
				fileName := file.Name()[:len(file.Name())-len(ext)]
				var newExt string
				if ext == ".cbr" {
					newExt = ".rar"
				} else {
					newExt = ".zip"
				}

				os.Rename(filepath.Join(path, file.Name()), filepath.Join(path, fileName)+newExt)
			}

		}
	}

	return true
}

type RecursiveProgress struct {
	current int
	total   int
}

func ProcessZipsAndRars(path string, job *Job) bool {
	fileCount := GetFileProcessCount(path)
	recursiveProgress := RecursiveProgress{0, fileCount}
	return ProcessZipsAndRarsImpl(path, job, &recursiveProgress)
}

func GetFileProcessCount(path string) int {
	var count int
	entries, err := os.ReadDir(path)
	if err != nil {
		return 0
	}

	for _, file := range entries {
		if file.IsDir() {
			count += GetFileProcessCount(filepath.Join(path, file.Name()))
		} else {
			ext := filepath.Ext(file.Name())
			if ext == ".rar" || ext == ".zip" {
				count++
			}
		}
	}

	return count
}
func ProcessZipsAndRarsImpl(path string, job *Job, recursiveProgress *RecursiveProgress) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}
	for _, file := range entries {
		if file.IsDir() {
			ProcessZipsAndRarsImpl(filepath.Join(path, file.Name()), job, recursiveProgress)
		} else {
			ext := filepath.Ext(file.Name())
			if ext == ".rar" || ext == ".zip" {
				var res bool
				filePath := filepath.Join(path, file.Name())
				if ext == ".rar" {
					res = UnRarFiles(filePath, path)
				} else {
					res = UnzipFilesImpl(filePath, map[string]bool{}, path)
				}

				if res {
					err := os.Remove(filePath)
					if err != nil {
						return false
					}
				}

				recursiveProgress.current++
			}

		}
		job.Progress.Completion = float32(recursiveProgress.current) / float32(recursiveProgress.total)
	}

	return true
}

func ZipImages(path string, token string, job *Job) string {
	files := getAllFilesInDir(path)
	if len(files) == 0 {
		return ""
	}

	outputPath := filepath.Join(path, token+".zip")
	if ZipFiles(outputPath, files, job) {
		return outputPath
	}
	return ""
}

func getAllFilesInDir(path string) []string {
	var files []string
	entries, err := os.ReadDir(path)
	if err != nil {
		return files
	}

	for _, file := range entries {
		if file.IsDir() {
			files = append(files, getAllFilesInDir(filepath.Join(path, file.Name()))...)
		} else {
			files = append(files, filepath.Join(path, file.Name()))
		}
	}

	return files
}

func CleanFolder(path string) bool {
	entries, err := os.ReadDir(path)
	if err != nil {
		return false
	}

	for _, file := range entries {
		if file.IsDir() {
			os.RemoveAll(filepath.Join(path, file.Name()))
		}
	}

	return true
}

func RenameOutput(path string) string {
	base := filepath.Base(path)
	ext := filepath.Ext(base)

	outputPath := "/comics/compiled/" + strings.TrimSuffix(base, ext) + ".cbz"
	os.Rename(path, outputPath)

	return outputPath
}

func GetAvailableFiles(fileDirectory string) []File {
	entries, err := os.ReadDir(fileDirectory)
	out := []File{}
	if err != nil {
		log.Println(err)
		return out
	}
	for _, file := range entries {
		ext := filepath.Ext(file.Name())
		if ext == ".cbz" {
			info, err := file.Info()
			if err != nil {
				log.Println(err)
				continue
			}
			out = append(out, File{file.Name(), parseFileSize(uint64(info.Size())), ""})
		}
	}
	return out
}

func parseFileSize(size uint64) string {
	var out float32
	if size < 1024 {
		out = float32(size)
		return fmt.Sprintf("%.1fB", out)
	} else if size < 1048576 {
		out = float32(size) / 1024
		return fmt.Sprintf("%.1fKB", out)
	} else if size < 1073741824 {
		out = float32(size) / 1048576
		return fmt.Sprintf("%.1fMB", out)
	} else {
		out = float32(size) / 1073741824
		return fmt.Sprintf("%.1fGB", out)
	}
}

func RandomString(n int) string {
	const letters = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	b := make([]byte, n)
	for i := range b {
		b[i] = letters[rand.Intn(len(letters))]
	}
	return string(b)
}
