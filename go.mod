module monica-adaptor

go 1.15

require (
	github.com/fsnotify/fsnotify v1.4.9
	github.com/gin-contrib/pprof v1.3.0
	github.com/gin-gonic/gin v1.7.2
	github.com/go-playground/validator/v10 v10.6.1 // indirect
	github.com/go-redis/redis v6.15.9+incompatible
	github.com/go-sql-driver/mysql v1.5.0
	github.com/golang/protobuf v1.5.2
	github.com/golang/snappy v0.0.3
	github.com/jmoiron/sqlx v1.3.4
	github.com/json-iterator/go v1.1.11
	github.com/natefinch/lumberjack v2.0.0+incompatible
	github.com/olivere/elastic/v7 v7.0.25
	github.com/onsi/ginkgo v1.16.4 // indirect
	github.com/onsi/gomega v1.13.0 // indirect
	github.com/pkg/errors v0.9.1
	github.com/prometheus/prometheus v2.5.0+incompatible
	github.com/spf13/viper v1.8.1
	go.uber.org/zap v1.17.0
	google.golang.org/protobuf v1.27.1 // indirect
	gopkg.in/natefinch/lumberjack.v2 v2.0.0 // indirect
)

replace github.com/olivere/elastic/v7 v7.0.25 => ./pkg/elastic
