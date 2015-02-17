package controllers

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/kayac/alphawing/app/models"
	"github.com/kayac/alphawing/app/routes"

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

	app, err := bundle.App(c.Txn)
	if err != nil {
		panic(err)
	}

	installUrl, err := c.UriFor(fmt.Sprintf("bundle/%d/download", bundle.Id))
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
	bundle_for_update.Description = bundle.Description
	if err := bundle_for_update.Update(c.Txn); err != nil {
		panic(err)
	}

	c.Flash.Success("Updated!")
	return c.Redirect(routes.BundleControllerWithValidation.GetBundle(bundle_for_update.Id))
}

func (c BundleControllerWithValidation) PostDeleteBundle(bundleId int) revel.Result {
	bundle := c.Bundle
	if err := bundle.Delete(c.Txn, c.GoogleService); err != nil {
		panic(err)
	}

	if err := c.createAudit(models.ResourceBundle, bundleId, models.ActionDelete); err != nil {
		panic(err)
	}

	c.Flash.Success("Deleted!")
	return c.Redirect(routes.AppControllerWithValidation.GetApp(bundle.AppId))
}

func (c BundleControllerWithValidation) GetDownloadApkBundle(bundleId int) revel.Result {
	resp, file, err := c.GoogleService.DownloadFile(c.Bundle.FileId)
	if err != nil {
		panic(err)
	}

	modtime, err := time.Parse(time.RFC3339, file.ModifiedDate)
	if err != nil {
		panic(err)
	}

	err = c.createAudit(models.ResourceBundle, bundleId, models.ActionDownload)
	if err != nil {
		panic(err)
	}

	c.Response.ContentType = "application/vnd.android.package-archive"
	return c.RenderBinary(resp.Body, file.OriginalFilename, revel.Attachment, modtime)
}

func (c BundleControllerWithValidation) GetDownloadIpaBundle(bundleId int) revel.Result {
	return c.Render()
}

func (c *BundleControllerWithValidation) CheckNotFound() revel.Result {
	param := c.Params.Route["bundleId"]
	if 0 < len(param) {
		bundleId, err := strconv.Atoi(param[0])
		if err != nil {
			if err == sql.ErrNoRows {
				c.NotFound("NotFound")
			}
			panic(err)
		}
		bundle, err := models.GetBundle(c.Txn, bundleId)
		if err != nil {
			panic(err)
		}
		c.Bundle = bundle
	}
	return nil
}

func (c *BundleControllerWithValidation) CheckForbidden() revel.Result {
	if c.Bundle != nil {
		bundle := c.Bundle
		s, err := c.userGoogleService()
		if err != nil {
			panic(err)
		}
		_, err = s.GetFile(bundle.FileId)
		if err != nil {
			return c.Forbidden("Forbidden")
		}
	}
	return nil
}
