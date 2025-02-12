package routes

import (
	"animatedcloak-backend/modules/labyauth"
	"bytes"
	"fmt"
	"image/gif"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/valyala/fasthttp"
)

var uploadDir = "uploads"

func init() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		panic("Error creating upload directory")
	}
}

func UploadCloak(ctx *fasthttp.RequestCtx) {
	profile, _ := ctx.UserValue("labyuser").(labyauth.LabyPayload)

	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Error parsing form data")
		return
	}

	fileHeaders := form.File["cloak"]
	if len(fileHeaders) == 0 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "NO_FILE"}`)
		return
	}

	fileHeader := fileHeaders[0]
	srcFile, err := fileHeader.Open()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_OPENING_FILE"}`)
		return
	}
	defer srcFile.Close()

	fileName := fileHeader.Filename
	if !strings.HasSuffix(strings.ToLower(fileName), ".gif") {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "INVALID_FILE_TYPE"}`)
		return
	}

	fileBytes, err := io.ReadAll(srcFile)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_READING_FILE"}`)
		return
	}

	img, err := gif.DecodeAll(bytes.NewReader(fileBytes))
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "INVALID_DATA"}`)
		return
	}

	bounds := img.Image[0].Bounds()
	width := bounds.Dx()
	height := bounds.Dy()
	if width != 352 || height != 272 {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "INVALID_DIMENSIONS"}`)
		return
	}

	filePath := filepath.Join(uploadDir, fmt.Sprintf("%s.gif", profile.UUID))
	err = os.WriteFile(filePath, fileBytes, os.ModePerm)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_SAVING_FILE"}`)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusOK)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.SetBodyString(`{"success": true}`)
}
