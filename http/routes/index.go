package routes

import (
	"animatedcloak-backend/modules/labyauth"
	"fmt"

	"github.com/valyala/fasthttp"
)

func Index(ctx *fasthttp.RequestCtx) {
	profile, _ := ctx.UserValue("labyuser").(labyauth.LabyPayload)

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.SetBodyString(fmt.Sprintf("Hello, %s!", profile.Username))
}
