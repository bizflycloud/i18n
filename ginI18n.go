package i18n

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"github.com/nicksnyder/go-i18n/v2/i18n"
	"golang.org/x/text/language"
)

var _ GinI18n = (*GinI18nImpl)(nil)

type GinI18nImpl struct {
	bundle          *i18n.Bundle
	currentContext  *gin.Context
	localizerByLng  map[string]*i18n.Localizer
	defaultLanguage language.Tag
	getLngHandler   GetLngHandler
}

// getMessage get localize message by lng and messageID
func (i *GinI18nImpl) GetMessage(param interface{}) (string, error) {
	lng := i.getLngHandler(i.currentContext, i.defaultLanguage.String())
	localizer := i.getLocalizerByLng(lng)

	localizeConfig, err := i.getLocalizeConfig(param)
	if err != nil {
		return "", err
	}

	message, err := localizer.Localize(localizeConfig)
	if err != nil {
		return "", err
	}

	return message, nil
}

// mustGetMessage ...
func (i *GinI18nImpl) MustGetMessage(param interface{}) string {
	message, _ := i.GetMessage(param)
	return message
}

func (i *GinI18nImpl) SetCurrentContext(ctx context.Context) {
	i.currentContext = ctx.(*gin.Context)
}

func (i *GinI18nImpl) SetBundle(cfg *BundleCfg) {
	bundle := i18n.NewBundle(cfg.DefaultLanguage)
	bundle.RegisterUnmarshalFunc(cfg.FormatBundleFile, cfg.UnmarshalFunc)

	i.bundle = bundle
	i.defaultLanguage = cfg.DefaultLanguage

	i.loadMessageFiles(cfg)
	i.setLocalizerByLng(cfg.AcceptLanguage)
}

func (i *GinI18nImpl) SetGetLngHandler(handler GetLngHandler) {
	i.getLngHandler = handler
}

// loadMessageFiles load all file localize to bundle
func (i *GinI18nImpl) loadMessageFiles(config *BundleCfg) {
	for _, lng := range config.AcceptLanguage {
		path := filepath.Join(config.RootPath, lng.String()) + "." + config.FormatBundleFile
		if err := i.loadMessageFile(config, path); err != nil {
			panic(err)
		}
	}
}

func (i *GinI18nImpl) loadMessageFile(config *BundleCfg, path string) error {
	buf, err := config.Loader.LoadMessage(path)
	if err != nil {
		return err
	}

	if _, err = i.bundle.ParseMessageFileBytes(buf, path); err != nil {
		return err
	}
	return nil
}

// setLocalizerByLng set localizer by language
func (i *GinI18nImpl) setLocalizerByLng(acceptLanguage []language.Tag) {
	i.localizerByLng = map[string]*i18n.Localizer{}
	for _, lng := range acceptLanguage {
		lngStr := lng.String()
		i.localizerByLng[lngStr] = i.newLocalizer(lngStr)
	}

	// set defaultLanguage if it isn't exist
	defaultLng := i.defaultLanguage.String()
	if _, hasDefaultLng := i.localizerByLng[defaultLng]; !hasDefaultLng {
		i.localizerByLng[defaultLng] = i.newLocalizer(defaultLng)
	}
}

// newLocalizer create a localizer by language
func (i *GinI18nImpl) newLocalizer(lng string) *i18n.Localizer {
	lngDefault := i.defaultLanguage.String()
	lngs := []string{
		lng,
	}

	if lng != lngDefault {
		lngs = append(lngs, lngDefault)
	}

	localizer := i18n.NewLocalizer(
		i.bundle,
		lngs...,
	)
	return localizer
}

// getLocalizerByLng get localizer by language
func (i *GinI18nImpl) getLocalizerByLng(lng string) *i18n.Localizer {
	localizer, hasValue := i.localizerByLng[lng]
	if hasValue {
		return localizer
	}

	return i.localizerByLng[i.defaultLanguage.String()]
}

func (i *GinI18nImpl) getLocalizeConfig(param interface{}) (*i18n.LocalizeConfig, error) {
	switch paramValue := param.(type) {
	case string:
		localizeConfig := &i18n.LocalizeConfig{
			MessageID: paramValue,
		}
		return localizeConfig, nil
	case *i18n.LocalizeConfig:
		return paramValue, nil
	}

	msg := fmt.Sprintf("un supported localize param: %v", param)
	return nil, errors.New(msg)
}
