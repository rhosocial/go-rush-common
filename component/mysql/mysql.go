package mysql

import "fmt"

type EnvMySQLServer struct {
	Host     string `yaml:"Host" default:"localhost"`
	Port     uint16 `yaml:"Port" default:"3306"`
	Username string `yaml:"Username" default:"root"`
	Password string `yaml:"Password" default:"123456"`
	DB       string `yaml:"DB" default:"node"`
	Charset  string `yaml:"charset" default:"utf8mb4"`
	Location string `yaml:"loc" default:"Local"`
}

func (e EnvMySQLServer) GetDSN() string {
	return fmt.Sprintf("%s:%s@tcp(%s:%d)/%s?charset=%s&parseTime=true&loc=%s", e.Username, e.Password, e.Host, e.Port, e.DB, e.Charset, e.Location)
}
