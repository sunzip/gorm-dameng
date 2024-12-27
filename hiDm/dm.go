package hiDm

import (
	"fmt"
	"time"

	dameng "github.com/godoes/gorm-dameng"
	"go.opentelemetry.io/otel/trace"
	"gorm.io/gorm"
)

// New new mysql
func New(user string, pwd string, addr string, port int, connectTimeout int, schema string, tracer trace.TracerProvider, options map[string]string, cfg *gorm.Config) (*gorm.DB, error) {
	if options == nil {
		options = make(map[string]string)
	}
	options["schema"] = schema
	options["appName"] = "GORM 连接达梦数据库"
	if connectTimeout <= 0 {
		connectTimeout = 30000
	}
	options["connectTimeout"] = fmt.Sprintf("%d", connectTimeout)

	// dm://user:password@host:port?schema=SYSDBA[&...]
	dsn := dameng.BuildUrl(user, pwd, addr, port, options)

	db, err := gorm.Open(dameng.Open(dsn), cfg)
	if err != nil {
		return nil, err
	}

	instance, err := db.DB()
	if err != nil {
		return nil, err
	}
	instance.SetMaxIdleConns(5)
	instance.SetMaxOpenConns(50)
	instance.SetConnMaxLifetime(time.Hour)
	// if tracer != nil {
	// 	_ = db.Use(plugins.NewPlugin(plugins.WithTracerProvider(tracer)))
	// }
	// _ = db.Use(plugins.NewPrometheus(plugins.PromMetadata{Addr: addr}))

	return db, nil
}
