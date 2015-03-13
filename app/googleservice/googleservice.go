package googleservice

import (
	"errors"
	"net/http"
	"strings"

	"google.golang.org/api/oauth2/v1"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goauth2/oauth/jwt"
	"code.google.com/p/google-api-go-client/drive/v2"
)

type GoogleService struct {
	AccessToken        string
	Client             *http.Client
	OAuth2Service      *oauth2.Service
	DriveService       *drive.Service
	AboutService       *drive.AboutService
	FilesService       *drive.FilesService
	PermissionsService *drive.PermissionsService
}

type Config struct {
	ClientId     string
	CliendSecret string
	CallbackUrl  string
	Scope        []string
}

func createOAuthClient(token *oauth.Token) *http.Client {
	transport := &oauth.Transport{
		Token: token,
	}
	return transport.Client()
}

func (c *Config) NewOAuthConfig(tokenCache oauth.Cache) *oauth.Config {
	return &oauth.Config{
		ClientId:     config.ClientId,
		ClientSecret: config.ClientSecret,
		AuthURL:      "https://accounts.google.com/o/oauth2/auth",
		TokenURL:     "https://accounts.google.com/o/oauth2/token",
		RedirectURL:  config.CallbackUrl,
		Scope:        strings.Join(config.Scope, " "),
		TokenCache:   tokenCache,
	}
}

type ServiceAccountConfig struct {
	ClientEmail string
	PrivateKey  string
}

func GetServiceAccountToken(config *ServiceAccountConfig) (*oauth.Token, error) {
	token := jwt.NewToken(config.ClientEmail, strings.Join([]string{drive.DriveScope}, " "), []byte(config.PrivateKey))

	client := &http.Client{}
	oauthToken, err := token.Assert(client)
	if err != nil {
		return nil, err
	}

	return oauthToken, nil
}

func NewGoogleService(token *oauth.Token) (*GoogleService, error) {
	client := createOAuthClient(token)

	oauth2Service, err := oauth2.New(client)
	if err != nil {
		return nil, err
	}

	driveService, err := drive.New(client)
	if err != nil {
		return nil, err
	}

	return &GoogleService{
		AccessToken:        token.AccessToken,
		Client:             client,
		OAuth2Service:      oauth2Service,
		DriveService:       driveService,
		AboutService:       drive.NewAboutService(driveService),
		FilesService:       drive.NewFilesService(driveService),
		PermissionsService: drive.NewPermissionsService(driveService),
	}, nil
}

func (s *GoogleService) CreateFolder(folderName string) (*drive.File, error) {
	driveFolder := &drive.File{
		Title:    folderName,
		MimeType: "application/vnd.google-apps.folder",
	}
	return s.FilesService.Insert(driveFolder).Do()
}

func (s *GoogleService) CreateUserPermission(email string, role string) *drive.Permission {
	return &drive.Permission{
		Role:  role,
		Type:  "user",
		Value: email,
	}
}

func (s *GoogleService) InsertPermission(fileId string, permission *drive.Permission) (*drive.Permission, error) {
	return s.PermissionsService.Insert(fileId, permission).Do()
}

func (s *GoogleService) GetPermissionList(fileId string) (*drive.PermissionList, error) {
	return s.PermissionsService.List(fileId).Do()
}

func (s *GoogleService) DeletePermission(fileId string, permissionId string) error {
	return s.PermissionsService.Delete(fileId, permissionId).Do()
}

func (s *GoogleService) GetPermissionId(fileId string, email string) (string, error) {
	permList, err := s.GetPermissionList(fileId)
	if err != nil {
		return "", err
	}

	var permId string
	for _, perm := range permList.Items {
		if perm.EmailAddress == email {
			permId = perm.Id
			break
		}
	}

	if !permId {
		return "", errors.New("Not found email in folder")
	}

	return permId, nil
}
