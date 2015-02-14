package controllers

import (
	"database/sql"
	"errors"
	"os"
	"path/filepath"
	"strconv"

	"github.com/kayac/alphawing/app/models"
	"github.com/kayac/alphawing/app/routes"

	"github.com/revel/revel"
)

type AppController struct {
	AuthController
	App *models.App
}

// not found, permission check
type AppControllerWithValidation struct {
	AppController
}

// ------------------------------------------------------
// AppController
func (c AppController) GetCreateApp() revel.Result {
	app := &models.App{}
	return c.Render(app)
}

func (c AppController) PostCreateApp(app models.App) revel.Result {
	c.Validation.Required(app.Title).Message("Title is required.")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.AppController.GetCreateApp())
	}

	if err := models.CreateApp(c.Txn, c.GoogleService, &app); err != nil {
		panic(err)
	}

	tokeninfo, err := c.tokenInfo()
	if err != nil {
		panic(err)
	}
	authority := &models.Authority{
		Email: tokeninfo.Email,
	}
	if err := app.CreateAuthority(c.Txn, c.GoogleService, authority); err != nil {
		panic(err)
	}

	if err = c.createAudit(models.ResourceApp, app.Id, models.ActionCreate); err != nil {
		panic(err)
	}

	c.Flash.Success("Created!")
	return c.Redirect(routes.AppControllerWithValidation.GetApp(app.Id))
}

// ------------------------------------------------------
// AppControllerWithValidation
func (c AppControllerWithValidation) GetApp(appId int) revel.Result {
	app := c.App

	authorities, err := app.Authorities(c.Txn)
	if err != nil {
		panic(err)
	}

	bundles, err := app.Bundles(c.Txn)
	if err != nil {
		panic(err)
	}

	return c.Render(app, authorities, bundles)
}

func (c AppControllerWithValidation) GetUpdateApp(appId int) revel.Result {
	app := c.App
	return c.Render(app)
}

func (c AppControllerWithValidation) PostUpdateApp(appId int, app models.App) revel.Result {
	if appId != app.Id {
		c.Flash.Error("Parameter is invalid.")
		c.Redirect(routes.AppControllerWithValidation.GetUpdateApp(app.Id))
	}

	c.Validation.Required(app.Title).Message("Title is required.")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.AppControllerWithValidation.GetUpdateApp(app.Id))
	}

	if err := app.Update(c.Txn); err != nil {
		panic(err)
	}

	if err := c.GoogleService.UpdateFileTitle(c.App.FileId, app.Title); err != nil {
		panic(err)
	}

	c.Flash.Success("Updated!")
	return c.Redirect(routes.AppControllerWithValidation.GetApp(app.Id))
}

func (c AppControllerWithValidation) PostRefreshToken(appId int, app models.App) revel.Result {
	if appId != app.Id {
		c.Flash.Error("Parameter is invalid")
		c.Redirect(routes.AppControllerWithValidation.GetApp(app.Id))
	}

	if err := app.RefreshToken(c.Txn); err != nil {
		panic(err)
	}

	c.Flash.Success("Refreshed!")
	return c.Redirect(routes.AppControllerWithValidation.GetApp(app.Id))
}

func (c AppControllerWithValidation) PostDeleteApp(appId int) revel.Result {
	app := c.App

	app.Delete(c.Txn, c.GoogleService)

	if err := c.createAudit(models.ResourceApp, appId, models.ActionDelete); err != nil {
		panic(err)
	}

	c.Flash.Success("Deleted!")
	return c.Redirect(routes.AlphaWingController.Index())
}

func (c AppControllerWithValidation) GetCreateBundle(appId int) revel.Result {
	app := c.App
	bundle := &models.Bundle{AppId: appId}
	return c.Render(app, bundle)
}

func (c AppControllerWithValidation) PostCreateBundle(appId int, bundle models.Bundle, file *os.File) revel.Result {
	if appId != bundle.AppId {
		c.Flash.Error("Parameter is invalid.")
		c.Redirect(routes.AppControllerWithValidation.GetApp(appId))
	}

	var filename string
	if _, ok := c.Params.Files["file"]; ok {
		filename = c.Params.Files["file"][0].Filename
	}
	ext := filepath.Ext(filename)

	c.Validation.Required(file != nil).Message("File is required.")
	c.Validation.Required(models.BundleFileExtension(ext).IsValid()).Message("File extension is not valid.")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.AppControllerWithValidation.GetCreateBundle(appId))
	}

	bundle.File = file
	bundle.PlatformType = models.BundleFileExtension(ext).PlatformType()

	if err := c.App.CreateBundle(c.Txn, c.GoogleService, &bundle); err != nil {
		if aperr, ok := err.(*models.AppParseError); ok {
			c.Flash.Error(aperr.Error())
			return c.Redirect(routes.AppControllerWithValidation.GetCreateBundle(appId))
		}
		panic(err)
	}

	if err := c.createAudit(models.ResourceBundle, bundle.Id, models.ActionCreate); err != nil {
		panic(err)
	}

	c.Flash.Success("Created!")
	return c.Redirect(routes.BundleControllerWithValidation.GetBundle(bundle.Id))
}

func (c AppControllerWithValidation) PostCreateAuthority(appId int, email string) revel.Result {
	app := c.App

	c.Validation.Required(email).Message("Email is required.")
	c.Validation.Email(email).Message("Email is invalid.")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.AppControllerWithValidation.GetApp(appId))
	}

	found, err := app.HasAuthorityForEmail(c.Txn, email)
	if err != nil {
		panic(err)
	}
	c.Validation.Required(!found).Message(email + " is already registered.")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.AppControllerWithValidation.GetApp(appId))
	}

	authority := &models.Authority{
		Email: email,
	}
	app.CreateAuthority(c.Txn, c.GoogleService, authority)

	if err := c.createAudit(models.ResourceAuthority, authority.Id, models.ActionCreate); err != nil {
		panic(err)
	}

	c.Flash.Success("Registered!")
	return c.Redirect(routes.AppControllerWithValidation.GetApp(appId))
}

func (c AppControllerWithValidation) PostDeleteAuthority(appId, authorityId int) revel.Result {
	app := c.App

	authority, err := models.GetAuthority(c.Txn, authorityId)
	if err != nil {
		panic(err)
	}

	if appId != authority.AppId {
		c.Flash.Error("Parameter is invalid.")
		return c.Redirect(routes.AppControllerWithValidation.GetApp(appId))
	}

	if err := app.DeleteAuthority(c.Txn, c.GoogleService, authority); err != nil {
		panic(err)
	}

	if err := c.createAudit(models.ResourceAuthority, authority.Id, models.ActionDelete); err != nil {
		panic(err)
	}

	c.Flash.Success("Deleted!")
	return c.Redirect(routes.AppControllerWithValidation.GetApp(appId))
}

func (c *AppControllerWithValidation) CheckNotFound() revel.Result {
	param := c.Params.Route["appId"]
	if len(param) == 0 {
		panic(errors.New("AppId is Required."))
	}
	appId, err := strconv.Atoi(param[0])
	if err != nil {
		panic(err)
	}
	app, err := models.GetApp(c.Txn, appId)
	if err != nil {
		if err == sql.ErrNoRows {
			c.NotFound("NotFound")
		}
		panic(err)
	}
	c.App = app
	return nil
}

func (c *AppControllerWithValidation) CheckForbidden() revel.Result {
	if c.App == nil {
		c.NotFound("NotFound")
	}
	app := c.App
	s, err := c.userGoogleService()
	if err != nil {
		panic(err)
	}
	_, err = s.GetFile(app.FileId)
	if err != nil {
		return c.Forbidden("Forbidden")
	}
	return nil
}
