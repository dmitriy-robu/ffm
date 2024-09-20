package data

import "mime/multipart"

type VideoData struct {
	File       multipart.File
	Header     *multipart.FileHeader
	Hash       *string
	FileID     *string
	UploadPath string
}
