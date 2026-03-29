package authz

import (
	_ "embed"
	"fmt"
	"strings"

	"github.com/casbin/casbin/v3"
	"github.com/casbin/casbin/v3/model"
	gormadapter "github.com/casbin/gorm-adapter/v3"
	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	PolicyTableName = "casbin_rule"

	PrincipalContextKey = "authz_principal"
	ServiceContextKey   = "authz_service"

	RoleAdmin = "admin"
	RoleUser  = "user"

	SubjectOwner = "owner"

	ResourceCategory = "category"
	ResourceListing  = "listing"

	ActionCreate          = "create"
	ActionUpdate          = "update"
	ActionDelete          = "delete"
	ActionUploadImage     = "upload_image"
	ActionDeleteImage     = "delete_image"
	ActionSetPrimaryImage = "set_primary_image"
	ActionReorderImages   = "reorder_images"
	ActionUploadVideo     = "upload_video"
	ActionDeleteVideo     = "delete_video"
)

//go:embed model.conf
var modelConfig string

type Principal struct {
	UserID uuid.UUID
	Role   string
}

func (p Principal) Subject() string {
	return p.UserID.String()
}

func NewPrincipal(userID uuid.UUID, role string) (Principal, error) {
	if userID == uuid.Nil {
		return Principal{}, fmt.Errorf("authz: user id is required")
	}

	if strings.TrimSpace(role) == "" {
		return Principal{}, fmt.Errorf("authz: role is required")
	}

	return Principal{UserID: userID, Role: role}, nil
}

type Request struct {
	Principal Principal
	Resource  string
	Action    string
	OwnerID   *uuid.UUID
}

type Service struct {
	enforcer *casbin.SyncedEnforcer
}

func NewService(db *gorm.DB) (*Service, error) {
	if db == nil {
		return nil, fmt.Errorf("authz: db is nil")
	}

	gormadapter.TurnOffAutoMigrate(db)

	adapter, err := gormadapter.NewAdapterByDBUseTableName(db, "", PolicyTableName)
	if err != nil {
		return nil, fmt.Errorf("authz: create adapter: %w", err)
	}

	loadedModel, err := NewModel()
	if err != nil {
		return nil, fmt.Errorf("authz: load model: %w", err)
	}

	enforcer, err := casbin.NewSyncedEnforcer(loadedModel, adapter)
	if err != nil {
		return nil, fmt.Errorf("authz: create enforcer: %w", err)
	}

	if err := enforcer.LoadPolicy(); err != nil {
		return nil, fmt.Errorf("authz: load policy: %w", err)
	}

	return &Service{enforcer: enforcer}, nil
}

func NewModel() (model.Model, error) {
	return model.NewModelFromString(strings.TrimSpace(modelConfig))
}

func NewServiceFromEnforcer(enforcer *casbin.SyncedEnforcer) *Service {
	return &Service{enforcer: enforcer}
}

func (s *Service) Enforcer() *casbin.SyncedEnforcer {
	return s.enforcer
}

func (s *Service) LoadPolicy() error {
	if s == nil || s.enforcer == nil {
		return fmt.Errorf("authz: enforcer is nil")
	}

	return s.enforcer.LoadPolicy()
}

func (s *Service) Enforce(request Request) (bool, error) {
	if s == nil || s.enforcer == nil {
		return false, fmt.Errorf("authz: enforcer is nil")
	}

	owner := ""
	if request.OwnerID != nil {
		owner = request.OwnerID.String()
	}

	return s.enforcer.Enforce(
		request.Principal.Subject(),
		request.Principal.Role,
		request.Resource,
		request.Action,
		owner,
	)
}
