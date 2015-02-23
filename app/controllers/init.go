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
	PermittedDomains           []string
	OrganizationName           string
	WebApplicationClientId     string
	WebApplicationClientSecret string
	WebApplicationCallbackUrl  string
	ServiceAccountClientEmail  string
	ServiceAccountPrivateKey   string
	AaptPath                   string
	PagerDefaultLimit          int
}

func init() {
	// config
	revel.OnAppStart(LoadConfig)

	// gorp
	revel.OnAppStart(InitDB)
	revel.InterceptMethod((*GorpController).Begin, revel.BEFORE)
	revel.InterceptMethod((*GorpController).Commit, revel.AFTER)
	revel.InterceptMethod((*GorpController).Rollback, revel.FINALLY)

	// service account
	revel.InterceptMethod((*AlphaWingController).InitGoogleService, revel.BEFORE)

	// auth
	revel.InterceptMethod((*AlphaWingController).InitOAuthConfig, revel.BEFORE)
	revel.InterceptMethod((*AlphaWingController).SetLoginInfo, revel.BEFORE)
	revel.InterceptMethod((*AuthController).CheckLogin, revel.BEFORE)
	revel.InterceptMethod((*AppControllerWithValidation).CheckNotFound, revel.BEFORE)
	revel.InterceptMethod((*AppControllerWithValidation).CheckForbidden, revel.BEFORE)
	revel.InterceptMethod((*BundleControllerWithValidation).CheckNotFound, revel.BEFORE)
	revel.InterceptMethod((*BundleControllerWithValidation).CheckForbidden, revel.BEFORE)

	// document
	revel.OnAppStart(GenerateApiDocument)

	// args
	revel.InterceptMethod((*AlphaWingController).InitRenderArgs, revel.AFTER)
}

func LoadConfig() {
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

	aaptPath := revel.Config.StringDefault("aapt.path", "/usr/local/bin/aapt")

	pagerDefaultLimit := revel.Config.IntDefault("app.pager.default.limit", 25)

	Conf = &Config{
		PermittedDomains:           strings.Split(permittedDomain, ","),
		OrganizationName:           organizationName,
		WebApplicationClientId:     webApplicationClientId,
		WebApplicationClientSecret: webApplicationClientSecret,
		WebApplicationCallbackUrl:  webApplicationCallbackUrl,
		ServiceAccountClientEmail:  serviceAccountClientEmail,
		ServiceAccountPrivateKey:   serviceAccountPrivateKey,
		AaptPath:                   aaptPath,
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
