package fileupload

import (
	"go-fitness/external/logger/sl"
	"go-fitness/internal/api/data"
	"log/slog"
	"mime/multipart"
	"net/http"
)

func ParseAndExtractFile(r *http.Request, log *slog.Logger) (*data.VideoData, error) {
	const op string = "FileUploadService.UploadFile"

	log = log.With(
		sl.String("op", op),
	)

	if err := r.ParseMultipartForm(8 << 30); err != nil {
		log.Error("failed_to_parse_multipart_form", sl.Err(err))
		return nil, err
	}

	file, header, err := r.FormFile("file")
	if err != nil {
		log.Error("failed_to_get_form_file", sl.Err(err))
		return nil, err
	}
	defer func(file multipart.File) {
		if err := file.Close(); err != nil {
			log.Error("failed_to_close_file", sl.Err(err))
		}
	}(file)

	videoData := &data.VideoData{
		File:   file,
		Header: header,
	}

	return videoData, nil
}
