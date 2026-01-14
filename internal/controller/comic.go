package controller

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"strings"
	"sync"

	"github.com/ankitbahl/comic-compiler-backend/internal/helpers"
	"github.com/ankitbahl/comic-compiler-backend/internal/service"
)

func WriteJSON(w http.ResponseWriter, status int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(data)
}

func Comics(w http.ResponseWriter, r *http.Request) {
	resp := service.GetAllComics()
	WriteJSON(w, http.StatusOK, resp)
}

func ComicInfo(w http.ResponseWriter, r *http.Request) {
	comicName := r.URL.Query().Get("comic")
	if len(comicName) == 0 {
		w.WriteHeader(http.StatusBadRequest)
	} else {
		resp := service.GetComicInfo(comicName)
		WriteJSON(w, http.StatusOK, resp)
	}
}

type CompileBody struct {
	Files []string `json:"files"`
}

type CompileRes struct {
	JobID string `json:"job_id"`
}

var jobs = sync.Map{}

func CompileComic(w http.ResponseWriter, r *http.Request) {
	comicName := r.URL.Query().Get("comic")
	var reqBody CompileBody

	if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
		log.Println(err)
		http.Error(w, "Invalid JSON", http.StatusBadRequest)
	}

	defer r.Body.Close()
	token := strings.Split(comicName, ".")[0]
	job := &helpers.Job{ID: token, Status: helpers.JobPending}
	jobs.Store(token, job)

	go func() {
		job.Status = helpers.JobRunning
		outputPath := service.CompileZip(comicName, reqBody.Files, token, job)
		job.Status = helpers.JobDone
		job.OutPath = outputPath
	}()
	WriteJSON(w, http.StatusOK, CompileRes{JobID: token})
}

type ProgressRes struct {
	JobStatus   string           `json:"jobStatus"`
	JobProgress helpers.Progress `json:"jobProgress"`
}

func GetJobProgress(w http.ResponseWriter, r *http.Request) {
	jobId := r.URL.Query().Get("job_id")
	val, ok := jobs.Load(jobId)
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	job, ok := val.(*helpers.Job)
	if !ok {
		log.Println("Error with job parsing")
		w.WriteHeader(http.StatusBadRequest)
		return
	}
	WriteJSON(w, http.StatusOK, ProgressRes{JobStatus: string(job.Status), JobProgress: job.Progress})
}

func DownloadComic(w http.ResponseWriter, r *http.Request) {
	comic := r.URL.Query().Get("comic")
	outputPath := filepath.Join("/comics/compiled", comic)
	http.ServeFile(w, r, outputPath)
}

func GetDownloadableComics(w http.ResponseWriter, r *http.Request) {
	files := helpers.GetAvailableFiles("/comics/compiled")
	WriteJSON(w, http.StatusOK, files)
}
