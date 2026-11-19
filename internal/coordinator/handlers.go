
package coordinator

import (
	"encoding/json"
	"net/http"
    "log"
    "mime"
    "strconv"
)

func (s *Service) UploadHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	r.Body = http.MaxBytesReader(w, r.Body, 1024*1024*1024)

	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Invalid upload or file too large", http.StatusBadRequest)
		return
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		http.Error(w, "Missing file field", http.StatusBadRequest)
		return
	}
	defer file.Close()

	contentType := header.Header.Get("Content-Type")
	if contentType == "" {
		contentType = "application/octet-stream"
	}

	result, err := s.Upload(
		r.Context(),
		header.Filename,
		contentType,
		header.Size,
		file,
	)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusCreated)

	_ = json.NewEncoder(w).Encode(result)
}



func (s *Service) DownloadHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	fileID := r.PathValue("id")

	file, _, err := s.GetFile(r.Context(), fileID)
	if err != nil {
		http.Error(w, "File not found", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", file.ContentType)
	w.Header().Set(
		"Content-Disposition",
		mime.FormatMediaType("attachment", map[string]string{
			"filename": file.Name,
		}),
	)
	w.Header().Set("Content-Length", strconv.FormatInt(file.Size, 10))

	_, err = s.Download(r.Context(), fileID, w)
	if err != nil {
		// The response may already contain partial file data.
		// Log the error; we cannot change the HTTP status now.
		log.Printf("download %s failed: %v", fileID, err)
	}
}

func (s *Service) DeleteHandler(
	w http.ResponseWriter,
	r *http.Request,
) {
	fileID := r.PathValue("id")

	if err := s.Delete(r.Context(), fileID); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
