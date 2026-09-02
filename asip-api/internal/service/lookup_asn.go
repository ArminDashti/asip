package service

import (
	"context"
	"errors"
	"fmt"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/mapper"
	"github.com/ArminDashti/as-ip/server/internal/repository"
)

func (s *LookupService) SearchAsn(ctx context.Context, asnIdentifier string) (dto.AsnResponse, error) {
	normalized, err := normalizeRequired(asnIdentifier, "asn")
	if err != nil {
		return dto.AsnResponse{}, err
	}

	record, err := s.repository.FindByAsnIdentifier(ctx, normalized)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.AsnResponse{}, fmt.Errorf("no ASN data found for %s: %w", normalized, ErrNotFound)
		}
		return dto.AsnResponse{}, err
	}

	return mapper.ToAsnResponse(record), nil
}

func (s *LookupService) MapAsnToAs(ctx context.Context, asnIdentifier string) (dto.AsMappingResponse, error) {
	normalized, err := normalizeRequired(asnIdentifier, "asn")
	if err != nil {
		return dto.AsMappingResponse{}, err
	}

	record, err := s.repository.FindByAsnIdentifier(ctx, normalized)
	if err != nil {
		if errors.Is(err, repository.ErrNotFound) {
			return dto.AsMappingResponse{}, fmt.Errorf("no AS mapping found for ASN %s: %w", normalized, ErrNotFound)
		}
		return dto.AsMappingResponse{}, err
	}

	return dto.AsMappingResponse{
		Asn: record.AsnNumber,
		As:  mapper.ToAsSummary(record),
	}, nil
}

func (s *LookupService) ListAsns(ctx context.Context) (dto.AsnCatalogResponse, error) {
	rows, err := s.repository.ListAsns(ctx)
	if err != nil {
		return dto.AsnCatalogResponse{}, err
	}

	items := make([]dto.AsnItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.AsnItem{
			Asn:    row.Number,
			Handle: row.Name,
		})
	}

	return dto.AsnCatalogResponse{Items: items, Total: len(items)}, nil
}
