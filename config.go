package main

import (
	"fmt"
	"io/ioutil"

	"github.com/pelletier/go-toml"
)

type Config struct {
	Title    string         `toml:"title"`
	Database DatabaseConfig `toml:"database"`
}

type DatabaseConfig struct {
	Type     string `toml:"Type"`
	Database string `toml:"Database"`
	Password string `toml:"Password"`
	Username string `toml:"Username"`
	Host     string `toml:"Host"`
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
	return fmt.Sprintf("%v://%v:%v@%v/%v", self.Type, self.Username, self.Password, self.Host, self.Database)
}
