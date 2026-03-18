package authz

import (
	"testing"

	"github.com/casbin/casbin/v3"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewModel_LoadsEmbeddedModel(t *testing.T) {
	loadedModel, err := NewModel()

	require.NoError(t, err)
	assert.NotNil(t, loadedModel)
	assert.NotNil(t, loadedModel["r"])
	assert.NotNil(t, loadedModel["p"])
	assert.NotNil(t, loadedModel["m"])
}

func TestService_Enforce_AdminCategoryCreate(t *testing.T) {
	loadedModel, err := NewModel()
	require.NoError(t, err)

	enforcer, err := casbin.NewSyncedEnforcer(loadedModel)
	require.NoError(t, err)
	_, err = enforcer.AddPolicy(RoleAdmin, ResourceCategory, ActionCreate)
	require.NoError(t, err)

	service := NewServiceFromEnforcer(enforcer)
	allowed, err := service.Enforce(Request{
		Principal: Principal{UserID: uuid.New(), Role: RoleAdmin},
		Resource:  ResourceCategory,
		Action:    ActionCreate,
	})

	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestService_Enforce_OwnerListingUpdate(t *testing.T) {
	loadedModel, err := NewModel()
	require.NoError(t, err)

	enforcer, err := casbin.NewSyncedEnforcer(loadedModel)
	require.NoError(t, err)
	_, err = enforcer.AddPolicy(SubjectOwner, ResourceListing, ActionUpdate)
	require.NoError(t, err)

	ownerID := uuid.New()
	service := NewServiceFromEnforcer(enforcer)
	allowed, err := service.Enforce(Request{
		Principal: Principal{UserID: ownerID, Role: RoleUser},
		Resource:  ResourceListing,
		Action:    ActionUpdate,
		OwnerID:   &ownerID,
	})

	require.NoError(t, err)
	assert.True(t, allowed)
}

func TestService_Enforce_DefaultDeny(t *testing.T) {
	loadedModel, err := NewModel()
	require.NoError(t, err)

	enforcer, err := casbin.NewSyncedEnforcer(loadedModel)
	require.NoError(t, err)

	service := NewServiceFromEnforcer(enforcer)
	allowed, err := service.Enforce(Request{
		Principal: Principal{UserID: uuid.New(), Role: RoleUser},
		Resource:  ResourceCategory,
		Action:    ActionCreate,
	})

	require.NoError(t, err)
	assert.False(t, allowed)
}
