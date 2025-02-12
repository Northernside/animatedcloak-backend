package routes

import (
	"io"
	"os"
	"path/filepath"

	"github.com/valyala/fasthttp"
)

func GetCloak(ctx *fasthttp.RequestCtx) {
	uuid := string(ctx.QueryArgs().Peek("uuid"))
	if uuid == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "MISSING_UUID"}`)
		return
	}

	filePath := filepath.Join(uploadDir, uuid+".gif")
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		ctx.SetStatusCode(fasthttp.StatusNotFound)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "CLOAK_NOT_FOUND"}`)
		return
	}

	file, err := os.Open(filePath)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_OPENING_FILE"}`)
		return
	}
	defer file.Close()

	ctx.Response.Header.Set("Content-Type", "image/gif")

	// copy the file to response body
	if _, err := io.Copy(ctx, file); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_READING_FILE"}`)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
}
