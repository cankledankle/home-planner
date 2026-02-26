package sftpgo

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"time"
)

type Client struct {
	baseURL  string
	username string
	password string
	client   *http.Client
}

type User struct {
	ID                 int64               `json:"id,omitempty"`
	Username           string              `json:"username"`
	Email              string              `json:"email,omitempty"`
	Status             int                 `json:"status"`
	ExpirationDate     int64               `json:"expiration_date,omitempty"`
	Password           string              `json:"password,omitempty"`
	PublicKeys         []string            `json:"public_keys,omitempty"`
	HomeDir            string              `json:"home_dir,omitempty"`
	UID                int                 `json:"uid,omitempty"`
	GID                int                 `json:"gid,omitempty"`
	MaxSessions        int                 `json:"max_sessions,omitempty"`
	QuotaSize          int64               `json:"quota_size,omitempty"`
	QuotaFiles         int                 `json:"quota_files,omitempty"`
	Permissions        map[string][]string `json:"permissions"`
	UsedQuotaSize      int64               `json:"used_quota_size,omitempty"`
	UsedQuotaFiles     int                 `json:"used_quota_files,omitempty"`
	LastQuotaUpdate    int64               `json:"last_quota_update,omitempty"`
	UploadBandwidth    int64               `json:"upload_bandwidth,omitempty"`
	DownloadBandwidth  int64               `json:"download_bandwidth,omitempty"`
	Filters            Filters             `json:"filters,omitempty"`
	Filesystem         Filesystem          `json:"filesystem"`
	CreatedAt          int64               `json:"created_at,omitempty"`
	UpdatedAt          int64               `json:"updated_at,omitempty"`
	LastLogin          int64               `json:"last_login,omitempty"`
	FirstDownload      int64               `json:"first_download,omitempty"`
	FirstUpload        int64               `json:"first_upload,omitempty"`
	TotalDataTransfer  int64               `json:"total_data_transfer,omitempty"`
	DeleteDataTransfer int64               `json:"delete_data_transfer,omitempty"`
	Description        string              `json:"description,omitempty"`
}

type Filters struct {
	AllowedIP          []string `json:"allowed_ip,omitempty"`
	DeniedIP           []string `json:"denied_ip,omitempty"`
	DenyLogin          bool     `json:"deny_login,omitempty"`
	TwoFactorAuth      bool     `json:"two_factor_auth,omitempty"`
	TLSUsername        string   `json:"tls_username,omitempty"`
	Hook               string   `json:"hook,omitempty"`
	DisableFsHooks     bool     `json:"disable_fs_hooks,omitempty"`
	MaxUploadFileSize  int64    `json:"max_upload_file_size,omitempty"`
	MaxShareExpiration int      `json:"max_share_expiration,omitempty"`
}

type Filesystem struct {
	Provider     int           `json:"provider"`
	S3Config     *S3Config     `json:"s3config,omitempty"`
	GCSConfig    *GCSConfig    `json:"gcsconfig,omitempty"`
	AzBlobConfig *AzBlobConfig `json:"azblobconfig,omitempty"`
	CryptConfig  *CryptConfig  `json:"cryptconfig,omitempty"`
	SFTPConfig   *SFTPConfig   `json:"sftpconfig,omitempty"`
}

type S3Config struct {
	Bucket            string `json:"bucket"`
	Region            string `json:"region,omitempty"`
	AccessKey         string `json:"access_key,omitempty"`
	AccessSecret      string `json:"access_secret,omitempty"`
	Endpoint          string `json:"endpoint,omitempty"`
	StorageClass      string `json:"storage_class,omitempty"`
	ACL               string `json:"acl,omitempty"`
	KeyPrefix         string `json:"key_prefix,omitempty"`
	UploadPartSize    int64  `json:"upload_part_size,omitempty"`
	UploadConcurrency int    `json:"upload_concurrency,omitempty"`
	ForcePathStyle    bool   `json:"force_path_style,omitempty"`
}

type GCSConfig struct {
	Bucket               string `json:"bucket"`
	StorageClass         string `json:"storage_class,omitempty"`
	ACL                  string `json:"acl,omitempty"`
	CredentialsFile      string `json:"credentials_file,omitempty"`
	AutomaticCredentials int    `json:"automatic_credentials,omitempty"`
	KeyPrefix            string `json:"key_prefix,omitempty"`
	UploadPartSize       int64  `json:"upload_part_size,omitempty"`
	UploadConcurrency    int    `json:"upload_concurrency,omitempty"`
}

type AzBlobConfig struct {
	Container         string `json:"container"`
	AccountName       string `json:"account_name,omitempty"`
	AccountKey        string `json:"account_key,omitempty"`
	SASUrl            string `json:"sas_url,omitempty"`
	Endpoint          string `json:"endpoint,omitempty"`
	UploadPartSize    int64  `json:"upload_part_size,omitempty"`
	UploadConcurrency int    `json:"upload_concurrency,omitempty"`
	KeyPrefix         string `json:"key_prefix,omitempty"`
	UseEmulator       bool   `json:"use_emulator,omitempty"`
	AccessTier        string `json:"access_tier,omitempty"`
}

type CryptConfig struct {
	Passphrase string `json:"passphrase,omitempty"`
}

type SFTPConfig struct {
	Endpoint                string   `json:"endpoint,omitempty"`
	Username                string   `json:"username,omitempty"`
	Password                string   `json:"password,omitempty"`
	PrivateKey              string   `json:"private_key,omitempty"`
	Fingerprints            []string `json:"fingerprints,omitempty"`
	Prefix                  string   `json:"prefix,omitempty"`
	DisableCouncurrentReads bool     `json:"disable_concurrent_reads,omitempty"`
}

type APIResponse struct {
	Error  string `json:"error,omitempty"`
	Status int    `json:"status,omitempty"`
}

func NewClient() (*Client, error) {
	baseURL := os.Getenv("SFTPGO_API_URL")
	if baseURL == "" {
		return nil, fmt.Errorf("SFTPGO_API_URL not configured")
	}

	username := os.Getenv("SFTPGO_ADMIN_USERNAME")
	if username == "" {
		return nil, fmt.Errorf("SFTPGO_ADMIN_USERNAME not configured")
	}

	password := os.Getenv("SFTPGO_ADMIN_PASSWORD")
	if password == "" {
		return nil, fmt.Errorf("SFTPGO_ADMIN_PASSWORD not configured")
	}

	return &Client{
		baseURL:  baseURL,
		username: username,
		password: password,
		client:   &http.Client{Timeout: 30 * time.Second},
	}, nil
}

func (c *Client) doRequest(ctx context.Context, method, path string, body interface{}) (*http.Response, error) {
	var bodyReader io.Reader
	if body != nil {
		jsonBody, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request body: %w", err)
		}
		bodyReader = bytes.NewReader(jsonBody)
	}

	url := c.baseURL + path
	req, err := http.NewRequestWithContext(ctx, method, url, bodyReader)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(c.username, c.password)

	resp, err := c.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	return resp, nil
}

func (c *Client) CreateUser(ctx context.Context, user *User) error {
	resp, err := c.doRequest(ctx, http.MethodPost, "/api/v2/users", user)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to create SFTPGo user: %s (status: %d)", string(body), resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(user)
}

func (c *Client) GetUser(ctx context.Context, username string) (*User, error) {
	resp, err := c.doRequest(ctx, http.MethodGet, fmt.Sprintf("/api/v2/users/%s", username), nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to get SFTPGo user: %s (status: %d)", string(body), resp.StatusCode)
	}

	var user User
	if err := json.NewDecoder(resp.Body).Decode(&user); err != nil {
		return nil, fmt.Errorf("failed to decode user: %w", err)
	}

	return &user, nil
}

func (c *Client) UpdateUser(ctx context.Context, username string, user *User) error {
	resp, err := c.doRequest(ctx, http.MethodPut, fmt.Sprintf("/api/v2/users/%s", username), user)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to update SFTPGo user: %s (status: %d)", string(body), resp.StatusCode)
	}

	return json.NewDecoder(resp.Body).Decode(user)
}

func (c *Client) DeleteUser(ctx context.Context, username string) error {
	resp, err := c.doRequest(ctx, http.MethodDelete, fmt.Sprintf("/api/v2/users/%s", username), nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil
	}

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("failed to delete SFTPGo user: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

func (c *Client) ListUsers(ctx context.Context, offset, limit int) ([]User, error) {
	url := fmt.Sprintf("/api/v2/users?offset=%d&limit=%d", offset, limit)
	resp, err := c.doRequest(ctx, http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("failed to list SFTPGo users: %s (status: %d)", string(body), resp.StatusCode)
	}

	var users []User
	if err := json.NewDecoder(resp.Body).Decode(&users); err != nil {
		return nil, fmt.Errorf("failed to decode users: %w", err)
	}

	return users, nil
}

func (c *Client) GetServerStatus(ctx context.Context) error {
	resp, err := c.doRequest(ctx, http.MethodGet, "/api/v2/status", nil)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("SFTPGo server not responding: %s (status: %d)", string(body), resp.StatusCode)
	}

	return nil
}

func (c *Client) UserExists(ctx context.Context, username string) (bool, error) {
	user, err := c.GetUser(ctx, username)
	if err != nil {
		return false, err
	}
	return user != nil, nil
}
