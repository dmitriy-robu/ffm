package service

import (
	"context"
	"encoding/json"
	"fmt"
	"go-fitness/external/logger/sl"
	"go-fitness/internal/api/data"
	"go-fitness/internal/api/types"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/drive/v3"
	"google.golang.org/api/option"
	"io"
	"log/slog"
	"mime/multipart"
	"net/http"
	"os"
	"strconv"
	"strings"
)

type GoogleDriveService struct {
	log               *slog.Logger
	config            *oauth2.Config
	workoutService    *WorkoutService
	programRepository ProgramRepository
}

type ProgramRepository interface {
	GetProgramByName(context.Context, string) (types.Program, error)
	GetGoalByName(context.Context, string) (types.Goal, error)
	GetLevelByName(context.Context, string) (types.Level, error)
	GetPeriodByName(context.Context, string) (types.Period, error)
	CreateProgramMonth(context.Context, int64, int) (int64, error)
	GetProgramMonth(context.Context, int64, int) (types.ProgramMonth, error)
}

type FileInfo struct {
	Name     string
	Id       string
	MimeType string
	Children []*FileInfo
}

func NewGoogleDriveService(
	log *slog.Logger,
	workoutService *WorkoutService,
	programRepository ProgramRepository,
) *GoogleDriveService {
	return &GoogleDriveService{
		log:               log,
		programRepository: programRepository,
		workoutService:    workoutService,
	}
}

func (s *GoogleDriveService) initConfig() error {
	const op string = "service.GoogleDriveService.initConfig"

	log := s.log.With(
		sl.String("op", op),
	)

	b, err := os.ReadFile("client_secret.json")
	if err != nil {
		log.Error("Unable to read client secret file", sl.Err(err))
		return err
	}

	config, err := google.ConfigFromJSON(b, drive.DriveReadonlyScope)
	if err != nil {
		log.Error("Unable to parse client secret file to config", sl.Err(err))
		return err
	}

	s.config = config
	return nil
}

func (s *GoogleDriveService) GetAuthURL(ctx context.Context) string {
	const op string = "service.GoogleDriveService.GetAuthURL"

	log := s.log.With(
		sl.String("op", op),
	)

	if err := s.initConfig(); err != nil {
		log.Error("Failed to initialize config", sl.Err(err))
		return ""
	}

	authURL := s.config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	authURL = strings.ReplaceAll(authURL, "\\u0026", "&")
	log.Info("Auth URL generated", sl.String("authURL", authURL))

	return authURL
}

func (s *GoogleDriveService) ExchangeCodeForToken(ctx context.Context, code string) error {
	const op string = "service.GoogleDriveService.ExchangeCodeForToken"

	log := s.log.With(
		sl.String("op", op),
	)

	tok, err := s.config.Exchange(ctx, code)
	if err != nil {
		log.Error("Unable to exchange code for token", sl.Err(err))
		return err
	}

	return s.saveToken("token.json", tok)
}

func (s *GoogleDriveService) collectVideoData(folder *FileInfo, videoDataList *[]data.WorkoutData, parentPath string) {
	for _, file := range folder.Children {
		if file.MimeType == "application/vnd.google-apps.folder" {
			s.collectVideoData(file, videoDataList, fmt.Sprintf("%s|%s", parentPath, file.Name))
		} else {
			paths := strings.Split(parentPath, "|")
			name := strings.TrimSuffix(file.Name, ".mp4")
			description := fmt.Sprintf("%s", strings.Join(paths[1:], "|"))
			videoData := data.WorkoutData{
				Name:        name,
				Description: description,
				VideoData: &data.VideoData{
					FileID: &file.Id,
				},
			}
			*videoDataList = append(*videoDataList, videoData)
		}
	}
}

func (s *GoogleDriveService) ProcessParse(ctx context.Context) {
	const op string = "service.GoogleDriveService.ProcessParse"

	log := s.log.With(
		sl.String("op", op),
	)

	log.Info("Processing Google Drive")

	client := s.getClient()
	if client == nil {
		log.Error("Client is nil, unable to proceed")
		return
	}

	srv, err := drive.NewService(ctx, option.WithHTTPClient(client))
	if err != nil {
		log.Error("Unable to retrieve drive Client", sl.Err(err))
		return
	}

	folderID, err := s.findFolderID(srv, "fitness-videos")
	if err != nil {
		log.Error("Unable to find folder", sl.Err(err))
		return
	}

	files, err := s.listFilesInFolder(srv, folderID)
	if err != nil {
		log.Error("Unable to list files in folder", sl.Err(err))
		return
	}

	rootFolder := &FileInfo{
		Name:     "fitness-videos",
		Id:       folderID,
		MimeType: "application/vnd.google-apps.folder",
		Children: files,
	}

	var videoUploadDataList []data.WorkoutData
	s.collectVideoData(rootFolder, &videoUploadDataList, "fitness-videos")

	for _, videoData := range videoUploadDataList {
		fmt.Printf("Name: %s\nDescription: %s\n", videoData.Name, videoData.Description)

		description := strings.Split(videoData.Description, "|")

		programName := description[0]
		/*	goalName := description[1]
			levelName := description[2]*/
		monthName, _ := strconv.Atoi(description[3])
		//periodName := description[4]

		program, err := s.programRepository.GetProgramByName(ctx, programName)
		if err != nil {
			log.Error("Unable to retrieve program", sl.Err(err))
			return
		}

		var prMonthID int64
		programMonth, err := s.programRepository.GetProgramMonth(ctx, program.ID, monthName)
		if err != nil {
			log.Error("Unable to retrieve program month", sl.Err(err))

			prMonthID, err = s.programRepository.CreateProgramMonth(ctx, program.ID, monthName)
			if err != nil {
				log.Error("Unable to create program month", sl.Err(err))
				return
			}
		} else {
			prMonthID = programMonth.ID
		}

		tmpFile, err := s.downloadTempFile(srv, *videoData.VideoData.FileID, videoData.Name)
		if err != nil {
			log.Error("Unable to download file", sl.Err(err))
			continue
		}
		defer func(name string) {
			err := os.Remove(name)
			if err != nil {
				log.Error("Unable to remove temporary file", sl.Err(err))
			}
		}(tmpFile.Name())

		fi, err := tmpFile.Stat()
		if err != nil {
			log.Error("Unable to retrieve file information", sl.Err(err))
			continue
		}

		fileHeader := &multipart.FileHeader{
			Filename: videoData.Name,
			Size:     fi.Size(),
		}
		file, err := os.Open(tmpFile.Name())
		if err != nil {
			log.Error("Unable to open temporary file", sl.Err(err))
			continue
		}
		defer file.Close()

		if err := s.workoutService.ProcessWorkout(ctx, data.WorkoutData{
			Name:           videoData.Name,
			Description:    videoData.Description,
			ProgramMonthID: &prMonthID,
			VideoData: &data.VideoData{
				File:   file,
				Header: fileHeader,
			},
		}); err != nil {
			log.Error("Unable to store workout", sl.Err(err))
		}
	}
}

func (s *GoogleDriveService) downloadTempFile(srv *drive.Service, fileId, fileName string) (*os.File, error) {
	const op string = "service.GoogleDriveService.downloadTempFile"

	log := s.log.With(
		sl.String("op", op),
	)

	resp, err := srv.Files.Get(fileId).Download()
	if err != nil {
		log.Error("Unable to retrieve file data", sl.Err(err))
		return nil, err
	}
	defer resp.Body.Close()

	tmpFile, err := os.CreateTemp("", fileName)
	if err != nil {
		log.Error("Unable to create temporary file", sl.Err(err))
		return nil, err
	}

	if _, err := io.Copy(tmpFile, resp.Body); err != nil {
		log.Error("Unable to write file data to disk", sl.Err(err))
		return nil, err
	}

	if _, err := tmpFile.Seek(0, 0); err != nil {
		log.Error("Unable to reset file pointer", sl.Err(err))
		return nil, err
	}

	log.Info("Temporary file created", sl.String("file", tmpFile.Name()))
	return tmpFile, nil
}

func (s *GoogleDriveService) printFolderStructure(folder *FileInfo, level int) {
	indent := strings.Repeat("  ", level)
	fmt.Printf("%s%s (%s)\n", indent, folder.Name, folder.Id)
	for _, file := range folder.Children {
		if file.MimeType == "application/vnd.google-apps.folder" {
			s.printFolderStructure(file, level+1)
		} else {
			fmt.Printf("%s  %s (%s)\n", indent, file.Name, file.Id)
		}
	}
}

func (s *GoogleDriveService) findFolderID(srv *drive.Service, folderName string) (string, error) {
	query := fmt.Sprintf("mimeType='application/vnd.google-apps.folder' and name='%s' and trashed=false", folderName)
	r, err := srv.Files.List().Q(query).Fields("files(id, name)").Do()
	if err != nil {
		return "", err
	}

	if len(r.Files) == 0 {
		return "", fmt.Errorf("folder '%s' not found", folderName)
	}

	return r.Files[0].Id, nil
}

func (s *GoogleDriveService) listFilesInFolder(srv *drive.Service, folderID string) ([]*FileInfo, error) {
	var files []*FileInfo
	query := fmt.Sprintf("'%s' in parents and trashed=false", folderID)
	pageToken := ""

	for {
		r, err := srv.Files.List().Q(query).Fields("nextPageToken, files(id, name, mimeType)").PageToken(pageToken).Do()
		if err != nil {
			return nil, err
		}

		for _, file := range r.Files {
			fileInfo := &FileInfo{
				Name:     file.Name,
				Id:       file.Id,
				MimeType: file.MimeType,
			}
			if file.MimeType == "application/vnd.google-apps.folder" {
				subFiles, err := s.listFilesInFolder(srv, file.Id)
				if err != nil {
					return nil, err
				}
				fileInfo.Children = subFiles
			}
			files = append(files, fileInfo)
		}

		if r.NextPageToken == "" {
			break
		}
		pageToken = r.NextPageToken
	}

	return files, nil
}

func (s *GoogleDriveService) downloadFile(srv *drive.Service, fileId string, fileName string) error {
	const op string = "service.GoogleDriveService.downloadFile"

	log := s.log.With(
		sl.String("op", op),
	)

	resp, err := srv.Files.Get(fileId).Download()
	if err != nil {
		log.Error("Unable to retrieve file data", sl.Err(err))
		return err
	}
	defer resp.Body.Close()

	outFile, err := os.Create(fileName)
	if err != nil {
		log.Error("Unable to create file", sl.Err(err))
		return err
	}
	defer outFile.Close()

	_, err = io.Copy(outFile, resp.Body)
	if err != nil {
		log.Error("Unable to write file data to disk", sl.Err(err))
		return err
	}

	log.Info("File downloaded", sl.String("fileName", fileName))
	return nil
}

func (s *GoogleDriveService) getClient() *http.Client {
	if s.config == nil {
		if err := s.initConfig(); err != nil {
			return nil
		}
	}

	tokFile := "token.json"
	tok, err := s.tokenFromFile(tokFile)
	if err != nil {
		fmt.Println("Token not found, please re-authenticate.")
		return nil
	}
	return s.config.Client(context.Background(), tok)
}

func (s *GoogleDriveService) tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	defer f.Close()
	tok := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(tok)
	return tok, err
}

func (s *GoogleDriveService) saveToken(path string, token *oauth2.Token) error {
	const op string = "service.GoogleDriveService.saveToken"

	log := s.log.With(
		sl.String("op", op),
	)

	fmt.Printf("Saving credential file to: %s\n", path)
	f, err := os.Create(path)
	if err != nil {
		log.Error("Unable to cache oauth token", sl.Err(err))
		return err
	}
	defer f.Close()
	return json.NewEncoder(f).Encode(token)
}
