package domain

import (
	"context"

	"github.com/hammo/influScope/pkg/models"
)

// AvatarStorage handles unstructured file uploads
type AvatarStorage interface {
	UploadAvatar(ctx context.Context, username string, imageData []byte) (string, error)
}

// EventPublisher handles async message brokering
type EventPublisher interface {
	PublishProfile(ctx context.Context, profile models.Influencer) error
	Close() error
}
