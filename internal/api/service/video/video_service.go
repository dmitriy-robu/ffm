package video

import (
	"bufio"
	"context"
	"crypto/sha256"
	"errors"
	"fmt"
	ffmpeg "github.com/u2takey/ffmpeg-go"
	"go-fitness/external/config"
	"go-fitness/external/logger/sl"
	"go-fitness/internal/api/data"
	"go-fitness/internal/api/enum"
	"go-fitness/internal/api/types"
	"io"
	"log/slog"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
)

type TaskQueue interface {
	AddTask(task TranscodeTask)
}

type VideoService struct {
	log       *slog.Logger
	cfg       *config.Config
	videoRepo VideoRepository
	worker    TaskQueue
}

type VideoRepository interface {
	Create(context.Context, types.Video) (int64, error)
	UpdateStatus(context.Context, int64, enum.VideoStatus) error
	GetByUUID(context.Context, string) (*types.Video, error)
	GetList(context.Context, map[string]interface{}) ([]types.Video, error)
	Delete(context.Context, int64) error
	UpdatePoster(context.Context, int64, string) error
	GetListWhereStatusProcessedAndPosterIsNull(context.Context) ([]types.Video, error)
	CheckIfVideoExistByHashName(context.Context, string) bool
}

func NewVideoService(
	log *slog.Logger,
	cfg *config.Config,
	videoRepo VideoRepository,
	worker TaskQueue,
) *VideoService {
	return &VideoService{
		log:       log,
		cfg:       cfg,
		videoRepo: videoRepo,
		worker:    worker,
	}
}

// readFile is a method to read file from the storage path
func (s *VideoService) readFile(path string) ([]byte, error) {
	const op = "Video.readFile"

	log := s.log.With(
		sl.String("op", op),
		sl.String("path", path),
	)

	slurp, err := os.ReadFile(path)
	if err != nil {
		log.Error("failed to read file", sl.Err(err))
		return nil, errors.New("failed to read file")
	}

	return slurp, nil
}

// ProcessGetVideoPlayListByUUID is a method to process video playlist by UUID and return the video file
func (s *VideoService) ProcessGetVideoPlayListByUUID(ctx context.Context, uuid string) ([]byte, error) {
	const op = "Video.ProcessGetVideoM3U8ByUUID"

	log := s.log.With(
		sl.String("op", op),
		sl.String("uuid", uuid),
	)

	video, err := s.videoRepo.GetByUUID(ctx, uuid)
	if err != nil {
		log.Error("failed to get video by uuid", sl.Err(err))
		return nil, errors.New("failed to get video by uuid")
	}

	playListPath := fmt.Sprintf("%s/%s/%s/playlist.m3u8",
		s.cfg.HTTPServer.StoragePath,
		s.cfg.Video.VideoPath,
		video.HashName,
	)

	return s.readFile(playListPath)
}

// ProcessGetVideoM3U8 is a method to process video M3U8 and return the video file
func (s *VideoService) ProcessGetVideoM3U8(url string) ([]byte, error) {
	const op = "Video.ProcessGetVideoM3U8"

	log := s.log.With(
		sl.String("op", op),
		sl.String("url", url),
	)

	//var lastError error

	hashName, resolution, err := s.parseURL(url)
	if err != nil {
		log.Error("failed to parse hash", sl.Err(err))
		return nil, errors.New("failed to parse hash")
	}

	videoPath := fmt.Sprintf("%s/%s/%s/%s",
		s.cfg.HTTPServer.StoragePath,
		s.cfg.Video.VideoPath,
		hashName,
		resolution,
	)

	video, err := s.readFile(videoPath)
	if err != nil {
		log.Error("failed to read file", sl.Err(err))
		//TODO: Implement this
		/*var resolutions = s.cfg.Video.Resolutions

		for _, res := range resolutions {
			if res == resolution {
				continue
			}

			res = fmt.Sprintf("%s.m3u8", res)
			videoPath = fmt.Sprintf("%s/%s/%s/%s", s.cfg.HTTPServer.StoragePath, s.cfg.Video.VideoPath, hashName, res)
			video, err = s.readFile(videoPath)
			if err == nil {
				return video, nil
			}
			lastError = err
		}

		log.Error("failed to find any suitable video file", sl.Err(lastError))*/
		return nil, errors.New("failed to find any suitable video file")
	}

	return video, nil
}

// ProcessGetVideoTS is a method to process video TS and return the video file
func (s *VideoService) ProcessGetVideoTS(ctx context.Context, url string) ([]byte, error) {
	const op = "Video.ProcessGetVideoTS"

	log := s.log.With(
		sl.String("op", op),
		sl.String("url", url),
	)

	hashName, resolution, err := s.parseURL(url)
	if err != nil {
		log.Error("failed to parse hash", sl.Err(err))
		return nil, errors.New("failed to parse hash")
	}

	videoPath := fmt.Sprintf("%s/%s/%s/%s",
		s.cfg.HTTPServer.StoragePath,
		s.cfg.Video.VideoPath,
		hashName,
		resolution,
	)

	return s.readFile(videoPath)
}

// parseURL is a method to parse the URL and return the hash and resolution
func (s *VideoService) parseURL(path string) (string, string, error) {
	const op string = "Video.parseURL"

	log := s.log.With(
		sl.String("op", op),
	)

	paths := strings.SplitN(strings.TrimLeft(path, "/"), "/", -1)

	if len(paths) < 2 {
		log.Error("invalid path", sl.String("path", path))
		return "", "", errors.New("invalid path")
	}

	return paths[len(paths)-2], paths[len(paths)-1], nil
}

// ProcessUpload is a method to process video upload and store it in the storage path
func (s *VideoService) ProcessUpload(
	ctx context.Context,
	data data.VideoData,
) (int64, error) {
	const op string = "Video.ProcessUpload"

	log := s.log.With(
		sl.String("op", op),
	)

	s.generateHash(&data)

	if ok := s.checkIfExistVideoInTableAndInFolderByHashName(ctx, *data.Hash); ok {
		log.Warn("video already exists")
		return 0, errors.New("video_already_exists")
	}

	uploadPath := fmt.Sprintf("%s/%s/%s", s.cfg.HTTPServer.StoragePath, s.cfg.Video.VideoPath, *data.Hash)

	data.UploadPath = uploadPath

	dstPath, err := s.uploadFile(data)
	if err != nil {
		log.Error("failed to upload file", sl.Err(err))
		return 0, errors.New("failed_to_upload_file")
	}

	duration, err := s.getVideoDuration(dstPath)
	if err != nil {
		log.Error("failed to get video duration", sl.Err(err))
		return 0, errors.New("failed_to_get_video_duration")
	}

	posterTime := s.posterTime(duration)
	posterTitle := s.posterTitle(posterTime)

	posterPath := filepath.Join(uploadPath, posterTitle)

	if err = s.createPoster(dstPath, posterPath, posterTime); err != nil {
		log.Error("failed to create poster", sl.Err(err))
		return 0, errors.New("failed_to_create_poster")
	}

	videoID, err := s.videoRepo.Create(ctx, types.Video{
		HashName: *data.Hash,
		Status:   enum.VideoStatusProcessing,
		Duration: duration,
		Poster:   &posterTitle,
	})
	if err != nil {
		log.Error("failed to create video", sl.Err(err))
		return 0, errors.New("failed_to_create_workout")
	}

	s.worker.AddTask(TranscodeTask{
		UploadPath: uploadPath,
		VideoID:    videoID,
		DstPath:    dstPath,
		ChunkHash:  *data.Hash,
	})

	return videoID, nil
}

// GetPosterByUUID is a method to get the poster by UUID
func (s *VideoService) GetPosterByUUID(ctx context.Context, uuid string) (string, error) {
	const op = "Video.GetPosterByUUID"

	log := s.log.With(
		sl.String("op", op),
	)

	video, err := s.videoRepo.GetByUUID(ctx, uuid)
	if err != nil {
		log.Error("failed to get video by uuid", sl.Err(err))
		return "", errors.New("failed to get video by uuid")
	}

	if video.Poster == nil {
		log.Error("poster does not exist")
		return "", errors.New("poster does not exist")
	}

	posterPath := fmt.Sprintf("%s/%s/%s/%s",
		s.cfg.HTTPServer.StoragePath,
		s.cfg.Video.VideoPath,
		video.HashName,
		*video.Poster,
	)

	return posterPath, nil
}

// checkIfExistVideoInTableAndInFolderByHashName is a method to check if the video exists in the table and folder by hash name
func (s *VideoService) checkIfExistVideoInTableAndInFolderByHashName(ctx context.Context, hash string) bool {
	const op = "Video.checkIfExistVideoInTableAndInFolderByHashName"

	log := s.log.With(
		sl.String("op", op),
		sl.String("hash", hash),
	)

	if !s.videoRepo.CheckIfVideoExistByHashName(ctx, hash) {
		log.Warn("video does not exist in database")
		return false
	}

	playlistPath := fmt.Sprintf("%s/%s/%s/playlist.m3u8",
		s.cfg.HTTPServer.StoragePath,
		s.cfg.Video.VideoPath,
		hash,
	)

	if _, err := os.Stat(playlistPath); os.IsNotExist(err) {
		log.Warn("playlist does not exist in folder", sl.String("path", playlistPath))
		return false
	}

	return true
}

// getVideoDuration is a method to get the duration of the video
func (s *VideoService) getVideoDuration(filePath string) (float64, error) {
	const op = "Video.getVideoDuration"

	log := s.log.With(
		sl.String("op", op),
		sl.String("file_path", filePath),
	)

	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-show_entries",
		"format=duration",
		"-of",
		"default=noprint_wrappers=1:nokey=1",
		filePath,
	)

	out, err := cmd.Output()
	if err != nil {
		log.Error("error running ffprobe", sl.Err(err))
		return 0, errors.New("error running ffprobe")
	}

	durationStr := strings.TrimSpace(string(out))

	duration, err := strconv.ParseFloat(durationStr, 64)
	if err != nil {
		log.Error("error converting duration", sl.Err(err))
		return 0, errors.New("error converting duration")
	}

	return duration, nil
}

// CreatePosterFromUploadedTsFiles is a method to create poster from uploaded TS files
func (s *VideoService) CreatePosterFromUploadedTsFiles(ctx context.Context) error {
	const op = "Video.CreatePosterFromUploadedTsFiles"

	log := s.log.With(
		sl.String("op", op),
	)

	videos, err := s.videoRepo.GetListWhereStatusProcessedAndPosterIsNull(ctx)
	if err != nil {
		log.Error("failed to get video by ID", sl.Err(err))
		return errors.New("failed to get video by ID")
	}

	for _, video := range videos {
		posterTime := s.posterTime(video.Duration)
		m3u8Path := fmt.Sprintf("%s/%s/%s/%s",
			s.cfg.HTTPServer.StoragePath,
			s.cfg.Video.VideoPath,
			video.HashName,
			"1080.m3u8",
		)

		tsFilePath, posterTime, err := s.findTsFileForTime(m3u8Path, posterTime)
		if err != nil {
			log.Error("failed to find ts file for time", sl.Err(err))
			return errors.New("failed to find ts file for time")
		}

		posterTitle := s.posterTitle(posterTime)

		posterPath := filepath.Join(s.cfg.HTTPServer.StoragePath, s.cfg.Video.VideoPath, video.HashName, posterTitle)

		log.Info("Creating poster", sl.String("tsFilePath", tsFilePath), sl.Float64("posterTime", posterTime))

		if err = s.createPoster(tsFilePath, posterPath, posterTime); err != nil {
			log.Error("failed to create poster", sl.Err(err))
			return errors.New("failed_to_create_poster")
		}

		if err := s.videoRepo.UpdatePoster(ctx, video.ID, posterTitle); err != nil {
			log.Error("failed to update poster", sl.Err(err))
			return errors.New("failed to update poster")
		}
	}

	return nil
}

// posterTime is a method to get the poster time
func (s *VideoService) posterTime(videoDuration float64) float64 {
	if videoDuration > 60 {
		return (10 * videoDuration) / 100
	}

	return videoDuration / 2
}

// posterTitle is a method to get the poster title
func (s *VideoService) posterTitle(posterTime float64) string {
	return fmt.Sprintf("%f.%s", posterTime, "jpg")
}

// findTsFileForTime is a method to find the TS file for the time
func (s *VideoService) findTsFileForTime(m3u8Path string, globalTime float64) (string, float64, error) {
	const op = "Video.findTsFileForTime"

	log := s.log.With(
		sl.String("op", op),
		sl.String("m3u8Path", m3u8Path),
		sl.Float64("globalTime", globalTime),
	)

	file, err := os.Open(m3u8Path)
	if err != nil {
		log.Error("failed to open m3u8 file", sl.Err(err))
		return "", 0, err
	}
	defer func(file *os.File) {
		err := file.Close()
		if err != nil {
			log.Error("failed to close file", sl.Err(err))
		}
	}(file)

	scanner := bufio.NewScanner(file)
	var cumulativeTime float64
	var tsFile string
	for scanner.Scan() {
		line := scanner.Text()
		if strings.HasPrefix(line, "#EXTINF:") {
			log.Info("line", sl.String("line", line))

			durationStr := strings.TrimPrefix(line, "#EXTINF:")

			durationStr = strings.TrimSuffix(durationStr, ",")

			duration, err := strconv.ParseFloat(durationStr, 64)
			if err != nil {
				log.Error("failed to parse duration", sl.Err(err))
				return "", 0, err
			}

			if cumulativeTime+duration >= globalTime {
				adjustedTime := globalTime - cumulativeTime

				return filepath.Join(filepath.Dir(m3u8Path), tsFile), adjustedTime, nil
			}
			cumulativeTime += duration
		} else if strings.HasSuffix(line, ".ts") {
			log.Info("line", sl.String("line", line))
			tsFile = line
		}
	}

	return "", 0, errors.New("could not find appropriate ts file")
}

// createPoster is a method to create poster for the video
func (s *VideoService) createPoster(videoPath string, posterPath string, posterTime float64) error {
	const op string = "Video.createPoster"

	log := s.log.With(
		sl.String("op", op),
		sl.String("video_path", videoPath),
		sl.String("poster_path", posterPath),
		sl.Float64("poster_time", posterTime),
	)

	seconds := fmt.Sprintf("%.2f", posterTime)

	cmd := ffmpeg.Input(videoPath).
		Filter("trim", ffmpeg.Args{fmt.Sprintf("start=%s", seconds)}).
		Output(posterPath, ffmpeg.KwArgs{"vframes": 1, "format": "image2", "vcodec": "mjpeg", "y": ""}).
		ErrorToStdOut()

	if err := cmd.Run(); err != nil {
		log.Error("failed to create poster", sl.Err(err))
		return errors.New("failed_to_create_poster")
	}

	return nil
}

// uploadFile is a method to upload file to the storage path
func (s *VideoService) uploadFile(data data.VideoData) (string, error) {
	const op string = "Video.uploadFile"

	log := s.log.With(
		sl.String("op", op),
		sl.Any("data", data),
	)

	if err := os.MkdirAll(data.UploadPath, 0755); err != nil {
		log.Error("failed to create upload directory", sl.Err(err))
		return "", errors.New("failed to create upload directory")
	}

	if _, err := os.Stat(data.UploadPath); os.IsNotExist(err) {
		err := os.Mkdir(data.UploadPath, os.ModePerm)
		if err != nil {
			log.Error("failed to create upload directory", sl.Err(err))
			return "", errors.New("failed to create upload directory")
		}
	}

	dstPath := filepath.Join(data.UploadPath, data.Header.Filename)

	dst, err := os.Create(dstPath)
	if err != nil {
		log.Error("failed to create file", sl.Err(err))
		return "", errors.New("failed to create file")
	}
	defer func(dst *os.File) {
		if err := dst.Close(); err != nil {
			log.Error("failed to close file", sl.Err(err))
		}
	}(dst)

	if _, err = io.Copy(dst, data.File); err != nil {
		log.Error("failed to copy file", sl.Err(err))
		return "", errors.New("failed to copy file")
	}

	return dstPath, nil
}

// DeleteAllVideoFilesIfDoestExistInTable is a method to delete all video files if does not exist in table
func (s *VideoService) DeleteAllVideoFilesIfDoestExistInTable(ctx context.Context) {
	const op string = "Video.DeleteAllVideoFilesIfDoestExistInTable"

	log := s.log.With("op", op)

	videos, err := s.videoRepo.GetList(ctx, nil)
	if err != nil {
		log.Error("failed to get videos", sl.Err(err))
		return
	}

	videoMap := make(map[string]bool)
	for _, video := range videos {
		videoMap[video.HashName] = true
	}

	videoFilesPath := filepath.Join(s.cfg.HTTPServer.StoragePath, s.cfg.Video.VideoPath)
	entries, err := os.ReadDir(videoFilesPath)
	if err != nil {
		log.Error("failed to read video directory", sl.Err(err))
		return
	}

	for _, entry := range entries {
		if entry.IsDir() {
			log.Info("found video file", entry.Name())
			if !videoMap[entry.Name()] {
				videoPath := filepath.Join(videoFilesPath, entry.Name())
				err := os.RemoveAll(videoPath)
				if err != nil {
					log.Error("failed to delete video file", entry.Name(), sl.Err(err))
				} else {
					log.Info("deleted video file", entry.Name())
				}
			}
		}
	}
}

func (s *VideoService) generateHash(data *data.VideoData) {
	const op string = "Video.generateHash"

	log := s.log.With(
		sl.String("op", op),
	)

	if data.Hash == nil {
		data.Hash = new(string)
	}

	h := sha256.New()
	h.Write([]byte(fmt.Sprintf("%s-%d", data.Header.Filename, data.Header.Size)))

	hash := fmt.Sprintf("%x", h.Sum(nil))

	log.Info("generated hash", sl.String("hash", hash))

	*data.Hash = hash
}

func (s *VideoService) UpdateStatus(ctx context.Context, videoID int64, status enum.VideoStatus) error {
	const op string = "Video.UpdateStatus"

	log := s.log.With(
		sl.String("op", op),
		sl.Int64("video_id", videoID),
		sl.String("status", status.String()),
	)

	if err := s.videoRepo.UpdateStatus(ctx, videoID, status); err != nil {
		log.Error("failed to update video status", sl.Err(err))
		return errors.New("failed to update video status")
	}

	return nil
}
