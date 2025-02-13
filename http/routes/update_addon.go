package routes

import (
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/valyala/fasthttp"
)

var (
	latest_version string
	addon_jar      []byte
)

func init() {
	if err := os.MkdirAll(uploadDir, os.ModePerm); err != nil {
		panic("Error creating addon directory")
	}

	files, err := filepath.Glob("addon/*.jar")
	if err != nil {
		panic(err)
	}

	if len(files) == 0 {
		panic("No addon.jar found")
	}

	latest_version = strings.TrimSuffix(files[0], ".jar")
	latest_version = strings.TrimPrefix(latest_version, "addon/")

	addon_jar, err = os.ReadFile(files[0])
	if err != nil {
		panic(err)
	}
}

func UpdateAddon(ctx *fasthttp.RequestCtx) {
	version := string(ctx.QueryArgs().Peek("version"))
	if version == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "MISSING_VERSION"}`)
		return
	}

	if isOutdated(version) {
		ctx.SetStatusCode(fasthttp.StatusOK)
		ctx.Response.Header.Set("Content-Type", "application/java-archive")
		ctx.Response.Header.Set("Content-Disposition", "attachment; filename=animated-cloaks-"+latest_version+".jar")
		ctx.SetBody(addon_jar)
		return
	}

	ctx.SetStatusCode(fasthttp.StatusNoContent)
	ctx.Response.Header.Set("Content-Type", "application/json")
	ctx.SetBodyString(`{"error": "UP_TO_DATE"}`)
}

func isOutdated(version string) bool {
	latestParts := strings.Split(latest_version, ".")
	versionParts := strings.Split(version, ".")

	for i := 0; i < len(latestParts); i++ {
		latestPart, _ := strconv.Atoi(latestParts[i])
		versionPart, _ := strconv.Atoi(versionParts[i])

		if latestPart > versionPart {
			return true
		} else if latestPart < versionPart {
			return false
		}
	}

	return false
}
