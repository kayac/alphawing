package models

import (
	cryptorand "crypto/rand"
	"encoding/binary"
	mathrand "math/rand"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/goauth2/oauth/jwt"
	"code.google.com/p/google-api-go-client/drive/v2"
	"code.google.com/p/google-api-go-client/googleapi"
	"code.google.com/p/google-api-go-client/oauth2/v2"
	"code.google.com/p/google-api-go-client/storage/v1"
	gcs "google.golang.org/cloud/storage"
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

type GoogleStorageConfig struct {
	ProjectId          string
	BucketPrefix       string
	BucketLocation     string
	BucketStorageClass string
}

type GoogleService struct {
	AccessToken          string
	Client               *http.Client
	OAuth2Service        *oauth2.Service
	DriveService         *drive.Service
	AboutService         *drive.AboutService
	FilesService         *drive.FilesService
	PermissionsService   *drive.PermissionsService
	StorageService       *storage.Service
	Config               *GoogleStorageConfig
	ServiceAccountConfig *ServiceAccountConfig
}

type CapacityInfo struct {
	Used               string
	Total              string
	PercentageRemained string
}

var rand *mathrand.Rand

func init() {
	var n int64
	binary.Read(cryptorand.Reader, binary.LittleEndian, &n)
	rand = mathrand.New(mathrand.NewSource(n))
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

func NewGoogleService(token *oauth.Token, config *GoogleStorageConfig, serviceAccountConfig *ServiceAccountConfig) (*GoogleService, error) {
	client := createOAuthClient(token)

	oauth2Service, err := oauth2.New(client)
	if err != nil {
		return nil, err
	}

	driveService, err := drive.New(client)
	if err != nil {
		return nil, err
	}

	storageService, err := storage.New(client)
	if err != nil {
		return nil, err
	}

	return &GoogleService{
		AccessToken:          token.AccessToken,
		Client:               client,
		OAuth2Service:        oauth2Service,
		DriveService:         driveService,
		AboutService:         drive.NewAboutService(driveService),
		FilesService:         drive.NewFilesService(driveService),
		PermissionsService:   drive.NewPermissionsService(driveService),
		StorageService:       storageService,
		Config:               config,
		ServiceAccountConfig: serviceAccountConfig,
	}, nil
}

func (s *GoogleService) GetUserInfo() (*oauth2.Userinfoplus, error) {
	return s.OAuth2Service.Userinfo.Get().Do()
}

func (s *GoogleService) GetTokenInfo() (*oauth2.Tokeninfo, error) {
	return s.OAuth2Service.Tokeninfo().Access_token(s.AccessToken).Do()
}

func (s *GoogleService) CreateBucket() (*storage.Bucket, error) {
	bucketName := s.Config.BucketPrefix + strconv.FormatInt(rand.Int63(), 16)
	bucket := &storage.Bucket{
		Name:         bucketName,
		Location:     s.Config.BucketLocation,
		StorageClass: s.Config.BucketStorageClass,
	}
	return s.StorageService.Buckets.Insert(s.Config.ProjectId, bucket).Do()
}

func (s *GoogleService) InsertFile(file *os.File, filename, bucketName string) (*storage.Object, error) {
	contentType := "application/octet-stream"
	if strings.HasSuffix(filename, ".apk") {
		contentType = "application/vnd.android.package-archive"
	}
	obj := &storage.Object{
		Name:        filename,
		ContentType: contentType,
	}
	return s.StorageService.Objects.Insert(bucketName, obj).Media(file).Do()
}

func (s *GoogleService) GetObject(objId string) (*storage.Object, error) {
	array := strings.SplitN(objId, "/", 2)
	bucketName := array[0]
	objectName := array[1]
	return s.StorageService.Objects.Get(bucketName, objectName).Do()
}

func (s *GoogleService) GetBucket(bucketName string) (*storage.Bucket, error) {
	return s.StorageService.Buckets.Get(bucketName).Do()
}

func (s *GoogleService) DownloadFile(objId string) (*http.Response, *storage.Object, error) {
	file, err := s.GetObject(objId)
	if err != nil {
		return nil, nil, err
	}
	resp, err := s.Client.Get(file.MediaLink)
	if err != nil {
		return nil, nil, err
	}
	return resp, file, nil
}

func (s *GoogleService) GetDownloadURL(objId string) (string, error) {
	array := strings.SplitN(objId, "/", 2)
	bucketName := array[0]
	objectName := array[1]

	opts := &gcs.SignedURLOptions{
		GoogleAccessID: s.ServiceAccountConfig.ClientEmail,
		PrivateKey:     []byte(s.ServiceAccountConfig.PrivateKey),
		Method:         "GET",
		Expires:        time.Now().Add(time.Second * 120),
	}

	return gcs.SignedURL(bucketName, objectName, opts)
}

func (s *GoogleService) GetSharedFileList() ([]string, error) {
	buckets, err := s.StorageService.Buckets.List(s.Config.ProjectId).Do()
	if err != nil {
		return nil, err
	}

	var ids []string
	for _, b := range buckets.Items {
		ids = append(ids, b.Name)
	}
	return ids, nil
}

func (s *GoogleService) DeleteObject(objId string) error {
	array := strings.SplitN(objId, "/", 2)
	bucketName := array[0]
	objectName := array[1]
	return s.StorageService.Objects.Delete(bucketName, objectName).Do()
}

func (s *GoogleService) DeleteBucket(bucketName string) error {
	allObjects := []string{}
	pageToken := ""
	for {
		call := s.StorageService.Objects.List(bucketName).MaxResults(1)
		if pageToken != "" {
			call = call.PageToken(pageToken)
		}
		result, err := call.Do()
		if err != nil {
			return err
		}
		for _, o := range result.Items {
			allObjects = append(allObjects, o.Name)
		}

		pageToken = result.NextPageToken
		if pageToken == "" {
			break
		}
	}

	for _, objectName := range allObjects {
		err := s.StorageService.Objects.Delete(bucketName, objectName).Do()
		if err != nil {
			return err
		}
	}
	return s.StorageService.Buckets.Delete(bucketName).Do()
}

func (s *GoogleService) CreateUserPermission(email string, role string) *storage.BucketAccessControl {
	return &storage.BucketAccessControl{
		Role:   role,
		Entity: email,
	}
}

func (s *GoogleService) InsertPermission(bucketName string, permission *storage.BucketAccessControl) (*storage.BucketAccessControl, error) {
	ret, err := s.StorageService.BucketAccessControls.Insert(bucketName, &storage.BucketAccessControl{
		Role:   permission.Role,
		Entity: "group-" + permission.Entity,
	}).Do()
	if err == nil {
		return ret, nil
	}
	return s.StorageService.BucketAccessControls.Insert(bucketName, &storage.BucketAccessControl{
		Role:   permission.Role,
		Entity: "user-" + permission.Entity,
	}).Do()
}

func (s *GoogleService) DeletePermission(bucketName string, permissionId string) error {
	return s.StorageService.BucketAccessControls.Delete(bucketName, permissionId).Do()
}

func ParseGoogleApiError(apiErr error) (int, string, error) {
	if googleErr, ok := apiErr.(*googleapi.Error); ok {
		return googleErr.Code, googleErr.Errors[0].Reason, nil
	}
	return 0, "", apiErr
}
