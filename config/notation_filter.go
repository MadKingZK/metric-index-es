package config

import (
	"errors"
	"regexp"
)

// InitRegexp 初始化正则配置
func initRegexp() (err error) {
	MetricNameRegexTemp := regexp.MustCompile(Conf.MetricRegex.MetricNameRegex)
	if MetricNameRegexTemp == nil {
		return errors.New("init regexp failed, parse MetricNameRegex failed")
	}
	Conf.MetricFilter.MetricNameRegex = MetricNameRegexTemp

	LabelNameRegexTemp := regexp.MustCompile(Conf.MetricRegex.LabelNameRegex)
	if LabelNameRegexTemp == nil {
		return errors.New("init regexp failed, parse LabelNameRegex failed")
	}
	Conf.MetricFilter.LabelNameRegex = LabelNameRegexTemp

	LabelValueRegexTemp := regexp.MustCompile(Conf.MetricRegex.LabelValueRegex)
	if LabelValueRegexTemp == nil {
		return errors.New("init regexp failed, parse LabelValueRegex failed")
	}
	Conf.MetricFilter.LabelValueRegex = LabelValueRegexTemp

	return
}
