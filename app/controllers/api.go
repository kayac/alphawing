package controllers

import (
	"net/http"
	"os"

	"github.com/kayac/alphawing/app/models"

	"github.com/revel/revel"
)

type JsonResponse struct {
	Status  int      `json:"status"`
	Message []string `json:"message"`
}

type JsonResponseUploadBundle struct {
	*JsonResponse
	Content *models.BundleJsonResponse `json:"content"`
}

type JsonResponseListBundle struct {
	*JsonResponse
	Content *models.BundlesJsonResponse `json:"content"`
}

type ApiController struct {
	AlphaWingController
}

func (c ApiController) NewJsonResponse(stat int, mes []string) *JsonResponse {
	return &JsonResponse{
		Status:  stat,
		Message: mes,
	}
}

func (c ApiController) NewJsonResponseUploadBundle(stat int, mes []string, content *models.BundleJsonResponse) *JsonResponseUploadBundle {
	return &JsonResponseUploadBundle{
		c.NewJsonResponse(stat, mes),
		content,
	}
}

func (c ApiController) NewJsonResponseListBundle(stat int, mes []string, content *models.BundlesJsonResponse) *JsonResponseListBundle {
	return &JsonResponseListBundle{
		c.NewJsonResponse(stat, mes),
		content,
	}
}

func (c ApiController) GetDocument() revel.Result {
	return c.Render()
}

func (c ApiController) PostUploadBundle(token string, description string, file *os.File) revel.Result {
	app, err := models.GetAppByApiToken(c.Txn, token)
	if err != nil {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJson(c.NewJsonResponseUploadBundle(c.Response.Status, []string{"Token is invalid."}, nil))
	}

	c.Validation.Required(file != nil).Message("File is required.")
	if c.Validation.HasErrors() {
		var errors []string
		for _, err := range c.Validation.Errors {
			errors = append(errors, err.String())
		}
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(c.NewJsonResponseUploadBundle(c.Response.Status, errors, nil))
	}

	bundle := &models.Bundle{
		Description: description,
		File:        file,
	}

	if err := app.CreateBundle(c.Txn, c.GoogleService, Conf.AaptPath, bundle); err != nil {
		if aperr, ok := err.(*models.ApkParseError); ok {
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJson(c.NewJsonResponseUploadBundle(c.Response.Status, []string{aperr.Error()}, nil))
		}
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(c.NewJsonResponseUploadBundle(c.Response.Status, []string{err.Error()}, nil))
	}

	content, err := bundle.JsonResponse(&c)
	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(c.NewJsonResponseUploadBundle(c.Response.Status, []string{err.Error()}, nil))
	}

	c.Response.Status = http.StatusOK
	return c.RenderJson(c.NewJsonResponseUploadBundle(c.Response.Status, []string{"Bundle is created!"}, content))
}

func (c ApiController) GetListBundle(token string, page int) revel.Result {
	app, err := models.GetAppByApiToken(c.Txn, token)
	if err != nil {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJson(c.NewJsonResponseListBundle(c.Response.Status, []string{"Token is invalid."}, nil))
	}

	bundles, totalCount, err := app.BundlesWithPager(c.Txn, page, Conf.PagerDefaultLimit)
	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(c.NewJsonResponseListBundle(c.Response.Status, []string{err.Error()}, nil))
	}

	bundlesJsonResponse, err := bundles.JsonResponse(&c)
	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(c.NewJsonResponseListBundle(c.Response.Status, []string{err.Error()}, nil))
	}

	content := &models.BundlesJsonResponse{
		totalCount,
		page,
		Conf.PagerDefaultLimit,
		bundlesJsonResponse,
	}

	c.Response.Status = http.StatusOK

	return c.RenderJson(c.NewJsonResponseListBundle(c.Response.Status, []string{"Bundle List"}, content))
}
