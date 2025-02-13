package routes

import (
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/valyala/fasthttp"
)

func init() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		panic("Error creating upload directory")
	}
}

func UploadAddon(ctx *fasthttp.RequestCtx) {
	form, err := ctx.MultipartForm()
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.SetBodyString("Error parsing form data")
		return
	}

	fileHeaders := form.File["addon"]
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
	if !strings.HasSuffix(strings.ToLower(fileName), ".jar") {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "INVALID_FILE_TYPE"}`)
		return
	}

	fileName = strings.TrimSuffix(fileName, ".jar")
	fileBytes, err := io.ReadAll(srcFile)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_READING_FILE"}`)
		return
	}

	versionPattern := `^\d+\.\d+\.\d+$`
	if !regexp.MustCompile(versionPattern).MatchString(fileName) {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "INVALID_VERSION_FORMAT"}`)
		return
	}

	files, err := filepath.Glob("addon/*.jar")
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_LISTING_FILES"}`)
		return
	}

	for _, file := range files {
		err = os.Remove(file)
		if err != nil {
			ctx.SetStatusCode(fasthttp.StatusInternalServerError)
			ctx.Response.Header.Set("Content-Type", "application/json")
			ctx.SetBodyString(`{"error": "ERROR_DELETING_FILE"}`)
			return
		}
	}

	filePath := filepath.Join("addon/", fileName+".jar")
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

	latest_version = fileName
	addon_jar = fileBytes
}
