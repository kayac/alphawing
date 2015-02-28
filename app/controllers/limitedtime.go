package controllers

import (
	"database/sql"
	"fmt"
	"strconv"
	"time"

	"github.com/kayac/alphawing/app/models"
	"github.com/revel/revel"
)

type LimitedTimeController struct {
	AlphaWingController
	Bundle *models.Bundle
}

func (c *LimitedTimeController) GetDownloadPlist(bundleId int) revel.Result {
	bundle := c.Bundle

	t, err := models.NewLimitedTimeTokenInfoByKey(Conf.Secret)
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

func (c *LimitedTimeController) CheckValidLimitedTimeToken() revel.Result {
	bundle := c.Bundle

	if bundle == nil {
		revel.ERROR.Printf("Bundle is not found.")
		return c.NotFound("")
	}

	token := c.Params.Query.Get(models.TokenKey)
	seed := c.Params.Query.Get(models.SeedKey)
	limit := c.Params.Query.Get(models.LimitKey)

	c.Validation.Required(token)
	c.Validation.Required(seed)
	c.Validation.Required(limit)
	if c.Validation.HasErrors() {
		revel.ERROR.Printf("Parameters are invalid.")
		return c.NotFound("")
	}

	tokenInfo, err := models.NewLimitedTimeTokenInfo(token, seed, limit)
	if err != nil {
		revel.ERROR.Printf(err.Error())
		return c.NotFound("")
	}

	ok, err := tokenInfo.IsValid(Conf.Secret)
	if err != nil {
		revel.ERROR.Printf(err.Error())
		return c.NotFound("")
	}
	if !ok {
		revel.ERROR.Printf("Token is invalid.")
		return c.NotFound("")
	}

	return nil
}

func (c *LimitedTimeController) CheckNotFound() revel.Result {
	bundleIdStr := c.Params.Get("bundleId")

	c.Validation.Required(bundleIdStr)
	if c.Validation.HasErrors() {
		revel.ERROR.Printf("BundleId is required.")
		return c.NotFound("")
	}

	bundleId, err := strconv.Atoi(bundleIdStr)
	if err != nil {
		revel.ERROR.Printf(err.Error())
		return c.NotFound("")
	}

	bundle, err := models.GetBundle(Dbm, bundleId)
	if err != nil {
		if err == sql.ErrNoRows {
			revel.ERROR.Printf("Bundle is not found.")
		} else {
			revel.ERROR.Printf(err.Error())
		}
		return c.NotFound("")
	}
	c.Bundle = bundle

	return nil
}
