package config

import (
	"fmt"
	"os"
	"regexp"
	"time"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Conf 保存配置的全局变量
var Conf = new(Config)

// Config 配置入口
type Config struct {
	*AppConfig   `mapstructure:"app"`
	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
	*Remote      `mapstructure:"remote"`
	*MetricRegex `mapstructure:"metric_filter"`
	*MetricFilter
	*MetricStore `mapstructure:"metric_store"`
}

// AppConfig 项目配置
type AppConfig struct {
	Name    string `mapstructure:"name"`
	Mode    string
	Version string `mapstructure:"version"`
	Port    int    `mapstructure:"port"`
}

// LogConfig 日志配置
type LogConfig struct {
	Level      string `mapstructure:"level"`
	Filename   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

// MySQLConfig mysql数据库配置
type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	Password     string `mapstructure:"password"`
	DBName       string `mapstructure:"db_name"`
	Port         int    `mapstructure:"port"`
	MaOpenConns  int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

// RedisConfig redis配置
type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Port     int    `mapstructure:"port"`
	Password string `mapstructure:"password"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

//MetricRegex metric过滤配置，不符合要求的metric直接丢弃
type MetricRegex struct {
	MetricNameRegex  string `mapstructure:"metricNameRegex"`
	MetricValueRegex string `mapstructure:"metricValueRegex"`
	LabelNameRegex   string `mapstructure:"labelNameRegex"`
	LabelValueRegex  string `mapstructure:"labelValueRegex"`
}

// MetricFilter 程序启动后，将MetricRegex中string初始化成*regexp.Regexp
type MetricFilter struct {
	MetricNameRegex *regexp.Regexp
	LabelNameRegex  *regexp.Regexp
	LabelValueRegex *regexp.Regexp
}

// MetricStore 监控指标名存储配置
type MetricStore struct {
	*Cache `mapstructure:"cache"`
	*Store `mapstructure:"store"`
}

// Cache 缓存配置
type Cache struct {
	IsExpire        bool          `mapstructure:"isexpire"`
	Expire          int           `mapstructure:"expire"`
	DistInterval    int           `mapstructure:"dist_interval"`
	DefaultExpire   time.Duration `mapstructure:"default_expire"`
	CleanupInterval time.Duration `mapstructure:"cleanup_interval"`
	WorkerNum       int           `mapstructure:"worker_num"`
	FlushLens       int           `mapstructure:"flush_lens"`
	FlushInterval   time.Duration `mapstructure:"flush_interval"`
}

// Store 存储配置
type Store struct {
	URL           []string      `mapstructure:"url"`
	UserName      string        `mapstructure:"username"`
	Password      string        `mapstructure:"password"`
	IndexName     string        `mapstructure:"index_name"`
	WorkerNum     int           `mapstructure:"worker_num"`
	FlushBytes    int           `mapstructure:"flush_bytes"`
	FlushInterval time.Duration `mapstructure:"flush_interval"`
	ChanSize      int           `mapstructure:"chan_size"`
}

// Remote 转发metrics到远端服务的相关配置
type Remote struct {
	*Write `mapstructure:"write"`
	*Send  `mapstructure:"send"`
}

// Write VictoriaMetrics remote write配置
type Write struct {
	URL         string `mapstructure:"url"`
	ContentType string `mapstructure:"content_type"`
}

// Send VictoriaMetrics /api/put 发送数据配置
type Send struct {
	URL         string `mapstructure:"url"`
	ContentType string `mapstructure:"content_type"`
}

// Init 初始化配置
func Init() (err error) {
	//viper.SetConfigFile("config.yaml")
	env := os.Getenv("GO_ENV")
	viper.SetConfigName(env)
	viper.SetConfigType("yml")
	viper.AddConfigPath("./config/")
	if err = viper.ReadInConfig(); err != nil {
		fmt.Println("viper.ReadInConfig() ")
	}

	// 反序列化配置到全局变量Conf中
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal failed, err: %v\n", err)
	}

	// 设置环境标志
	Conf.AppConfig.Mode = env
	// 初始化Conf.MetricFilter
	Conf.MetricFilter = new(MetricFilter)
	// 初始化filter正则
	if err = initRegexp(); err != nil {
		fmt.Printf("init filter regex fialed, err:%v\n", err)
		return
	}

	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("config has already update...")
		// 反序列化配置更新到全局变量Conf中
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal failed, err: %v\n", err)
		}
		// 初始化filter正则
		if err := initRegexp(); err != nil {
			fmt.Printf("init filter regex fialed, err:%v\n", err)
		}
	})

	return
}
