package controllers

import (
	"database/sql"
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

func (c ApiController) NewJsonResponseDeleteBundle(stat int, mes []string) *JsonResponse {
	return c.NewJsonResponse(stat, mes)
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

func (c ApiController) PostDeleteBundle(token string, file_id string) revel.Result {
	_, err := models.GetAppByApiToken(c.Txn, token)
	if err != nil {
		c.Response.Status = http.StatusUnauthorized
		return c.RenderJson(c.NewJsonResponseDeleteBundle(c.Response.Status, []string{"Token is invalid."}))
	}

	c.Validation.Required(file_id).Message("file_id is required.")
	if c.Validation.HasErrors() {
		var errors []string
		for _, err := range c.Validation.Errors {
			errors = append(errors, err.String())
		}
		c.Response.Status = http.StatusBadRequest
		return c.RenderJson(c.NewJsonResponseDeleteBundle(c.Response.Status, errors))
	}

	bundle, err := models.GetBundleByFileId(c.Txn, file_id)
	if err != nil {
		if err == sql.ErrNoRows {
			c.Response.Status = http.StatusInternalServerError
			return c.RenderJson(c.NewJsonResponseDeleteBundle(c.Response.Status, []string{"Bundle not found."}))
		}
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(c.NewJsonResponseDeleteBundle(c.Response.Status, []string{err.Error()}))
	}

	err = bundle.Delete(c.Txn, c.GoogleService)
	if err != nil {
		c.Response.Status = http.StatusInternalServerError
		return c.RenderJson(c.NewJsonResponseDeleteBundle(c.Response.Status, []string{err.Error()}))
	}

	c.Response.Status = http.StatusOK
	return c.RenderJson(c.NewJsonResponseDeleteBundle(c.Response.Status, []string{"Bundle is deleted!"}))
}
