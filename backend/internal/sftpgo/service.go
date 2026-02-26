package sftpgo

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"os"
	"strings"
)

type Service struct {
	client *Client
}

type UserCredentials struct {
	Username   string `json:"username"`
	Password   string `json:"password"`
	Host       string `json:"host"`
	Port       int    `json:"port"`
	Permission string `json:"permission"`
}

type CreateUserRequest struct {
	UserID     string `json:"user_id"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Permission string `json:"permission"` // "read" or "readwrite"
}

type UpdatePermissionRequest struct {
	Permission string `json:"permission"` // "read" or "readwrite"
}

func NewService() (*Service, error) {
	client, err := NewClient()
	if err != nil {
		return nil, err
	}
	return &Service{client: client}, nil
}

func (s *Service) IsConfigured() bool {
	return s.client != nil
}

func generatePassword() (string, error) {
	bytes := make([]byte, 16)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func sanitizeUsername(email string) string {
	// Remove special characters and create a valid SFTPGo username
	username := strings.ToLower(email)
	username = strings.ReplaceAll(username, "@", "_")
	username = strings.ReplaceAll(username, ".", "_")
	username = strings.ReplaceAll(username, "+", "_")
	// Remove any other non-alphanumeric characters
	var result strings.Builder
	for _, r := range username {
		if (r >= 'a' && r <= 'z') || (r >= '0' && r <= '9') || r == '_' {
			result.WriteRune(r)
		}
	}
	return result.String()
}

func getR2Config(userPrefix string) *S3Config {
	accountID := os.Getenv("R2_ACCOUNT_ID")
	accessKey := os.Getenv("R2_ACCESS_KEY_ID")
	secretKey := os.Getenv("R2_SECRET_ACCESS_KEY")
	bucketName := os.Getenv("R2_BUCKET_NAME")

	if accountID == "" || accessKey == "" || secretKey == "" || bucketName == "" {
		return nil
	}

	return &S3Config{
		Bucket:         bucketName,
		Region:         "auto",
		AccessKey:      accessKey,
		AccessSecret:   secretKey,
		Endpoint:       fmt.Sprintf("https://%s.r2.cloudflarestorage.com", accountID),
		ForcePathStyle: true,
		KeyPrefix:      userPrefix,
	}
}

func getPermissions(permission string) map[string][]string {
	basePerms := []string{"list", "download"}
	if permission == "readwrite" {
		basePerms = append(basePerms, "upload", "delete", "rename", "create_dirs")
	}

	return map[string][]string{
		"/": basePerms,
	}
}

func (s *Service) CreateUser(ctx context.Context, req CreateUserRequest) (*UserCredentials, error) {
	if s.client == nil {
		return nil, fmt.Errorf("SFTPGo service not configured")
	}

	password, err := generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	username := sanitizeUsername(req.Email)

	// Check if user already exists
	exists, err := s.client.UserExists(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to check user existence: %w", err)
	}
	if exists {
		// Generate a unique username by appending a number
		baseUsername := username
		for i := 1; i < 100; i++ {
			username = fmt.Sprintf("%s_%d", baseUsername, i)
			exists, err = s.client.UserExists(ctx, username)
			if err != nil {
				return nil, fmt.Errorf("failed to check user existence: %w", err)
			}
			if !exists {
				break
			}
		}
		if exists {
			return nil, fmt.Errorf("could not generate unique username")
		}
	}

	r2Config := getR2Config(fmt.Sprintf("users/%s/", username))
	if r2Config == nil {
		return nil, fmt.Errorf("R2 not configured")
	}

	user := &User{
		Username:    username,
		Email:       req.Email,
		Status:      1, // Enabled
		Password:    password,
		Description: fmt.Sprintf("Home Planner user: %s (%s)", req.Name, req.UserID),
		Permissions: getPermissions(req.Permission),
		Filesystem: Filesystem{
			Provider: 1, // S3
			S3Config: r2Config,
		},
	}

	if err := s.client.CreateUser(ctx, user); err != nil {
		return nil, fmt.Errorf("failed to create SFTPGo user: %w", err)
	}

	// Get the SFTP host and port from environment
	sftpHost := os.Getenv("SFTPGO_HOST")
	if sftpHost == "" {
		sftpHost = "localhost"
	}

	return &UserCredentials{
		Username:   username,
		Password:   password,
		Host:       sftpHost,
		Port:       2022,
		Permission: req.Permission,
	}, nil
}

func (s *Service) DeleteUser(ctx context.Context, email string) error {
	if s.client == nil {
		return fmt.Errorf("SFTPGo service not configured")
	}

	username := sanitizeUsername(email)

	// Get all users and find by username pattern (in case we appended numbers)
	users, err := s.client.ListUsers(ctx, 0, 1000)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range users {
		if user.Username == username || strings.HasPrefix(user.Username, username+"_") {
			if err := s.client.DeleteUser(ctx, user.Username); err != nil {
				return fmt.Errorf("failed to delete SFTPGo user %s: %w", user.Username, err)
			}
		}
	}

	return nil
}

func (s *Service) RotateCredentials(ctx context.Context, email string) (*UserCredentials, error) {
	if s.client == nil {
		return nil, fmt.Errorf("SFTPGo service not configured")
	}

	username := sanitizeUsername(email)

	// Find the user
	user, err := s.client.GetUser(ctx, username)
	if err != nil {
		return nil, fmt.Errorf("failed to get SFTPGo user: %w", err)
	}
	if user == nil {
		// Try with numbered suffixes
		users, err := s.client.ListUsers(ctx, 0, 1000)
		if err != nil {
			return nil, fmt.Errorf("failed to list users: %w", err)
		}
		for _, u := range users {
			if u.Username == username || strings.HasPrefix(u.Username, username+"_") {
				user = &u
				username = u.Username
				break
			}
		}
	}
	if user == nil {
		return nil, fmt.Errorf("SFTPGo user not found")
	}

	password, err := generatePassword()
	if err != nil {
		return nil, fmt.Errorf("failed to generate password: %w", err)
	}

	user.Password = password
	if err := s.client.UpdateUser(ctx, username, user); err != nil {
		return nil, fmt.Errorf("failed to update SFTPGo user: %w", err)
	}

	sftpHost := os.Getenv("SFTPGO_HOST")
	if sftpHost == "" {
		sftpHost = "localhost"
	}

	return &UserCredentials{
		Username:   username,
		Password:   password,
		Host:       sftpHost,
		Port:       2022,
		Permission: getPermissionLevel(user.Permissions),
	}, nil
}

func (s *Service) RevokeAccess(ctx context.Context, email string) error {
	if s.client == nil {
		return fmt.Errorf("SFTPGo service not configured")
	}

	username := sanitizeUsername(email)

	// Find and disable the user
	users, err := s.client.ListUsers(ctx, 0, 1000)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range users {
		if user.Username == username || strings.HasPrefix(user.Username, username+"_") {
			user.Status = 0 // Disabled
			if err := s.client.UpdateUser(ctx, user.Username, &user); err != nil {
				return fmt.Errorf("failed to disable SFTPGo user %s: %w", user.Username, err)
			}
		}
	}

	return nil
}

func (s *Service) UpdatePermission(ctx context.Context, email string, req UpdatePermissionRequest) error {
	if s.client == nil {
		return fmt.Errorf("SFTPGo service not configured")
	}

	username := sanitizeUsername(email)

	// Find the user
	users, err := s.client.ListUsers(ctx, 0, 1000)
	if err != nil {
		return fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range users {
		if user.Username == username || strings.HasPrefix(user.Username, username+"_") {
			user.Permissions = getPermissions(req.Permission)
			if err := s.client.UpdateUser(ctx, user.Username, &user); err != nil {
				return fmt.Errorf("failed to update SFTPGo user permissions: %w", err)
			}
			return nil
		}
	}

	return fmt.Errorf("SFTPGo user not found")
}

func (s *Service) GetCredentials(ctx context.Context, email string) (*UserCredentials, error) {
	if s.client == nil {
		return nil, fmt.Errorf("SFTPGo service not configured")
	}

	username := sanitizeUsername(email)

	// Find the user
	users, err := s.client.ListUsers(ctx, 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	for _, user := range users {
		if user.Username == username || strings.HasPrefix(user.Username, username+"_") {
			sftpHost := os.Getenv("SFTPGO_HOST")
			if sftpHost == "" {
				sftpHost = "localhost"
			}

			return &UserCredentials{
				Username:   user.Username,
				Password:   "", // Don't return the password
				Host:       sftpHost,
				Port:       2022,
				Permission: getPermissionLevel(user.Permissions),
			}, nil
		}
	}

	return nil, nil // User doesn't have SFTP credentials yet
}

func getPermissionLevel(perms map[string][]string) string {
	rootPerms, ok := perms["/"]
	if !ok {
		return "read"
	}

	for _, p := range rootPerms {
		if p == "upload" || p == "delete" {
			return "readwrite"
		}
	}

	return "read"
}

func (s *Service) ListAllUsers(ctx context.Context) ([]UserCredentials, error) {
	if s.client == nil {
		return nil, fmt.Errorf("SFTPGo service not configured")
	}

	users, err := s.client.ListUsers(ctx, 0, 1000)
	if err != nil {
		return nil, fmt.Errorf("failed to list users: %w", err)
	}

	sftpHost := os.Getenv("SFTPGO_HOST")
	if sftpHost == "" {
		sftpHost = "localhost"
	}

	credentials := make([]UserCredentials, 0, len(users))
	for _, user := range users {
		// Only include users created by Home Planner (they have descriptions)
		if strings.Contains(user.Description, "Home Planner user:") {
			credentials = append(credentials, UserCredentials{
				Username:   user.Username,
				Host:       sftpHost,
				Port:       2022,
				Permission: getPermissionLevel(user.Permissions),
			})
		}
	}

	return credentials, nil
}

func (s *Service) CheckHealth(ctx context.Context) error {
	if s.client == nil {
		return fmt.Errorf("SFTPGo service not configured")
	}
	return s.client.GetServerStatus(ctx)
}
