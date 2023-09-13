package config

type GoodsSrvConfig struct {
	Name string `mapstructure:"name"`
}

type ServerConfig struct {
	Name             string         `mapstructure:"name"`
	Port             int            `mapstructure:"port"`
	GoodsSrvInfo     GoodsSrvConfig `mapstructure:"goods_srv"`
	InventorySrvInfo GoodsSrvConfig `mapstructure:"inventory_srv"`
	OrderSrvInfo     GoodsSrvConfig `mapstructure:"order_srv"`
	JWTInfo          JWTConfig      `mapstructure:"jwt"`
	ConsulInfo       ConsulConfig   `mapstructure:"consul"`
	Tags             []string       `mapstructure:"tags" json:"tags"`
}

type JWTConfig struct {
	SigningKey string `mapstructure:"key"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host"`
	Port int    `mapstructure:"port"`
}
