package routes

import (
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/valyala/fasthttp"
)

var (
	cache      = make(map[string]cacheEntry)
	cacheMutex sync.Mutex
)

type cacheEntry struct {
	data      []byte
	timestamp time.Time
}

func GetCloak(ctx *fasthttp.RequestCtx) {
	uuid := string(ctx.QueryArgs().Peek("uuid"))
	if uuid == "" {
		ctx.SetStatusCode(fasthttp.StatusBadRequest)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "MISSING_UUID"}`)
		return
	}

	cacheMutex.Lock()
	if entry, found := cache[uuid]; found {
		if time.Since(entry.timestamp) < 5*time.Second {
			ctx.Response.Header.Set("Content-Type", "image/gif")
			ctx.SetBody(entry.data)
			ctx.SetStatusCode(fasthttp.StatusOK)
			cacheMutex.Unlock()
			return
		}

		delete(cache, uuid)
	}
	cacheMutex.Unlock()

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

	data, err := io.ReadAll(file)
	if err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_READING_FILE"}`)
		return
	}

	if _, err := ctx.Write(data); err != nil {
		ctx.SetStatusCode(fasthttp.StatusInternalServerError)
		ctx.Response.Header.Set("Content-Type", "application/json")
		ctx.SetBodyString(`{"error": "ERROR_WRITING_RESPONSE"}`)
		return
	}

	cacheMutex.Lock()
	cache[uuid] = cacheEntry{
		data:      data,
		timestamp: time.Now(),
	}
	cacheMutex.Unlock()

	ctx.SetStatusCode(fasthttp.StatusOK)
}
