package api

import (
	"encoding/json"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"go-fitness/external/logger/sl"
	"golang.org/x/text/language"
	"log/slog"
)

func NewBundle(log *slog.Logger) *i18n.Bundle {
	bundle := i18n.NewBundle(language.Romanian)
	bundle.RegisterUnmarshalFunc("json", json.Unmarshal)

	_, err := bundle.LoadMessageFile("/app/lang/active.ro.json")
	if err != nil {
		log.Error("failed to load active.ro.json", sl.Err(err))
		panic(err)
	}

	return bundle
}

func NewLocalizer(log *slog.Logger, bundle *i18n.Bundle) *i18n.Localizer {
	return i18n.NewLocalizer(bundle, "ro")
}
