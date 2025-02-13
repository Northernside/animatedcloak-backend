package http

import (
	"animatedcloak-backend/http/routes"
	"animatedcloak-backend/modules/env"
	"animatedcloak-backend/modules/labyauth"
	"fmt"

	"github.com/valyala/fasthttp"
)

var (
	endpointList = make(map[string]func(*fasthttp.RequestCtx))
)

func init() {
	adminHandler("PUT", "/api/update", routes.UploadAddon)
	defaultHandler("GET", "/api/update", routes.UpdateAddon)

	userHandler("PUT", "/api/cloak", routes.UploadCloak)
	defaultHandler("GET", "/api/cloak", routes.GetCloak)
}

func StartAPI() {
	// start fasthttp server

	server := &fasthttp.Server{
		Handler: func(ctx *fasthttp.RequestCtx) {
			if handler, ok := endpointList[fmt.Sprintf("%s:%s", string(ctx.Method()), string(ctx.Path()))]; ok {
				handler(ctx)
				return
			}

			ctx.Response.SetStatusCode(fasthttp.StatusNotFound)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBodyString(`{"error": "Not Found"}`)
		},
	}

	if err := server.ListenAndServe(fmt.Sprintf("%s:%s", env.GetEnv("HTTP_BINDING", "127.0.0.1"), env.GetEnv("HTTP_PORT", "13337"))); err != nil {
		panic("error in ListenAndServe: " + err.Error())
	}
}

func defaultHandler(method, path string, handler fasthttp.RequestHandler) {
	endpointList[fmt.Sprintf("%s:%s", method, path)] = func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Method()) != method {
			ctx.Response.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBodyString(`{"error": "Method Not Allowed"}`)
			return
		}

		handler(ctx)
	}
}

func userHandler(method, path string, handler fasthttp.RequestHandler) {
	endpointList[fmt.Sprintf("%s:%s", method, path)] = func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Method()) != method {
			ctx.Response.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBodyString(`{"error": "Method Not Allowed"}`)
			return
		}

		token := string(ctx.Request.Header.Cookie("token"))
		if token == "" {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(`{"error": "Unauthorized"}`)
			return
		}

		labyPayload, err := labyauth.VerifyToken(token)
		if err != nil {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(`{"error": "Unauthorized", "message": "` + err.Error() + `"}`)
			return
		}

		ctx.SetUserValue("labyuser", labyPayload)
		handler(ctx)
	}
}

func adminHandler(method, path string, handler fasthttp.RequestHandler) {
	endpointList[fmt.Sprintf("%s:%s", method, path)] = func(ctx *fasthttp.RequestCtx) {
		if string(ctx.Method()) != method {
			ctx.Response.SetStatusCode(fasthttp.StatusMethodNotAllowed)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.Response.SetBodyString(`{"error": "Method Not Allowed"}`)
			return
		}

		token := string(ctx.Request.Header.Cookie("token"))
		if token == "" {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(`{"error": "Unauthorized"}`)
			return
		}

		labyPayload, err := labyauth.VerifyToken(token)
		if err != nil {
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.SetStatusCode(fasthttp.StatusUnauthorized)
			ctx.SetBodyString(`{"error": "Unauthorized", "message": "` + err.Error() + `"}`)
			return
		}

		for _, uuid := range labyauth.UUIDWhitelist {
			if uuid == labyPayload.UUID {
				ctx.SetUserValue("labyuser", labyPayload)
				handler(ctx)
				return
			}
		}

		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetStatusCode(fasthttp.StatusForbidden)
		ctx.SetBodyString(`{"error": "Forbidden"}`)
		return
	}
}
