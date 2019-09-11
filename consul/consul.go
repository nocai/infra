package consul

import (
	"fmt"
	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/go-kit/kit/sd/consul"
	"github.com/hashicorp/consul/api"
	"github.com/hashicorp/consul/api/watch"
	"github.com/nocai/infra"
	"os"
	"time"
)

func Register(l log.Logger, addr string, port int, serviceName string) consul.Client {
	consulConfig := api.DefaultConfig()
	if len(addr) > 0 {
		consulConfig.Address = addr
	}
	consulApi, err := api.NewClient(consulConfig)
	if err != nil {
		_ = level.Error(l).Log("msg", err)
		os.Exit(1)
	}

	// 服务注册
	consulClient := consul.NewClient(consulApi)
	registration := registration(port, serviceName)
	if err := consulClient.Register(registration); err != nil {
		_ = level.Error(l).Log("msg", err)
		os.Exit(1)
	}
	//defer func() {
	//	// NOTE:服务注销
	//	if err := consulClient.Deregister(registration); err != nil {
	//		_ = level.Error(l).Log("msg", err)
	//	}
	//}()

	// KV
	kv(l, consulApi, addr, serviceName)
	return consulClient
}

func registration(port int, serverName string) *api.AgentServiceRegistration {
	localIP := infra.LocalIP()
	return &api.AgentServiceRegistration{
		ID:      fmt.Sprintf("%v:%v:%d", serverName, localIP, port),
		Name:    serverName,
		Address: localIP,
		Port:    port,
		Tags:    []string{serverName, "urlprefix-/" + serverName + " strip=/" + serverName},
		Check: &api.AgentServiceCheck{
			DeregisterCriticalServiceAfter: (5 * time.Second).String(),
			HTTP:                           fmt.Sprintf("http://%s:%d/echo?Req=%v", localIP, port, serverName),
			Timeout:                        "5s",
			Interval:                       "5s",
		},
	}
}

func kv(l log.Logger, consulApi *api.Client, consulAddr, servicePath string) {
	_ = level.Info(l).Log("msg", fmt.Sprintf("servicePath:%s", servicePath))

	kvp, _, err := consulApi.KV().Get(servicePath, nil)
	if err != nil || kvp == nil {
		_ = level.Error(l).Log("msg", fmt.Sprintf("err:%v, kvp:%v", err, kvp))
		os.Exit(-1)
	}
	if err = Unmarshal(kvp.Value); err != nil {
		_ = level.Error(l).Log("msg", err)
		os.Exit(-1)
	}

	go func() {
		if r := recover(); r != nil {
			_ = level.Error(l).Log("msg", r)
		}
		kvWatch(l, consulAddr, servicePath)
	}()
}

func kvWatch(l log.Logger, consulAddr, servicePath string) {
	var (
		plan *watch.Plan
		err  error
	)

	if plan, err = watch.Parse(map[string]interface{}{"type": "key", "key": servicePath,}); err != nil {
		_ = level.Error(l).Log("msg", err)
		os.Exit(1)
	}
	plan.HybridHandler = func(_ watch.BlockingParamVal, raw interface{}) {
		_ = level.Info(l).Log("msg", "update config file")
		if kvp, ok := raw.(*api.KVPair); ok {
			if err = Unmarshal(kvp.Value); err != nil {
				_ = level.Error(l).Log("msg", err)
			}
		}
	}

	if err := plan.Run(consulAddr); err != nil {
		_ = level.Error(l).Log("msg", err)
		os.Exit(1)
	}
}
