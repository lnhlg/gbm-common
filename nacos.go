package common

import (
	nacosconfig "github.com/go-kratos/kratos/contrib/config/nacos/v2"
	"github.com/go-kratos/kratos/contrib/registry/nacos/v2"
	kconfig "github.com/go-kratos/kratos/v2/config"
	"github.com/nacos-group/nacos-sdk-go/clients"
	"github.com/nacos-group/nacos-sdk-go/common/constant"
	"github.com/nacos-group/nacos-sdk-go/vo"
)

type NacosCfgSource struct {
	// cfg *config
	sc       []constant.ServerConfig
	userName string
	password string
}

func NewNacosCfgSource(cfgFile string) (*NacosCfgSource, error) {
	cfg, err := NewConfig(cfgFile)
	if err != nil {
		return nil, err
	}
	sc := make([]constant.ServerConfig, len(cfg.Nacos.Servers))
	for i := range cfg.Nacos.Servers {
		sc[i] = *constant.NewServerConfig(
			cfg.Nacos.Servers[i].Host,
			cfg.Nacos.Servers[i].Port,
		)
	}

	return &NacosCfgSource{
		sc:       sc,
		userName: cfg.Nacos.UserName,
		password: cfg.Nacos.Password,
	}, nil
}

func (nfs *NacosCfgSource) NacosSource(namespaceid, dataid, group string) (kconfig.Source, error) {

	// sc := make([]constant.ServerConfig, len(nfs.cfg.Nacos))
	// for i := range nfs.cfg.Nacos {
	// 	sc[i] = *constant.NewServerConfig(nfs.cfg.Nacos[i].Host, nfs.cfg.Nacos[i].Port)
	// }

	cc := &constant.ClientConfig{
		NamespaceId:         namespaceid, //namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "/tmp/nacos/log",
		CacheDir:            "/tmp/nacos/cache",
		LogLevel:            "debug",
		Username:            nfs.userName,
		Password:            nfs.password,
	}

	// a more graceful way to create naming client
	client, err := clients.NewConfigClient(
		vo.NacosClientParam{
			ClientConfig:  cc,
			ServerConfigs: nfs.sc,
		},
	)
	if err != nil {

		return nil, err
	}

	source := nacosconfig.NewConfigSource(client,
		nacosconfig.WithDataID(dataid),
		nacosconfig.WithGroup(group),
	)

	return source, nil
}

func (nfs *NacosCfgSource) NacosNaming(NamespaceId string) (*nacos.Registry, error) {

	client, err := clients.NewNamingClient(
		vo.NacosClientParam{
			ServerConfigs: nfs.sc,
			ClientConfig: &constant.ClientConfig{
				NamespaceId: NamespaceId,
				TimeoutMs:   5000,
				Username:    nfs.userName,
				Password:    nfs.password,
			},
		},
	)

	if err != nil {

		return nil, err
	}

	r := nacos.New(client)

	return r, nil
}
