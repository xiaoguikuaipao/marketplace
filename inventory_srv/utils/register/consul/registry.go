package consul

import (
	"fmt"
	"strconv"

	"github.com/hashicorp/consul/api"
)

type register struct {
	Host string
	Port int
}

type RegisterClient interface {
	Register(address string, port int, tags []string, id int, name string) error
	DeRegister(id string) error
}

func NewRegistryClient(address string, port int) RegisterClient {
	return &register{
		Host: address,
		Port: port,
	}
}

var consulClient *api.Client

func (r *register) Register(address string, port int, tags []string, id int, name string) error {
	//服务注册
	cfg := api.DefaultConfig()
	cfg.Address = fmt.Sprintf("%s:%d",
		r.Host,
		r.Port)
	var err error
	consulClient, err = api.NewClient(cfg)
	if err != nil {
		panic(err)
	}
	//生成检查对象
	check := &api.AgentServiceCheck{
		Interval:                       "5s",
		Timeout:                        "5s",
		GRPC:                           fmt.Sprintf("%s:%d", address, port),
		DeregisterCriticalServiceAfter: "10s",
	}
	//生成注册对象
	registration := new(api.AgentServiceRegistration)
	registration.Name = name
	registration.ID = strconv.Itoa(id)
	registration.Port = port
	registration.Tags = tags
	registration.Address = address
	registration.Check = check

	err = consulClient.Agent().ServiceRegister(registration)
	if err != nil {
		panic(err)
	}
	return err
}
func (r *register) DeRegister(id string) error {
	if err := consulClient.Agent().ServiceDeregister(id); err != nil {
		return err
	}
	return nil
}
