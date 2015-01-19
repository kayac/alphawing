package controllers

import (
	"github.com/kayac/alphawing/app/routes"
	r "github.com/revel/revel"
	"net/url"
)

type AuthController struct {
	AlphaWingController
}

func (c *AuthController) CheckLogin() r.Result {
	if c.isLogin() {
		return nil
	}

	loginUrl := routes.AlphaWingController.GetLogin()
	next := c.Request.URL.Path
	return c.Redirect(loginUrl + "?next=" + url.QueryEscape(next))
}
