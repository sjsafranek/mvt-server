package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
	"github.com/sjsafranek/goutils"
)

const (
	DEFAULT_DATABASE_ENGINE   = "postgres"
	DEFAULT_DATABASE_DATABASE = "geodev"
	DEFAULT_DATABASE_PASSWORD = "dev"
	DEFAULT_DATABASE_USERNAME = "geodev"
	DEFAULT_DATABASE_HOST     = "localhost"
)

type Config struct {
	Title    string         `toml:"title"`
	Server   ServerConfig   `toml:"server"`
	Database DatabaseConfig `toml:"database"`
}

type ServerConfig struct {
	Port   int    `toml:"port"`
	Secret string `toml:"secret"`
}

type DatabaseConfig struct {
	Type     string `toml:"type"`
	Database string `toml:"database"`
	Password string `toml:"password"`
	Username string `toml:"username"`
	Host     string `toml:"host"`
}

func (self *Config) UseDefaults() error {
	self.Title = "MVT-Server"
	self.Server.Port = DEFAULT_PORT
	self.Server.Secret = utils.RandomString(10)
	self.Database.Type = DEFAULT_DATABASE_ENGINE
	self.Database.Database = DEFAULT_DATABASE_DATABASE
	self.Database.Password = DEFAULT_DATABASE_PASSWORD
	self.Database.Username = DEFAULT_DATABASE_USERNAME
	self.Database.Host = DEFAULT_DATABASE_HOST
	return self.Save("config.toml")
}

func (self *Config) Fetch(file string) error {
	b, err := ioutil.ReadFile(file)
	if nil != err {
		return err
	}
	return self.Unmarshal(string(b))
}

func (self *Config) Save(file string) error {
	contents, err := self.Marshal()
	if nil != err {
		return err
	}
	return ioutil.WriteFile(file, []byte(contents), 0644)
}

func (self *Config) Unmarshal(data string) error {
	return toml.Unmarshal([]byte(data), self)
}

func (self Config) Marshal() (string, error) {
	b, err := toml.Marshal(self)
	if nil != err {
		return "", err
	}
	return string(b), nil
}

func (self *DatabaseConfig) ConnectionString() string {
	return fmt.Sprintf("%v://%v:%v@%v/%v?sslmode=disable", self.Type, self.Username, self.Password, self.Host, self.Database)
}
