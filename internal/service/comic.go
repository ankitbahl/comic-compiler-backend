package service

import (
	"github.com/ankitbahl/comic-compiler-backend/internal/helpers"
)

type Comic struct {
	Name string `json:"name"`
	Size string `json:"size"`
}

const fileDirectory = "/comics/"

func GetAllComics() []Comic {
	return filesToComic(helpers.GetFileList(fileDirectory))
}

func GetComicInfo(name string) []Comic {
	comicList := helpers.GetFileList(fileDirectory)
	var comicPath string
	for _, comic := range comicList {
		if comic.Name == name {
			comicPath = comic.Path
			break
		}
	}
	if comicPath == "" {
		return []Comic{}
	}

	return filesToComic(helpers.GetZippedFiles(comicPath))

}

func CompileZip(name string, files []string, token string, job *helpers.Job) string {
	job.Progress = helpers.Progress{ProgressStatus: helpers.Initializing, Completion: 0}
	comicList := helpers.GetFileList(fileDirectory)
	var comicPath string
	for _, comic := range comicList {
		if comic.Name == name {
			comicPath = comic.Path
			break
		}
	}
	if comicPath == "" {
		return ""
	}
	outPath := helpers.UnzipFiles(comicPath, files, name)
	if outPath == "" {
		return ""
	}
	helpers.RenameAllCBFiles(outPath)
	job.Progress = helpers.Progress{ProgressStatus: helpers.Extracting, Completion: 0}
	helpers.ProcessZipsAndRars(outPath, job)
	job.Progress = helpers.Progress{ProgressStatus: helpers.Compiling, Completion: 0}
	outputPath := helpers.ZipImages(outPath, token, job)
	helpers.CleanFolder(outPath)
	outputPath = helpers.RenameOutput(outputPath)
	return outputPath
}

func filesToComic(files []helpers.File) []Comic {
	comics := make([]Comic, len(files))
	for i, file := range files {
		comics[i] = Comic{Name: file.Name, Size: file.Size}
	}
	return comics
}
