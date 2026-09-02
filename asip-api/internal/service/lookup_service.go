package service

import (
	"errors"

	"github.com/ArminDashti/as-ip/server/internal/repository"
)

var (
	ErrBadRequest = errors.New("bad request")
	ErrNotFound   = errors.New("not found")
)

type LookupService struct {
	repository *repository.AsRepository
	resolver   dnsResolver
	httpClient rdapHTTPClient
}

func NewLookupService(repository *repository.AsRepository) *LookupService {
	return &LookupService{repository: repository, resolver: stdDNSResolver{}}
}
