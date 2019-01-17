package postFilters

import (
	"fmt"
	"strings"

	"github.com/Laisky/go-fluentd/libs"
	"github.com/Laisky/go-utils"
	"go.uber.org/zap"
)

type ESDispatcherFilterCfg struct {
	TagKey   string
	Tags     []string
	ReTagMap map[string]string
}

type ESDispatcherFilter struct {
	*BaseFilter
	*ESDispatcherFilterCfg
	supportedTags map[string]struct{}
}

// LoadReTagMap parse retag config
// app.spring.{env}: es-general -> {app.spring.sit: es-general}
func LoadReTagMap(env string, mapi interface{}) map[string]string {
	retagMap := map[string]string{}
	for tag, retagi := range mapi.(map[string]interface{}) {
		retagMap[strings.Replace(tag, "{env}", env, -1)] = strings.Replace(retagi.(string), "{env}", env, -1)
	}

	return retagMap
}

func NewESDispatcherFilter(cfg *ESDispatcherFilterCfg) *ESDispatcherFilter {
	utils.Logger.Info("new ESDispatcherFilter",
		zap.Strings("tags", cfg.Tags))
	f := &ESDispatcherFilter{
		BaseFilter:            &BaseFilter{},
		ESDispatcherFilterCfg: cfg,
	}

	f.supportedTags = map[string]struct{}{}
	for _, t := range f.Tags {
		f.supportedTags[t] = struct{}{}
	}

	return f
}

func (f *ESDispatcherFilter) Filter(msg *libs.FluentMsg) *libs.FluentMsg {
	var ok bool
	if _, ok = f.supportedTags[msg.Tag]; !ok {
		return msg
	}

	if msg.Message[f.TagKey].(string) == "" {
		utils.Logger.Error("msg tag should not empty", zap.String("msg", fmt.Sprint(msg)))
		f.DiscardMsg(msg)
		return nil
	}

	if msg.Tag, ok = f.ReTagMap[msg.Message[f.TagKey].(string)]; !ok {
		utils.Logger.Error("tag not exists in retagmap", zap.String("tag", msg.Message[f.TagKey].(string)))
		f.DiscardMsg(msg)
		return nil
	}

	utils.Logger.Debug("change msg tag",
		zap.String("old_tag", msg.Message[f.TagKey].(string)),
		zap.String("new_tag", msg.Tag))
	return msg
}
