package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/mapper"
	"github.com/ArminDashti/as-ip/server/internal/repository"
)

func (s *LookupService) SearchAs(ctx context.Context, asIdentifier string) (dto.AsnResponse, error) {
	normalized, err := normalizeRequired(asIdentifier, "as")
	if err != nil {
		return dto.AsnResponse{}, err
	}

	record, err := s.repository.FindByAsIdentifier(ctx, normalized)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.AsnResponse{}, fmt.Errorf("no AS data found for %s: %w", normalized, ErrNotFound)
		}
		return dto.AsnResponse{}, err
	}

	return mapper.ToAsnResponse(record), nil
}
