package main

import (
	"go-fitness/internal/api"
)

func main() {
	fx := api.NewApp()
	fx.Run()
}
