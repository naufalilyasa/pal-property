package service_test

import (
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/naufalilyasa/pal-property-backend/internal/domain"
	"github.com/naufalilyasa/pal-property-backend/internal/domain/entity"
	"github.com/naufalilyasa/pal-property-backend/internal/service"
	pkgauthz "github.com/naufalilyasa/pal-property-backend/pkg/authz"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestAuthzService(t *testing.T) service.AuthzService {
	t.Helper()

	loadedModel, err := pkgauthz.NewModel()
	require.NoError(t, err)

	enforcer, err := casbin.NewSyncedEnforcer(loadedModel)
	require.NoError(t, err)

	policies := [][]string{
		{pkgauthz.RoleAdmin, pkgauthz.ResourceCategory, pkgauthz.ActionCreate},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceCategory, pkgauthz.ActionUpdate},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceCategory, pkgauthz.ActionDelete},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceListing, pkgauthz.ActionUpdate},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceListing, pkgauthz.ActionDelete},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceListing, pkgauthz.ActionUploadImage},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceListing, pkgauthz.ActionDeleteImage},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceListing, pkgauthz.ActionSetPrimaryImage},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceListing, pkgauthz.ActionReorderImages},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceListing, pkgauthz.ActionUploadVideo},
		{pkgauthz.RoleAdmin, pkgauthz.ResourceListing, pkgauthz.ActionDeleteVideo},
		{pkgauthz.SubjectOwner, pkgauthz.ResourceListing, pkgauthz.ActionUpdate},
		{pkgauthz.SubjectOwner, pkgauthz.ResourceListing, pkgauthz.ActionDelete},
		{pkgauthz.SubjectOwner, pkgauthz.ResourceListing, pkgauthz.ActionUploadImage},
		{pkgauthz.SubjectOwner, pkgauthz.ResourceListing, pkgauthz.ActionDeleteImage},
		{pkgauthz.SubjectOwner, pkgauthz.ResourceListing, pkgauthz.ActionSetPrimaryImage},
		{pkgauthz.SubjectOwner, pkgauthz.ResourceListing, pkgauthz.ActionReorderImages},
		{pkgauthz.SubjectOwner, pkgauthz.ResourceListing, pkgauthz.ActionUploadVideo},
		{pkgauthz.SubjectOwner, pkgauthz.ResourceListing, pkgauthz.ActionDeleteVideo},
	}

	for _, policy := range policies {
		_, err = enforcer.AddPolicy(policy)
		require.NoError(t, err)
	}

	return service.NewAuthzService(pkgauthz.NewServiceFromEnforcer(enforcer))
}

func TestAuthzService_EnforceCategoryCreate_AdminAllowed(t *testing.T) {
	authzService := newTestAuthzService(t)
	principal := pkgauthz.Principal{UserID: uuid.New(), Role: pkgauthz.RoleAdmin}

	err := authzService.EnforceCategoryAction(principal, pkgauthz.ActionCreate)

	assert.NoError(t, err)
}

func TestAuthzService_EnforceCategoryCreate_UserForbidden(t *testing.T) {
	authzService := newTestAuthzService(t)
	principal := pkgauthz.Principal{UserID: uuid.New(), Role: pkgauthz.RoleUser}

	err := authzService.EnforceCategoryAction(principal, pkgauthz.ActionCreate)

	assert.ErrorIs(t, err, domain.ErrForbidden)
}

func TestAuthzService_EnforceListingUpdate_OwnerAllowed(t *testing.T) {
	authzService := newTestAuthzService(t)
	ownerID := uuid.New()
	principal := pkgauthz.Principal{UserID: ownerID, Role: pkgauthz.RoleUser}
	listing := &entity.Listing{UserID: ownerID}

	err := authzService.EnforceListingAction(principal, listing, pkgauthz.ActionUpdate)

	assert.NoError(t, err)
}

func TestAuthzService_EnforceListingUpdate_AdminAllowed(t *testing.T) {
	authzService := newTestAuthzService(t)
	principal := pkgauthz.Principal{UserID: uuid.New(), Role: pkgauthz.RoleAdmin}
	listing := &entity.Listing{UserID: uuid.New()}

	err := authzService.EnforceListingAction(principal, listing, pkgauthz.ActionUpdate)

	assert.NoError(t, err)
}

func TestAuthzService_EnforceListingUpdate_NonOwnerForbidden(t *testing.T) {
	authzService := newTestAuthzService(t)
	principal := pkgauthz.Principal{UserID: uuid.New(), Role: pkgauthz.RoleUser}
	listing := &entity.Listing{UserID: uuid.New()}

	err := authzService.EnforceListingAction(principal, listing, pkgauthz.ActionUpdate)

	assert.ErrorIs(t, err, domain.ErrForbidden)
}
