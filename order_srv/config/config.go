package config

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"db" json:"db"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

type ServerConfig struct {
	Name       string       `mapstructure:"name" json:"name"`
	MysqlInfo  MysqlConfig  `mapstructure:"mysql" json:"mysql"`
	ConsuIInfo ConsulConfig `mapstructure:"consul" json:"consul"`
	Tags       []string     `mapstructure:"tags" json:"tags"`

	GoodsSrvInfo     GoodsSrvConfig     `mapstructure:"goods_srv" json:"goods_srv"`
	InventorySrvInfo InventorySrvConfig `mapstructure:"inventory_srv" json:"inventory_srv"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}

type GoodsSrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}

type InventorySrvConfig struct {
	Name string `mapstructure:"name" json:"name"`
}
