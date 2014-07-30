package controllers

import r "github.com/revel/revel"

type AuthController struct {
	AlphaWingController
}

func (c *AuthController) CheckLogin() r.Result {
	if !c.isLogin() {
		url := c.OAuthConfig.AuthCodeURL("")
		return c.Redirect(url)
	}
	return nil
}
