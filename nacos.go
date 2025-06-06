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
	sc       []constant.ServerConfig
	userName string
	password string
}

type NacosHost struct {
	Host string
	Port uint64
}

// 修改 NewNacosCfgSource 函数，直接接收服务器配置、用户名和密码
func NewNacosCfgSource(
	svs []*NacosHost,
	userName,
	password string,
) *NacosCfgSource {
	sc := make([]constant.ServerConfig, len(svs))
	for i, s := range svs {
		sc[i] = constant.ServerConfig{
			IpAddr: s.Host,
			Port:   s.Port,
		}
	}
	return &NacosCfgSource{
		sc:       sc,
		userName: userName,
		password: password,
	}
}

func (nfs *NacosCfgSource) NacosSource(namespaceid, dataid, group string) (kconfig.Source, error) {
	cc := &constant.ClientConfig{
		NamespaceId:         namespaceid, //namespace id
		TimeoutMs:           5000,
		NotLoadCacheAtStart: true,
		LogDir:              "./tmp/nacos/log",
		CacheDir:            "./tmp/nacos/cache",
		LogLevel:            "debug",
		Username:            nfs.userName,
		Password:            nfs.password,
	}

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
