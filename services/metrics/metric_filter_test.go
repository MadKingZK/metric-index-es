package metrics

import (
	"fmt"
	"metric-index/config"
	"os"
	"path/filepath"
	"testing"
)

func TestMetricFilter(t *testing.T) {
	home, _ := os.UserHomeDir()
	err := os.Chdir(filepath.Join(home, "go", "metric-index"))
	if err != nil {
		panic(err)
	}

	if err := config.Init(); err != nil {
		fmt.Printf("init settings failed, err:%v\n", err)
		return
	}

	fmt.Println(config.Conf.MetricRegex)

	// 初始化filter正则
	if err := config.InitRegexp(); err != nil {
		fmt.Printf("init filter regex fialed, err:%v\n", err)
		return
	}

	cfg := config.Conf.MetricFilter
	fmt.Println("------>", cfg)

	metric := `namegroup_context_switches_total{ctxswitchtype="voluntary",groupname="anacron"} 20`
	matched := cfg.MetricNameRegex.MatchString(metric)
	fmt.Println(cfg.MetricNameRegex)

	fmt.Println(matched)
}
