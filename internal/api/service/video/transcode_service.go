package video

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"go-fitness/external/config"
	"go-fitness/external/logger/sl"
	"go-fitness/internal/api/enum"
	"log/slog"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"
)

type UpdateVideoStatus interface {
	UpdateStatus(context.Context, int64, enum.VideoStatus) error
}

type TranscodeService struct {
	log   *slog.Logger
	cfg   *config.Config
	video UpdateVideoStatus
}

func NewTranscodeService(
	log *slog.Logger,
	cfg *config.Config,
	video UpdateVideoStatus,
) *TranscodeService {
	return &TranscodeService{
		log:   log,
		cfg:   cfg,
		video: video,
	}
}

// ProcessTranscode is a method to process video transcoding and chunking
func (s *TranscodeService) ProcessTranscode(
	ctx context.Context,
	transcode TranscodeTask,
) error {
	const op = "TranscodeService.processTranscode"

	log := s.log.With(
		sl.String("op", op),
	)

	uploadSuccessful := false

	defer func() {
		if !uploadSuccessful {
			if updateErr := s.video.UpdateStatus(ctx, transcode.VideoID, enum.VideoStatusFailed); updateErr != nil {
				log.Error("failed to update video status to failed", sl.Err(updateErr))
			}

			if rmErr := os.RemoveAll(transcode.UploadPath); rmErr != nil {
				log.Error("failed to remove folder", sl.Err(rmErr))
			}
		}
	}()

	if err := s.transcodeAndChunk(transcode.UploadPath, transcode.DstPath); err != nil {
		log.Error("failed to transcode and chunk video", sl.Err(err))
		return errors.New("failed to transcode and chunk video")
	}

	if err := s.createMasterM8U3PlayList(transcode.UploadPath, transcode.ChunkHash); err != nil {
		log.Error("failed to create master m8u3 playlist", sl.Err(err))
		return errors.New("failed to create master m8u3 playlist")
	}

	if err := s.video.UpdateStatus(ctx, transcode.VideoID, enum.VideoStatusProcessed); err != nil {
		log.Error("failed to update video status to processed", sl.Err(err))
		return errors.New("failed to update video status to processed")
	}

	uploadSuccessful = true

	return nil
}

// transcodeAndChunk is a method to transcode and chunk video into smaller segments using ffmpeg
func (s *TranscodeService) transcodeAndChunk(uploadPath string, videoPath string) error {
	const op = "TranscodeService.transcodeAndChunk"

	log := s.log.With(
		sl.String("op", op),
		sl.String("upload_path", uploadPath),
		sl.String("video_path", videoPath),
	)

	log.Info("transcoding and chunking video")

	resolutions := s.cfg.Video.Resolutions

	videoDimensions, err := s.getVideoDimensions(videoPath)
	if err != nil {
		log.Error("failed to get video dimensions", sl.Err(err))
		return err
	}

	sort.Slice(resolutions, func(i, j int) bool {
		heightI, _ := strconv.Atoi(resolutions[i])
		heightJ, _ := strconv.Atoi(resolutions[j])
		return heightI < heightJ
	})

	videoCodec, err := s.getVideoCodec(videoPath)
	if err != nil {
		log.Error("failed to get video codec", sl.Err(err))
		return err
	}

	if videoCodec == "hevc" || videoCodec == "h265" {
		outputPath, err := s.h265ToH264(videoPath, uploadPath)
		if err != nil {
			log.Error("failed to transcode video from hevc to h264", sl.Err(err))
			return err
		}

		videoPath = outputPath
	}

	for _, res := range resolutions {
		scaleParam := fmt.Sprintf("scale=-2:%s", res)

		if videoDimensions["height"] > videoDimensions["width"] {
			scaleParam = fmt.Sprintf("scale=%s:-2", res)
		}

		if err = s.transcodeVideoCMD(videoPath, uploadPath, res, scaleParam); err != nil {
			log.Error("failed to transcode video", sl.Err(err))
			return err
		}
	}

	go func(deleteVideo string) {
		if err = os.Remove(deleteVideo); err != nil {
			log.Error("failed to remove file after successful upload", sl.Err(err))
		}
	}(videoPath)

	return nil
}

func (s *TranscodeService) h265ToH264(videoPath, uploadPath string) (string, error) {
	const op = "TranscodeService.h265ToH264"

	log := s.log.With(
		sl.String("op", op),
		sl.String("uploadPath", uploadPath),
	)

	log.Info("transcoding video from h265 to h264")

	outputPath := fmt.Sprintf("%s/h264.mp4", uploadPath)

	cmd := exec.Command("ffmpeg", "-i", videoPath,
		"-c:v", "libx264",
		"-preset", "ultrafast",
		"-crf", "23",
		"-c:a", "aac",
		"-b:a", "128k",
		outputPath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Error("failed to transcode video from hevc to h264",
			sl.Err(err),
			sl.String("stdout", out.String()),
			sl.String("stderr", stderr.String()),
		)
		return "", err
	}

	if err := os.Remove(videoPath); err != nil {
		log.Error("failed to remove videoPath", sl.Err(err))
		return "", err
	}

	return outputPath, nil
}

// getVideoCodec is a method to get the codec of the video
func (s *TranscodeService) getVideoCodec(videoPath string) (string, error) {
	const op string = "VideoService.getVideoCodec"

	log := s.log.With(
		sl.String("op", op),
		sl.String("video_path", videoPath),
	)

	cmd := exec.Command(
		"ffprobe",
		"-v", "error",
		"-select_streams", "v:0",
		"-show_entries",
		"stream=codec_name",
		"-of", "default=noprint_wrappers=1:nokey=1",
		videoPath)

	var out bytes.Buffer
	cmd.Stdout = &out
	if err := cmd.Run(); err != nil {
		log.Error("failed to get video codec", sl.Err(err))
		return "", errors.New("failed to get video codec")
	}

	codec := strings.TrimSpace(out.String())

	return codec, nil
}

// createMasterM8U3PlayList is a method to create master m8u3 playlist
func (s *TranscodeService) createMasterM8U3PlayList(uploadPath string, chunkHash string) error {
	const op = "TranscodeService.createMasterM8U3PlayList"

	log := s.log.With(
		sl.String("op", op),
		sl.String("upload_path", uploadPath),
	)

	log.Info("creating master m8u3 playlist")

	masterM8U3PlayListPath := fmt.Sprintf("%s/%s.m3u8", uploadPath, "playlist")

	masterM8U3PlayList, err := os.Create(masterM8U3PlayListPath)
	if err != nil {
		log.Error("failed to create master m8u3 playlist", sl.Err(err))
		return errors.New("failed to create master m8u3 playlist")
	}
	defer func(masterM8U3PlayList *os.File) {
		if err := masterM8U3PlayList.Close(); err != nil {
			log.Error("failed to close master m8u3 playlist", sl.Err(err))
		}
	}(masterM8U3PlayList)

	var buffer bytes.Buffer
	buffer.WriteString("#EXTM3U\n")
	buffer.WriteString("#EXT-X-VERSION:3\n")
	buffer.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=800000,RESOLUTION=640x360\n")
	buffer.WriteString(chunkHash + "/360.m3u8\n")
	buffer.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=1400000,RESOLUTION=854x480\n")
	buffer.WriteString(chunkHash + "/480.m3u8\n")
	buffer.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=2800000,RESOLUTION=1280x720\n")
	buffer.WriteString(chunkHash + "/720.m3u8\n")
	buffer.WriteString("#EXT-X-STREAM-INF:BANDWIDTH=5000000,RESOLUTION=1920x1080\n")
	buffer.WriteString(chunkHash + "/1080.m3u8\n")

	if _, err := masterM8U3PlayList.Write(buffer.Bytes()); err != nil {
		log.Error("failed to write master m8u3 playlist", sl.Err(err))
		return errors.New("failed to write master m8u3 playlist")
	}

	return nil
}

// getVideoDimensions is a method to get the dimensions of the video
func (s *TranscodeService) getVideoDimensions(videoPath string) (map[string]int, error) {
	const op string = "TranscodeService.getVideoDimensions"

	log := s.log.With(
		sl.String("op", op),
		sl.String("video_path", videoPath),
	)

	cmd := exec.Command(
		"ffprobe",
		"-v",
		"error",
		"-select_streams",
		"v:0",
		"-show_entries",
		"stream=width,height",
		"-of",
		"csv=p=0",
		videoPath,
	)

	output, err := cmd.Output()
	if err != nil {
		log.Error("failed to get video dimensions", sl.Err(err))
		return nil, errors.New("failed to get video dimensions")
	}

	dimensions := strings.Split(strings.TrimSpace(string(output)), ",")
	if len(dimensions) != 2 {
		log.Error("invalid video dimensions", sl.String("dimensions", string(output)))
		return nil, errors.New("invalid video dimensions")
	}

	width, err := strconv.Atoi(dimensions[0])
	if err != nil {
		log.Error("failed to convert width", sl.Err(err))
		return nil, errors.New("failed to convert width")
	}

	height, err := strconv.Atoi(dimensions[1])
	if err != nil {
		log.Error("failed to convert height", sl.Err(err))
		return nil, errors.New("failed to convert height")
	}

	return map[string]int{"width": width, "height": height}, nil
}

// transcodeVideoCMD is a method to transcode video using ffmpeg
func (s *TranscodeService) transcodeVideoCMD(
	videoPath,
	uploadPath,
	resolution,
	scaleParam string,
) error {
	const op string = "transcodeVideoCMD.transcodeVideoCMD"

	log := s.log.With(
		sl.String("op", op),
		sl.String("video_path", videoPath),
		sl.String("upload_path", uploadPath),
		sl.String("resolution", resolution),
		sl.String("scale_param", scaleParam),
	)

	log.Info("transcoding video", sl.String("resolution", resolution))

	outputPath := fmt.Sprintf("%s/%s.m3u8", uploadPath, resolution)

	segmentFilename := fmt.Sprintf("%s/%s_%%03d.ts", uploadPath, resolution)
	cmd := exec.Command("ffmpeg", "-i", videoPath,
		"-pix_fmt", "yuv420p", // Преобразуем видео в 8-битный формат
		"-profile:v", "main",
		"-level", "3.1",
		"-preset", "veryfast",
		//"-preset", "ultrafast",
		"-vf", scaleParam,
		"-start_number", "0",
		"-hls_time", "10",
		"-hls_list_size", "0",
		"-f", "hls",
		"-hls_segment_filename",
		segmentFilename,
		outputPath)

	var out bytes.Buffer
	var stderr bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &stderr

	if err := cmd.Run(); err != nil {
		log.Error("failed to transcode video",
			sl.String("resolution", resolution),
			sl.String("stdout", out.String()),
			sl.String("stderr", stderr.String()),
			sl.Err(err))
		return errors.New("failed to transcode video for resolution " + resolution)
	}

	log.Info("transcode video done", sl.String("resolution", resolution))

	return nil
}
