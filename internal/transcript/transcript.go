package transcript

import "context"

// Service describes contract for transcript providers.
type Service interface {
	Transcribe(ctx context.Context, audioPath string) (string, error)
}

// DummyService returns placeholder transcript text.
type DummyService struct{}

func NewDummyService() *DummyService {
	return &DummyService{}
}

func (d *DummyService) Transcribe(ctx context.Context, audioPath string) (string, error) {
	return "Transcription is not implemented yet.", nil
}
