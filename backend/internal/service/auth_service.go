package service

import (
	"errors"
	"strings"

	"github.com/clausify/backend/internal/config"
	"github.com/clausify/backend/internal/dto"
	"github.com/clausify/backend/internal/models"
	"github.com/clausify/backend/internal/repository"
	"github.com/clausify/backend/pkg/utils"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type AuthService struct {
	userRepo *repository.UserRepository
	orgRepo  *repository.OrganizationRepository
	cfg      *config.Config
	logger   *zap.Logger
}

func NewAuthService(
	userRepo *repository.UserRepository,
	orgRepo *repository.OrganizationRepository,
	cfg *config.Config,
	logger *zap.Logger,
) *AuthService {
	return &AuthService{userRepo: userRepo, orgRepo: orgRepo, cfg: cfg, logger: logger}
}

// Register creates a new organization and its first admin user.
func (s *AuthService) Register(req dto.RegisterRequest) (*dto.AuthResponse, error) {
	existing, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if existing != nil {
		return nil, errors.New("email already in use")
	}

	slug := slugify(req.OrgName)
	// Ensure slug uniqueness
	existingOrg, _ := s.orgRepo.FindBySlug(slug)
	if existingOrg != nil {
		slug = slug + "-" + uuid.New().String()[:8]
	}

	org := &models.Organization{
		Name:                req.OrgName,
		Slug:                slug,
		PlanTier:            "starter",
		MaxMonthlyDocuments: 100,
	}
	if err := s.orgRepo.Create(org); err != nil {
		return nil, err
	}

	hash, err := utils.HashPassword(req.Password)
	if err != nil {
		return nil, err
	}

	user := &models.User{
		OrgID:        org.ID,
		Email:        req.Email,
		PasswordHash: hash,
		FullName:     req.FullName,
		Role:         models.RoleAdmin,
		IsActive:     true,
	}
	if err := s.userRepo.Create(user); err != nil {
		return nil, err
	}

	token, err := utils.GenerateToken(s.cfg, user.ID, org.ID, user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     string(user.Role),
			OrgID:    org.ID,
			OrgName:  org.Name,
		},
	}, nil
}

// Login validates credentials and returns a JWT.
func (s *AuthService) Login(req dto.LoginRequest) (*dto.AuthResponse, error) {
	user, err := s.userRepo.FindByEmail(req.Email)
	if err != nil {
		return nil, err
	}
	if user == nil || !utils.CheckPassword(req.Password, user.PasswordHash) {
		return nil, errors.New("invalid email or password")
	}
	if !user.IsActive {
		return nil, errors.New("account is deactivated")
	}

	token, err := utils.GenerateToken(s.cfg, user.ID, user.OrgID, user.Email, string(user.Role))
	if err != nil {
		return nil, err
	}

	return &dto.AuthResponse{
		Token: token,
		User: dto.UserProfile{
			ID:       user.ID,
			Email:    user.Email,
			FullName: user.FullName,
			Role:     string(user.Role),
			OrgID:    user.OrgID,
			OrgName:  user.Organization.Name,
		},
	}, nil
}

func slugify(s string) string {
	s = strings.ToLower(s)
	s = strings.ReplaceAll(s, " ", "-")
	return s
}
