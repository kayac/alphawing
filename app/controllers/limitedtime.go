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

type LimitedTimeController struct {
	AlphaWingController
	Bundle *models.Bundle
}

func (c *LimitedTimeController) GetDownloadPlist(bundleId int) revel.Result {
	bundle := c.Bundle

	t, err := models.NewLimitedTimeTokenInfo()
	if err != nil {
		panic(err)
	}
	v := t.UrlValues()

	ipaUrl, err := c.UriFor(fmt.Sprintf("bundle/%d/download_ipa", bundle.Id))
	if err != nil {
		panic(err)
	}
	ipaUrl.RawQuery = v.Encode()

	r, err := bundle.PlistReader(Dbm, ipaUrl)
	if err != nil {
		panic(err)
	}

	c.Response.ContentType = "application/x-plist"
	return c.RenderBinary(r, models.PlistFileName, revel.Attachment, time.Now())
}

func (c *LimitedTimeController) GetDownloadIpa(bundleId int) revel.Result {
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

	c.Response.ContentType = "application/octet-stream"
	return c.RenderBinary(resp.Body, file.OriginalFilename, revel.Attachment, modtime)
}

func (c *LimitedTimeController) CheckValidToken() revel.Result {
	bundle := c.Bundle
	if c.Bundle == nil {
		return c.NotFound("NotFound")
	}

	token := c.Params.Get(models.TokenKey)
	seed := c.Params.Get(models.SeedKey)
	limit := c.Params.Get(models.LimitKey)

	c.Validation.Required(token).Message(models.TokenKey + " is required")
	c.Validation.Required(seed).Message(models.SeedKey + " is required")
	c.Validation.Required(limit).Message(models.LimitKey + "is required")
	if c.Validation.HasErrors() {
		c.Validation.Keep()
		c.FlashParams()
		return c.Redirect(routes.BundleControllerWithValidation.GetDownloadBundle(bundle.Id))
	}

	ok, err := models.LimitedTimeToken(token).IsValid(seed, limit)
	if err != nil {
		panic(err)
	}
	if !ok {
		return c.Forbidden("Forbidden")
	}

	return nil
}

func (c *LimitedTimeController) CheckNotFound() revel.Result {
	bundleIdStr := c.Params.Get("bundleId")
	if len(bundleIdStr) == 0 {
		return c.NotFound("NotFound")
	}

	bundleId, err := strconv.Atoi(bundleIdStr)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.NotFound("NotFound")
		}
		panic(err)
	}
	bundle, err := models.GetBundle(Dbm, bundleId)
	if err != nil {
		panic(err)
	}
	c.Bundle = bundle

	return nil
}
