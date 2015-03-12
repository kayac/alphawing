package models

import (
	"fmt"
	"net/http"
	"regexp"
	"strconv"
	"strings"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goauth2/oauth/jwt"
	"code.google.com/p/google-api-go-client/drive/v2"
	"code.google.com/p/google-api-go-client/googleapi"
	"code.google.com/p/google-api-go-client/oauth2/v2"
)

type WebApplicationConfig struct {
	ClientId     string
	ClientSecret string
	CallbackUrl  string
	Scope        []string
}

type ServiceAccountConfig struct {
	ClientEmail string
	PrivateKey  string
	Scope       []string
}

type GoogleService struct {
	AccessToken        string
	Client             *http.Client
	OAuth2Service      *oauth2.Service
	DriveService       *drive.Service
	AboutService       *drive.AboutService
	FilesService       *drive.FilesService
	PermissionsService *drive.PermissionsService
}

type CapacityInfo struct {
	Used               string
	Total              string
	PercentageRemained string
}

func CreateOAuthConfig(config *WebApplicationConfig, tokenCache oauth.Cache) *oauth.Config {
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

func GetServiceAccountToken(config *ServiceAccountConfig) (*oauth.Token, error) {
	token := jwt.NewToken(config.ClientEmail, strings.Join(config.Scope, " "), []byte(config.PrivateKey))

	client := &http.Client{}
	oauthToken, err := token.Assert(client)
	if err != nil {
		return nil, err
	}

	return oauthToken, nil
}

func createOAuthClient(token *oauth.Token) *http.Client {
	transport := &oauth.Transport{
		Token: token,
	}
	return transport.Client()
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

func (s *GoogleService) GetUserInfo() (*oauth2.Userinfoplus, error) {
	return s.OAuth2Service.Userinfo.Get().Do()
}

func (s *GoogleService) GetTokenInfo() (*oauth2.Tokeninfo, error) {
	return s.OAuth2Service.Tokeninfo().Access_token(s.AccessToken).Do()
}

func (s *GoogleService) GetAbout() (*drive.About, error) {
	return s.AboutService.Get().Do()
}

func (s *GoogleService) GetCapacityInfo() (*CapacityInfo, error) {
	about, err := s.GetAbout()
	if err != nil {
		return nil, err
	}

	format := "%.2f"
	divisor := 1000000000
	used := float64(about.QuotaBytesUsed) / float64(divisor)
	total := float64(about.QuotaBytesTotal) / float64(divisor)
	percentageRemained := (total - used) / total * float64(100)

	return &CapacityInfo{
		Used:               fmt.Sprintf(format, used),
		Total:              fmt.Sprintf(format, total),
		PercentageRemained: fmt.Sprintf(format, percentageRemained),
	}, nil
}

func ParseGoogleApiError(apiErr error) (int, string, error) {
	if googleErr, ok := apiErr.(*googleapi.Error); ok {
		return googleErr.Code, googleErr.Message, nil
	}

	reg, err := regexp.Compile(`googleapi: got HTTP response code (\d+) and error reading body: (.+)`)
	if err != nil {
		return 0, "", err
	}
	ret := reg.FindStringSubmatch(apiErr.Error())
	if len(ret) < 3 { // miss match
		return 0, "", apiErr
	}
	code, err := strconv.Atoi(ret[1])
	if err != nil {
		return 0, "", err
	}
	return code, ret[2], nil
}
