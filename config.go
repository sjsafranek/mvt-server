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

var (
	DATABASE_ENGINE   = DEFAULT_DATABASE_ENGINE
	DATABASE_DATABASE = DEFAULT_DATABASE_DATABASE
	DATABASE_PASSWORD = DEFAULT_DATABASE_PASSWORD
	DATABASE_USERNAME = DEFAULT_DATABASE_USERNAME
	DATABASE_HOST     = DEFAULT_DATABASE_HOST
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
	Port     int64  `toml:"port"`
}

func (self *Config) UseDefaults() error {
	self.Title = "MVT-Server"
	self.Server.Port = DEFAULT_PORT
	self.Server.Secret = utils.RandomString(10)
	self.Database.Type = DATABASE_ENGINE
	self.Database.Database = DATABASE_DATABASE
	self.Database.Password = DATABASE_PASSWORD
	self.Database.Username = DATABASE_USERNAME
	self.Database.Host = DATABASE_HOST
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
	return fmt.Sprintf("%v://%v:%v@%v:%v/%v?sslmode=disable", self.Type, self.Username, self.Password, self.Host, self.Port, self.Database)
}
