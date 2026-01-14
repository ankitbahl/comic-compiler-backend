package helpers

import (
	"archive/zip"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/nwaples/rardecode"
)

func GetZippedFiles(path string) []File {
	r, err := zip.OpenReader(path)
	if err != nil {
		return []File{}
	}
	defer r.Close()
	out := []File{}
	for _, f := range r.File {
		if f.UncompressedSize64 != 0 {
			out = append(out, File{
				Name: f.Name,
				Size: parseFileSize(f.UncompressedSize64),
			})
		}
	}

	return out
}

func UnRarFiles(rarPath string, destFolder string) bool {
	f, err := os.Open(rarPath)
	if err != nil {
		log.Println(err)
		return false
	}

	defer f.Close()

	r, err := rardecode.NewReader(f, "")
	if err != nil {
		log.Println(err)
		return false
	}

	for {
		hdr, err := r.Next()
		if err == io.EOF {
			break
		}

		if err != nil {
			log.Println(err)
			return false
		}

		outPath := filepath.Join(destFolder, hdr.Name)

		if hdr.IsDir {
			if err := os.MkdirAll(outPath, hdr.Mode()); err != nil {
				log.Println(err)
				return false
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(outPath), 0755); err != nil {
			log.Println(err)
			return false
		}

		outFile, err := os.OpenFile(outPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, hdr.Mode())
		if err != nil {
			log.Println(err)
			return false
		}

		if _, err := io.Copy(outFile, r); err != nil {
			outFile.Close()
			log.Println(err)
			return false
		}

		outFile.Close()
	}

	return true
}

func UnzipFiles(zipPath string, files []string, destFolder string) string {
	fileMap := map[string]bool{}
	for _, file := range files {
		fileMap[file] = true
	}
	dest := "/comics/compiled/" + destFolder

	res := UnzipFilesImpl(zipPath, fileMap, dest)
	if !res {
		return ""
	}
	return dest
}

func UnzipFilesImpl(zipPath string, fileMap map[string]bool, destFolder string) bool {
	r, err := zip.OpenReader(zipPath)
	if err != nil {
		log.Println(err)
		return false
	}
	defer r.Close()

	for _, f := range r.File {
		if len(fileMap) > 0 && !fileMap[f.Name] {
			continue
		}
		fPath := filepath.Join(destFolder, f.Name)

		if f.FileInfo().IsDir() {
			if err := os.MkdirAll(fPath, f.Mode()); err != nil {
				log.Println(err)
				return false
			}
			continue
		}

		if err := os.MkdirAll(filepath.Dir(fPath), 0755); err != nil {
			log.Println(err)
			return false
		}

		srcFile, err := f.Open()
		if err != nil {
			log.Println(err)
			return false
		}

		destFile, err := os.OpenFile(fPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, f.Mode())
		if err != nil {
			srcFile.Close()
			log.Println(err)
			return false
		}

		_, err = io.Copy(destFile, srcFile)
		if err != nil {
			srcFile.Close()
			destFile.Close()
			log.Println(err)
			return false
		}

		srcFile.Close()
		destFile.Close()
	}

	return true
}

func ZipFiles(outputPath string, files []string, job *Job) bool {
	out, err := os.Create(outputPath)
	if err != nil {
		log.Println(err)
		return false
	}
	defer out.Close()

	zw := zip.NewWriter(out)
	defer zw.Close()

	seen := make(map[string]int)
	for i, filePath := range files {
		err := addFileToZip(zw, filePath, seen)
		if err != nil {
			log.Println(err)
			return false
		}

		job.Progress.Completion = float32(i) / float32(len(files))
	}

	return true
}

func addFileToZip(zw *zip.Writer, filePath string, seen map[string]int) error {
	file, err := os.Open(filePath)
	if err != nil {
		return err
	}
	defer file.Close()

	info, err := file.Stat()
	if err != nil {
		return err
	}

	header, err := zip.FileInfoHeader(info)
	if err != nil {
		return err
	}

	origName := filepath.Base(filePath)
	ext := filepath.Ext(origName)
	base := strings.TrimSuffix(origName, ext)

	count := seen[origName]
	name := origName
	if count > 0 {
		name = fmt.Sprintf("%s_%d%s", base, count, ext)
	}

	seen[origName] = count + 1

	header.Name = name

	// Compression method
	header.Method = zip.Store

	writer, err := zw.CreateHeader(header)
	if err != nil {
		return err
	}

	_, err = io.Copy(writer, file)
	return err
}
