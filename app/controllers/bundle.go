package controllers

import (
	"database/sql"
	"fmt"
	"strconv"

	"github.com/kayac/alphawing/app/models"
	"github.com/kayac/alphawing/app/routes"

	"github.com/coopernurse/gorp"
	"github.com/revel/revel"
)

type BundleController struct {
	AuthController
	Bundle *models.Bundle
}

// not found, permission check
type BundleControllerWithValidation struct {
	BundleController
}

// ------------------------------------------------------
// BundleController
//func (c BundleController) Get|PostHogeBundle(args...) revel.Result {
//}

// ------------------------------------------------------
// BundleControllerWithValidation
func (c BundleControllerWithValidation) GetBundle(bundleId int) revel.Result {
	bundle := c.Bundle

	app, err := bundle.App(Dbm)
	if err != nil {
		panic(err)
	}

	installUrl, err := c.UriFor(fmt.Sprintf("bundle/%d", bundle.Id))
	if err != nil {
		panic(err)
	}

	return c.Render(bundle, app, installUrl)
}

func (c BundleControllerWithValidation) GetUpdateBundle(bundleId int) revel.Result {
	bundle := c.Bundle
	return c.Render(bundle)
}

func (c BundleControllerWithValidation) PostUpdateBundle(bundleId int, bundle models.Bundle) revel.Result {
	bundle_for_update := c.Bundle
	err := Transact(func(txn gorp.SqlExecutor) error {
		bundle_for_update.Description = bundle.Description
		return bundle_for_update.Update(txn)
	})
	if err != nil {
		panic(err)
	}

	c.Flash.Success("Updated!")
	return c.Redirect(routes.BundleControllerWithValidation.GetBundle(bundle_for_update.Id))
}

func (c BundleControllerWithValidation) PostDeleteBundle(bundleId int) revel.Result {
	bundle := c.Bundle
	err := Transact(func(txn gorp.SqlExecutor) error {
		return bundle.Delete(txn, c.GoogleService)
	})
	if err != nil {
		panic(err)
	}

	if err := c.createAudit(models.ResourceBundle, bundleId, models.ActionDelete); err != nil {
		panic(err)
	}

	c.Flash.Success("Deleted!")
	return c.Redirect(routes.AppControllerWithValidation.GetApp(bundle.AppId))
}

func (c BundleControllerWithValidation) GetDownloadBundle(bundleId int) revel.Result {
	bundle := c.Bundle

	plistUrl, err := c.UriFor(fmt.Sprintf("bundle/%d/download_plist", bundle.Id))
	if err != nil {
		panic(err)
	}

	signatureInfo := models.NewLimitedTimeSignatureInfo(plistUrl.Host, plistUrl.Path)
	signatureInfo.RefreshSignature(Conf.Secret)

	plistUrl.RawQuery = signatureInfo.UrlValues().Encode()

	return c.Render(plistUrl)
}

func (c BundleControllerWithValidation) GetDownloadApk(bundleId int) revel.Result {
	url, err := c.GoogleService.GetDownloadURL(c.Bundle.FileId)
	if err != nil {
		panic(err)
	}

	return c.Redirect(url)
}

func (c *BundleControllerWithValidation) CheckNotFound() revel.Result {
	bundleIdStr := c.Params.Get("bundleId")

	c.Validation.Required(bundleIdStr).Message("BundleId is required.")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.AlphaWingController.Index())
	}

	bundleId, err := strconv.Atoi(bundleIdStr)
	if err != nil {
		c.Flash.Error("BundleId is invalid.")
		return c.Redirect(routes.AlphaWingController.Index())
	}

	bundle, err := models.GetBundle(Dbm, bundleId)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.NotFound("Bundle is not found.")
		}
		panic(err)
	}
	c.Bundle = bundle

	return nil
}

func (c *BundleControllerWithValidation) CheckForbidden() revel.Result {
	bundle := c.Bundle
	if bundle == nil {
		return c.NotFound("Bundle is not found.")
	}

	app, err := models.GetApp(Dbm, bundle.AppId)
	if err != nil {
		panic(err)
	}

	s, err := c.userGoogleService()
	if err != nil {
		panic(err)
	}

	if _, err = s.GetBucket(app.FileId); err != nil {
		return c.Forbidden("Can't access the app.")
	}

	return nil
}
