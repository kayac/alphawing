package controllers

import (
	"encoding/json"
	"io/ioutil"
	"strings"

	"github.com/kayac/alphawing/app/models"

	"github.com/revel/revel"
)

var (
	Conf *Config
)

type Config struct {
	Secret                     string
	PermittedDomains           []string
	OrganizationName           string
	WebApplicationClientId     string
	WebApplicationClientSecret string
	WebApplicationCallbackUrl  string
	ServiceAccountClientEmail  string
	ServiceAccountPrivateKey   string
	PagerDefaultLimit          int
}

func init() {
	// config
	revel.OnAppStart(LoadConfig)

	// gorp
	revel.OnAppStart(InitDB)

	// service account
	revel.InterceptMethod((*AlphaWingController).InitGoogleService, revel.BEFORE)

	// auth
	revel.InterceptMethod((*AlphaWingController).InitOAuthConfig, revel.BEFORE)
	revel.InterceptMethod((*AlphaWingController).SetLoginInfo, revel.BEFORE)
	revel.InterceptMethod((*AuthController).CheckLogin, revel.BEFORE)

	// validate app
	revel.InterceptMethod((*AppControllerWithValidation).CheckNotFound, revel.BEFORE)
	revel.InterceptMethod((*AppControllerWithValidation).CheckForbidden, revel.BEFORE)

	// validate bundle
	revel.InterceptMethod((*BundleControllerWithValidation).CheckNotFound, revel.BEFORE)
	revel.InterceptMethod((*BundleControllerWithValidation).CheckForbidden, revel.BEFORE)
	revel.InterceptMethod((*LimitedTimeController).CheckNotFound, revel.BEFORE)

	// validate limited time token
	revel.InterceptMethod((*LimitedTimeController).CheckValidLimitedTimeToken, revel.BEFORE)

	// document
	revel.OnAppStart(GenerateApiDocument)

	// args
	revel.InterceptMethod((*AlphaWingController).InitRenderArgs, revel.AFTER)
}

func LoadConfig() {
	secret, found := revel.Config.String("app.secret")
	if !found {
		panic("undefined config: app.secret")
	}

	permittedDomain, found := revel.Config.String("app.permitteddomain")
	if !found {
		panic("undefined config: app.permitteddomain")
	}
	organizationName, _ := revel.Config.String("app.organizationname")

	webApplicationClientId, found := revel.Config.String("google.webapplication.clientid")
	if !found {
		panic("undefined config: google.webapplication.clientid")
	}
	webApplicationClientSecret, found := revel.Config.String("google.webapplication.clientsecret")
	if !found {
		panic("undefined config: google.webapplication.clientsecret")
	}
	webApplicationCallbackUrl, found := revel.Config.String("google.webapplication.callbackurl")
	if !found {
		panic("undefined config: google.webapplication.callbackurl")
	}

	serviceAccountKeyPath, found := revel.Config.String("google.serviceaccount.keypath")
	if !found {
		panic("undefined config: google.serviceaccount.keypath")
	}
	keyBytes, err := ioutil.ReadFile(serviceAccountKeyPath)
	if err != nil {
		panic(err)
	}
	var keyMap map[string]string
	if err := json.Unmarshal(keyBytes, &keyMap); err != nil {
		panic(err)
	}
	serviceAccountClientEmail := keyMap["client_email"]
	serviceAccountPrivateKey := keyMap["private_key"]

	pagerDefaultLimit := revel.Config.IntDefault("app.pager.default.limit", 25)

	Conf = &Config{
		Secret:                     secret,
		PermittedDomains:           strings.Split(permittedDomain, ","),
		OrganizationName:           organizationName,
		WebApplicationClientId:     webApplicationClientId,
		WebApplicationClientSecret: webApplicationClientSecret,
		WebApplicationCallbackUrl:  webApplicationCallbackUrl,
		ServiceAccountClientEmail:  serviceAccountClientEmail,
		ServiceAccountPrivateKey:   serviceAccountPrivateKey,
		PagerDefaultLimit:          pagerDefaultLimit,
	}
}

func GenerateApiDocument() {
	html, err := models.GenerateApiDocumentHtml(revel.BasePath + "/docs/api.md")
	if err != nil {
		panic(err)
	}

	ioutil.WriteFile(revel.AppPath+"/views/ApiController/GetDocument.html", []byte(html), 0644)
}
