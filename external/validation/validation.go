package validation

import (
	"fmt"
	"github.com/go-playground/validator/v10"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"strings"
)

func ValidationError(localizer *i18n.Localizer, errs validator.ValidationErrors) error {
	var errMsgs []string
	for _, err := range errs {
		var msg string
		switch err.ActualTag() {
		case "required":
			msg = localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "required_field",
				TemplateData: map[string]string{
					"Field": err.Field(),
				},
			})
		case "url":
			msg = localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "invalid_url",
				TemplateData: map[string]string{
					"Field": err.Field(),
				},
			})
		case "string":
			msg = localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "must_be_string",
				TemplateData: map[string]string{
					"Field": err.Field(),
				},
			})
		case "float64":
			msg = localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "must_be_float",
				TemplateData: map[string]string{
					"Field": err.Field(),
				},
			})
		case "int":
			msg = localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "must_be_int",
				TemplateData: map[string]string{
					"Field": err.Field(),
				},
			})
		case "min":
			msg = localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "min_value",
				TemplateData: map[string]string{
					"Field": err.Field(),
					"Param": err.Param(),
				},
			})
		case "max":
			msg = localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "max_value",
				TemplateData: map[string]string{
					"Field": err.Field(),
					"Param": err.Param(),
				},
			})
		default:
			msg = localizer.MustLocalize(&i18n.LocalizeConfig{
				MessageID: "invalid_field",
				TemplateData: map[string]string{
					"Field": err.Field(),
				},
			})
		}
		errMsgs = append(errMsgs, msg)
	}

	return fmt.Errorf(strings.Join(errMsgs, ", "))
}
