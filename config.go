package main

var config Config

type Config struct {
	Username string    `yaml:"username"`
	Token    string    `yaml:"token"`
	Settings Settings  `yaml:"settings"`
	Projects []Project `yaml:"projects"`
}
type Project struct {
	Url    string `yaml:"url"`
	Path   string `yaml:"path"`
	Branch string `yaml:"branch"`
	Check  string `yaml:"check"`
	Script string `yaml:"script"`
}

type Settings struct {
	CheckTime int `yaml:"check_time"`
}
