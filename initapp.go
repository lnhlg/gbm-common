package common

import (
	"os"

	"github.com/go-kratos/kratos/v2/config"
	"github.com/go-kratos/kratos/v2/config/file"
	"github.com/go-kratos/kratos/v2/log"
	"github.com/go-kratos/kratos/v2/middleware/tracing"
	"github.com/go-kratos/kratos/v2/registry"

	"gopkg.in/yaml.v3"
)

type app struct {
	id             string
	name           string
	version        string
	nacosNamespace string
	nacosCfg       *NacosCfgSource
}

type appResult struct {
	Reg     registry.Registrar
	Logger  log.Logger
	Metrics *Metrics
	Cfg     config.Config
}

func NewApp(
	id, name, version, nacosNamespace string,
	nacosCfg *NacosCfgSource,
) *app {
	return &app{
		id:             id,
		name:           name,
		version:        version,
		nacosNamespace: nacosNamespace,
		nacosCfg:       nacosCfg,
	}
}

func (a *app) Init(
	confPath string,
	s ...config.Source,
) (*appResult, error) {
	logger := log.With(log.NewStdLogger(os.Stdout),
		"ts", log.DefaultTimestamp,
		"caller", log.DefaultCaller,
		"service.id", a.id,
		"service.name", a.name,
		"service.version", a.version,

		"trace.id", tracing.TraceID(),
		"span.id", tracing.SpanID(),
	)

	reg, err := a.nacosCfg.NacosNaming(a.nacosNamespace)
	if err != nil {
		return nil, err
	}

	s = append(s, file.NewSource(confPath))
	c := config.New(
		config.WithSource(
			s...,
		),
		config.WithDecoder(func(kv *config.KeyValue, v map[string]interface{}) error {
			return yaml.Unmarshal(kv.Value, v)
		}),
	)
	defer c.Close()

	if err := c.Load(); err != nil {
		return nil, err
	}

	gbmMetrics, err := NewMetrics(a.name)
	if err != nil {
		return nil, err
	}

	return &appResult{
		Reg:     reg,
		Logger:  logger,
		Metrics: gbmMetrics,
		Cfg:     c,
	}, nil
}
