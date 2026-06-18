package service

import (
	"database/sql"
	"errors"

	"rare_backend/internal/module/auth/domain"
	"rare_backend/internal/module/auth/repo"
	"rare_backend/internal/pkg/hash"
	"rare_backend/internal/pkg/jwt"
)

type UserService struct {
	repo *repo.UserRepo
}

func NewUserService(r *repo.UserRepo) *UserService {
	return &UserService{repo: r}
}

func (s *UserService) Login(in domain.LoginInput) (*domain.LoginResult, error) {
	u, err := s.repo.FindActiveByPhone(in.Phone)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrInvalidCredentials
		}
		return nil, err
	}
	if !hash.CheckPassword(in.Password, u.PasswordHash) {
		return nil, domain.ErrInvalidCredentials
	}
	if err := s.repo.UpdateLoginInfo(u.ID, u.LoginCount+1); err != nil {
		return nil, err
	}
	token, err := jwt.GenerateToken(u.ID, u.Role)
	if err != nil {
		return nil, err
	}
	return &domain.LoginResult{
		UserID:    u.ID,
		Role:      u.Role,
		Nickname:  u.DisplayName,
		Phone:     in.Phone,
		Avatar:    u.Avatar,
		Token:     token,
		ExpiresIn: int(jwt.TokenExpireDuration().Seconds()),
	}, nil
}

func (s *UserService) Register(in domain.RegisterInput) error {
	count, err := s.repo.CountByPhone(in.Phone)
	if err != nil {
		return err
	}
	if count > 0 {
		return domain.ErrPhoneExists
	}
	hashed, err := hash.HashPassword(in.Password)
	if err != nil {
		return err
	}
	return s.repo.CreateRegister(in.Phone, hashed)
}

func (s *UserService) ListUsers(filter domain.UserListFilter) (*domain.UserListResult, error) {
	users, total, err := s.repo.List(filter)
	if err != nil {
		return nil, err
	}
	items := make([]domain.UserItem, 0, len(users))
	for _, u := range users {
		items = append(items, repo.ToUserItem(u))
	}
	return &domain.UserListResult{
		List:     items,
		Total:    total,
		Page:     filter.Page,
		PageSize: filter.PageSize,
	}, nil
}

func (s *UserService) GetUser(id int64) (*domain.UserItem, error) {
	u, err := s.repo.FindByID(id)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrUserNotFound
		}
		return nil, err
	}
	item := repo.ToUserItem(*u)
	return &item, nil
}

func (s *UserService) UpdateUser(id int64, in domain.UpdateUserInput) error {
	ok, err := s.repo.ExistsByID(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrUserNotFound
	}
	fields := map[string]interface{}{}
	if in.DisplayName != nil {
		fields["display_name"] = *in.DisplayName
	}
	if in.Avatar != nil {
		fields["avatar"] = *in.Avatar
	}
	if in.Role != nil {
		if *in.Role != 1 && *in.Role != 2 && *in.Role != 9 {
			return domain.ErrInvalidRole
		}
		fields["role"] = *in.Role
	}
	if in.Status != nil {
		if *in.Status != 0 && *in.Status != 1 {
			return domain.ErrInvalidStatus
		}
		fields["status"] = *in.Status
	}
	return s.repo.UpdateDynamic(id, fields)
}

func (s *UserService) DeleteUser(id int64) error {
	ok, err := s.repo.ExistsActiveByID(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrUserDisabled
	}
	return s.repo.SoftDelete(id)
}

func (s *UserService) UpdateUserRole(id int64, role int) error {
	if role != 1 && role != 2 && role != 9 {
		return domain.ErrInvalidRole
	}
	ok, err := s.repo.ExistsByID(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrUserNotFound
	}
	return s.repo.UpdateRole(id, role)
}

func (s *UserService) ResetPassword(id int64, newPassword string) error {
	ok, err := s.repo.ExistsByID(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrUserNotFound
	}
	hashed, err := hash.HashPassword(newPassword)
	if err != nil {
		return err
	}
	return s.repo.UpdatePassword(id, hashed)
}

func (s *UserService) CreateUser(in domain.CreateUserInput) (int64, error) {
	count, err := s.repo.CountByPhone(in.Phone)
	if err != nil {
		return 0, err
	}
	if count > 0 {
		return 0, domain.ErrPhoneExists
	}
	displayName := in.DisplayName
	if displayName == "" {
		phone := in.Phone
		if len(phone) >= 4 {
			displayName = "用户" + phone[len(phone)-4:]
		} else {
			displayName = "用户"
		}
	}
	in.DisplayName = displayName
	hashed, err := hash.HashPassword(in.Password)
	if err != nil {
		return 0, err
	}
	return s.repo.CreateAdmin(in, hashed)
}

// internal/module/auth/service/user_service.go
func (s *UserService) UpdateAvatar(id int64, avatarURL string) error {
	ok, err := s.repo.ExistsByID(id)
	if err != nil {
		return err
	}
	if !ok {
		return domain.ErrUserNotFound
	}
	return s.repo.UpdateAvatar(id, avatarURL)
}

func (s *UserService) DeactivateAccount(userID int64) error {
	u, err := s.repo.FindByID(userID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return domain.ErrUserNotFound
		}
		return err
	}
	if u.Status != 1 {
		return domain.ErrAccountDeactivated
	}
	if u.Role == jwt.RoleAdmin {
		return domain.ErrCannotDeactivateAdmin
	}
	return s.repo.DeactivateAccount(userID)
}
