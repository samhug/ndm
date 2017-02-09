package config

type Preferences struct {
	BackupDir string `mapstructure:"backup_dir,"`
	HostIP string `mapstructure:"host_ip,"`
}

type DeviceClass struct {
	Name   string `mapstructure:",key"`
	Script string `mapstructure:"script,"`
}

type Auth struct {
	Name     string `mapstructure:",key"`
	Username string `mapstructure:"username,"`
	Password string `mapstructure:"password,"`
}

type Device struct {
	Name  string
	Class *DeviceClass
	Addr  string
	Auth  *Auth
}

type Config struct {
	Preferences   *Preferences
	DeviceClasses map[string]*DeviceClass
	Auths         map[string]*Auth
	Devices       map[string]*Device
}
