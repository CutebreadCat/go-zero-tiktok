package user_service

import (
	"context"
	"io"

	"go_zero-tiktok/pkg/contract"
	"go_zero-tiktok/pkg/xerr"
)

type ObjectStorage interface {
	DeleteFile(objectKey string) error
	UploadFile(reader io.Reader, objectKey string) (url string, err error)
}

type ProfileService struct {
	userRepo IUserRepo
	storage  ObjectStorage
}

func NewProfileService(userRepo IUserRepo, storage ObjectStorage) *ProfileService {
	return &ProfileService{
		userRepo: userRepo,
		storage:  storage,
	}
}

func (s *ProfileService) GetUserByID(ctx context.Context, userID string) (*types.UserBaseinfo, error) {
	return s.userRepo.GetUserByID(ctx, userID)
}

func (s *ProfileService) UpdatePhoto(ctx context.Context, userID string, file io.Reader) (string, error) {
	objectKey := "user_photos/" + userID + "/" + "profile_photo.jpg"

	userinfo, err := s.userRepo.GetUserByID(ctx, userID)
	if err != nil {
		return "", err
	}

	if userinfo.PhotoURL != "" && userinfo.PhotoURL != "https://example.com/default_photo.jpg" {
		if err := s.storage.DeleteFile(objectKey); err != nil {
			return "", xerr.HandleDaoError(err, "UpdateUserPhoto.DeleteOldPhoto")
		}
	}

	photoURL, err := s.storage.UploadFile(file, objectKey)
	if err != nil {
		return "", xerr.HandleDaoError(err, "UpdateUserPhoto.UploadToOSS")
	}

	if err := s.userRepo.UpdateUserPhotoByID(ctx, userID, photoURL); err != nil {
		return "", xerr.HandleDaoError(err, "UpdateUserPhoto.UpdateUserPhoto")
	}

	return photoURL, nil
}
