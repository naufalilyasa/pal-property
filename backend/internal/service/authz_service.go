package service

import (
	"fmt"

	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
)

type AuthzService interface {
	EnforceCategoryAction(principal pkgauthz.Principal, action string) error
	EnforceListingAction(principal pkgauthz.Principal, listing *entity.Listing, action string) error
}

type authorizationService struct {
	enforcer *pkgauthz.Service
}

func NewAuthzService(enforcer *pkgauthz.Service) AuthzService {
	return &authorizationService{enforcer: enforcer}
}

func (s *authorizationService) EnforceCategoryAction(principal pkgauthz.Principal, action string) error {
	return s.enforce(pkgauthz.Request{
		Principal: principal,
		Resource:  pkgauthz.ResourceCategory,
		Action:    action,
	})
}

func (s *authorizationService) EnforceListingAction(principal pkgauthz.Principal, listing *entity.Listing, action string) error {
	if listing == nil {
		return domain.ErrNotFound
	}

	ownerID := listing.UserID

	return s.enforce(pkgauthz.Request{
		Principal: principal,
		Resource:  pkgauthz.ResourceListing,
		Action:    action,
		OwnerID:   &ownerID,
	})
}

func (s *authorizationService) enforce(request pkgauthz.Request) error {
	if s == nil || s.enforcer == nil {
		return fmt.Errorf("authz service unavailable")
	}

	allowed, err := s.enforcer.Enforce(request)
	if err != nil {
		return err
	}

	if !allowed {
		return domain.ErrForbidden
	}

	return nil
}
