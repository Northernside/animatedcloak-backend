package main

import (
	"animatedcloak-backend/http"
	"animatedcloak-backend/modules/env"
)

func main() {
	env.LoadEnvFile()

	http.StartAPI()
}
