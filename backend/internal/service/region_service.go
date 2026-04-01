package service

import (
	"context"
	"errors"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
)

type RegionOption struct {
	Code string `json:"code"`
	Name string `json:"name"`
}

type RegionSelection struct {
	Province *entity.AdministrativeRegion
	City     *entity.AdministrativeRegion
	District *entity.AdministrativeRegion
	Village  *entity.AdministrativeRegion
}

type RegionLookupService interface {
	ListProvinces(ctx context.Context) ([]RegionOption, error)
	ListCities(ctx context.Context, provinceCode string) ([]RegionOption, error)
	ListDistricts(ctx context.Context, cityCode string) ([]RegionOption, error)
	ListVillages(ctx context.Context, districtCode string) ([]RegionOption, error)
	ResolveHierarchy(ctx context.Context, provinceCode, cityCode, districtCode, villageCode string) (*RegionSelection, error)
}

type regionService struct {
	repo domain.RegionRepository
}

func NewRegionService(repo domain.RegionRepository) RegionLookupService {
	return &regionService{repo: repo}
}

func (s *regionService) ListProvinces(ctx context.Context) ([]RegionOption, error) {
	regions, err := s.repo.ListByLevel(ctx, 1)
	if err != nil {
		return nil, err
	}
	return toRegionOptions(regions), nil
}

func (s *regionService) ListCities(ctx context.Context, provinceCode string) ([]RegionOption, error) {
	if _, err := s.requireRegion(ctx, provinceCode, 1); err != nil {
		return nil, err
	}
	regions, err := s.repo.ListByParent(ctx, provinceCode)
	if err != nil {
		return nil, err
	}
	return toRegionOptions(filterRegionsByLevel(regions, 2)), nil
}

func (s *regionService) ListDistricts(ctx context.Context, cityCode string) ([]RegionOption, error) {
	if _, err := s.requireRegion(ctx, cityCode, 2); err != nil {
		return nil, err
	}
	regions, err := s.repo.ListByParent(ctx, cityCode)
	if err != nil {
		return nil, err
	}
	return toRegionOptions(filterRegionsByLevel(regions, 3)), nil
}

func (s *regionService) ListVillages(ctx context.Context, districtCode string) ([]RegionOption, error) {
	if _, err := s.requireRegion(ctx, districtCode, 3); err != nil {
		return nil, err
	}
	regions, err := s.repo.ListByParent(ctx, districtCode)
	if err != nil {
		return nil, err
	}
	return toRegionOptions(filterRegionsByLevel(regions, 4)), nil
}

func (s *regionService) ResolveHierarchy(ctx context.Context, provinceCode, cityCode, districtCode, villageCode string) (*RegionSelection, error) {
	province, err := s.requireRegion(ctx, provinceCode, 1)
	if err != nil {
		return nil, err
	}
	city, err := s.requireRegion(ctx, cityCode, 2)
	if err != nil {
		return nil, err
	}
	if city.ParentCode == nil || *city.ParentCode != province.Code {
		return nil, domain.ErrInvalidLocation
	}
	district, err := s.requireRegion(ctx, districtCode, 3)
	if err != nil {
		return nil, err
	}
	if district.ParentCode == nil || *district.ParentCode != city.Code {
		return nil, domain.ErrInvalidLocation
	}
	village, err := s.requireRegion(ctx, villageCode, 4)
	if err != nil {
		return nil, err
	}
	if village.ParentCode == nil || *village.ParentCode != district.Code {
		return nil, domain.ErrInvalidLocation
	}

	return &RegionSelection{
		Province: province,
		City:     city,
		District: district,
		Village:  village,
	}, nil
}

func (s *regionService) requireRegion(ctx context.Context, code string, level int) (*entity.AdministrativeRegion, error) {
	if code == "" {
		return nil, domain.ErrInvalidLocation
	}

	region, err := s.repo.FindByCode(ctx, code)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			return nil, domain.ErrInvalidLocation
		}
		return nil, err
	}
	if region.Level != level {
		return nil, domain.ErrInvalidLocation
	}
	return region, nil
}

func toRegionOptions(regions []entity.AdministrativeRegion) []RegionOption {
	options := make([]RegionOption, 0, len(regions))
	for _, region := range regions {
		options = append(options, RegionOption{Code: region.Code, Name: region.Name})
	}
	return options
}

func filterRegionsByLevel(regions []entity.AdministrativeRegion, level int) []entity.AdministrativeRegion {
	filtered := make([]entity.AdministrativeRegion, 0, len(regions))
	for _, region := range regions {
		if region.Level == level {
			filtered = append(filtered, region)
		}
	}
	return filtered
}
