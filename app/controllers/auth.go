package controllers

import (
	"net/url"

	"github.com/kayac/alphawing/app/routes"
	"github.com/revel/revel"
)

type AuthController struct {
	AlphaWingController
}

func (c *AuthController) CheckLogin() revel.Result {
	if c.isLogin() {
		return nil
	}

	loginUrl := routes.AlphaWingController.GetLogin()
	next := c.Request.URL.Path
	return c.Redirect(loginUrl + "?next=" + url.QueryEscape(next))
}
