package config

type MysqlConfig struct {
	Host     string `mapstructure:"host" json:"host"`
	Port     int    `mapstructure:"port" json:"port"`
	Name     string `mapstructure:"db" json:"db"`
	User     string `mapstructure:"user" json:"user"`
	Password string `mapstructure:"password" json:"password"`
}

type ServerConfig struct {
	Name      string       `mapstructure:"name" json:"name"`
	MysqlInfo MysqlConfig  `mapstructure:"mysql" json:"mysql" mapstructure:"mysql_info" json:"mysql_info"`
	ConsuInfo ConsulConfig `mapstructure:"consul" json:"consul" mapstructure:"consu_info" json:"consu_info"`
}

type ConsulConfig struct {
	Host string `mapstructure:"host" json:"host"`
	Port int    `mapstructure:"port" json:"port"`
}
