package service

import (
	"context"

	"github.com/ArminDashti/as-ip/server/internal/dto"
	"github.com/ArminDashti/as-ip/server/internal/mapper"
	"github.com/ArminDashti/as-ip/server/internal/model"
)

func (s *LookupService) ListCountries(ctx context.Context) (dto.CountryListResponse, error) {
	rows, err := s.repository.ListCountries(ctx)
	if err != nil {
		return dto.CountryListResponse{}, err
	}

	items := make([]dto.CountryItem, 0, len(rows))
	for _, row := range rows {
		items = append(items, dto.CountryItem{
			Code: row.Code,
			Name: row.Name,
		})
	}

	return dto.CountryListResponse{Items: items, Total: len(items)}, nil
}

func (s *LookupService) SearchCountry(ctx context.Context, country string) (dto.AsnListResponse, error) {
	countryCode, err := normalizeCountryCode(country)
	if err != nil {
		return dto.AsnListResponse{}, err
	}

	records, err := s.repository.ListByCountry(ctx, countryCode)
	if err != nil {
		return dto.AsnListResponse{}, err
	}

	return toAsnListResponse(records), nil
}

func toAsnListResponse(records []model.AsRecord) dto.AsnListResponse {
	items := make([]dto.AsnResponse, 0, len(records))
	for _, record := range records {
		items = append(items, mapper.ToAsnResponse(record))
	}

	return dto.AsnListResponse{Items: items, Total: len(items)}
}
