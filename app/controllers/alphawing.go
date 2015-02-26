package controllers

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"

	"code.google.com/p/go-uuid/uuid"
	"code.google.com/p/goauth2/oauth"
	"code.google.com/p/google-api-go-client/drive/v2"
	"code.google.com/p/google-api-go-client/oauth2/v2"

	"github.com/kayac/alphawing/app/models"
	"github.com/kayac/alphawing/app/routes"

	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
)

type AlphaWingController struct {
	GorpController
	LoginUserId   int
	GoogleService *models.GoogleService
	OAuthConfig   *oauth.Config
}

const LoginSessionKey = "LoginSessionKey"
const OAuthSessionKey = "OAuthSessionKey"

func (c AlphaWingController) Index() revel.Result {
	if !c.isLogin() {
		return c.Render()
	}

	s, err := c.userGoogleService()
	if err != nil {
		panic(err)
	}

	fileList, err := s.GetSharedFileList(Conf.ServiceAccountClientEmail)
	if err != nil {
		panic(err)
	}

	var fileIds []string
	for _, file := range fileList.Items {
		fileIds = append(fileIds, file.Id)
	}

	apps, err := models.GetApps(Dbm, fileIds)
	if err != nil {
		panic(err)
	}

	return c.Render(apps)
}

func (c AlphaWingController) GetLogin() revel.Result {
	next := extractPath(c.Params.Query.Get("next"))
	if len(next) == 0 {
		next = routes.AlphaWingController.Index()
	}

	if c.isLogin() {
		return c.Redirect(next)
	}

	sessionKey := uuid.NewRandom().String()
	c.Session[OAuthSessionKey] = sessionKey
	state := url.Values{}
	state.Set("session_key", sessionKey)
	state.Set("next", next)
	authUrl := c.OAuthConfig.AuthCodeURL(state.Encode())
	return c.Redirect(authUrl)
}

func (c AlphaWingController) GetLogout() revel.Result {
	c.logout()
	return c.Redirect(routes.AlphaWingController.Index())
}

func (c AlphaWingController) GetCallback() revel.Result {
	state, err := url.ParseQuery(c.Params.Query.Get("state"))
	if err != nil {
		panic(err)
	}
	if sessionKey := state.Get("session_key"); sessionKey != c.Session[OAuthSessionKey] {
		panic("invalid session key")
	}
	delete(c.Session, OAuthSessionKey)
	next := extractPath(state.Get("next"))

	code := c.Params.Query.Get("code")
	t := c.transport()
	_, err = t.Exchange(code)
	if err != nil {
		panic(err)
	}
	tokeninfo, err := c.tokenInfo()
	if err != nil {
		panic(err)
	}

	permitted := c.isPermittedEmail(tokeninfo.Email)
	c.Validation.Required(permitted).Message("can't login with unauthorized email")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.AlphaWingController.Index())
	}

	err = Transact(func(txn gorp.SqlExecutor) error {
		user, err := models.FindOrCreateUser(txn, tokeninfo.Email)
		if err != nil {
			return err
		}
		c.login(fmt.Sprint(user.Id))
		return nil
	})
	if err != nil {
		panic(err)
	}

	return c.Redirect(next)
}

func (c *AlphaWingController) UriFor(path string) (*url.URL, error) {
	scheme := "http"
	if c.Request.Header.Get("X-Forwarded-Proto") == "https" {
		scheme = "https"
	}

	return url.Parse(fmt.Sprintf("%s://%s/%s", scheme, c.Request.Host, path))
}

func (c *AlphaWingController) login(userId string) {
	c.Session[LoginSessionKey] = userId
}

func (c *AlphaWingController) logout() {
	delete(c.Session, LoginSessionKey)
}

func (c *AlphaWingController) isLogin() bool {
	_, found := c.Session[LoginSessionKey]
	return found
}

func (c *AlphaWingController) token() (*oauth.Token, error) {
	return c.OAuthConfig.TokenCache.Token()
}

func (c *AlphaWingController) userInfo() (*oauth2.Userinfoplus, error) {
	s, err := c.userGoogleService()
	if err != nil {
		return nil, err
	}
	return s.GetUserInfo()
}

func (c *AlphaWingController) tokenInfo() (*oauth2.Tokeninfo, error) {
	s, err := c.userGoogleService()
	if err != nil {
		return nil, err
	}
	return s.GetTokenInfo()
}

func (c *AlphaWingController) transport() *oauth.Transport {
	token, err := c.token()
	if err != nil {
		return &oauth.Transport{Config: c.OAuthConfig}
	}
	return &oauth.Transport{Config: c.OAuthConfig, Token: token}
}

func (c *AlphaWingController) isPermittedEmail(email string) bool {
	permitted, err := models.IsExistAuthorityForEmail(Dbm, email)
	if err != nil {
		panic(err)
	}

	if permitted {
		return true
	} else {
		emailParts := strings.Split(email, "@")
		domain := emailParts[1]
		for _, permittedDomain := range Conf.PermittedDomains {
			if domain == permittedDomain {
				permitted = true
				break
			}
		}
		return permitted
	}
}

func (c *AlphaWingController) createAudit(resource int, resourceId int, action int) error {
	err := Transact(func(txn gorp.SqlExecutor) error {
		audit := &models.Audit{
			UserId:     c.LoginUserId,
			Resource:   resource,
			ResourceId: resourceId,
			Action:     action,
		}
		return audit.Save(txn)
	})
	if err != nil {
		return err
	}
	return nil
}

func (c *AlphaWingController) SetLoginInfo() revel.Result {
	c.RenderArgs["islogin"] = c.isLogin()
	if c.isLogin() {
		tokeninfo, err := c.tokenInfo()
		if err != nil {
			code, _, _ := models.ParseGoogleApiError(err)
			switch {
			case code == 0 || (400 <= code && code <= 499):
				c.logout()
				c.RenderArgs["islogin"] = false
				return nil
			default:
				panic(err)
			}
		}
		c.RenderArgs["tokeninfo"] = tokeninfo

		userId, err := strconv.Atoi(c.Session[LoginSessionKey])
		if err != nil {
			panic(err)
		}
		c.LoginUserId = userId
	}
	return nil
}

func (c *AlphaWingController) InitOAuthConfig() revel.Result {
	config := &models.WebApplicationConfig{
		ClientId:     Conf.WebApplicationClientId,
		ClientSecret: Conf.WebApplicationClientSecret,
		CallbackUrl:  Conf.WebApplicationCallbackUrl,
		Scope:        []string{oauth2.UserinfoEmailScope, drive.DriveMetadataReadonlyScope},
	}
	tokenCache := &TokenSession{Session: c.Session}

	c.OAuthConfig = models.CreateOAuthConfig(config, tokenCache)

	return nil
}

func (c *AlphaWingController) InitGoogleService() revel.Result {
	config := &models.ServiceAccountConfig{
		ClientEmail: Conf.ServiceAccountClientEmail,
		PrivateKey:  Conf.ServiceAccountPrivateKey,
		Scope:       []string{drive.DriveScope},
	}

	token, err := models.GetServiceAccountToken(config)
	if err != nil {
		panic(err)
	}

	s, err := models.NewGoogleService(token)
	if err != nil {
		panic(err)
	}
	c.GoogleService = s

	capacityInfo, err := s.GetCapacityInfo()
	if err != nil {
		panic(err)
	}
	c.RenderArgs["capacityInfo"] = capacityInfo

	return nil
}

func (c *AlphaWingController) InitRenderArgs() revel.Result {
	c.RenderArgs["organizationName"] = Conf.OrganizationName

	return nil
}

func (c *AlphaWingController) userGoogleService() (*models.GoogleService, error) {
	token, err := c.token()
	if err != nil {
		return nil, err
	}

	s, err := models.NewGoogleService(token)
	if err != nil {
		return nil, err
	}

	return s, nil
}

func extractPath(next string) string {
	n, err := url.Parse(next)
	if err != nil {
		return routes.AlphaWingController.Index()
	}
	return n.Path
}

// ----------------------------------------------------------------------
// TokenSession
type TokenSession struct {
	Session revel.Session
}

const TokenSessionKey = "TokenSessionKey"

// http://code.google.com/p/goauth2/source/browse/oauth/oauth.go#59
func (ts *TokenSession) Token() (*oauth.Token, error) {
	token := &oauth.Token{}
	err := json.Unmarshal([]byte(ts.Session[TokenSessionKey]), token)
	if err != nil {
		return nil, err
	}
	return token, nil
}

func (ts *TokenSession) PutToken(token *oauth.Token) error {
	b, err := json.Marshal(token)
	if err != nil {
		return err
	}
	ts.Session[TokenSessionKey] = string(b)
	return nil
}
