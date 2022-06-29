package main

import (
	"github.com/spf13/viper"
	"log"
)

var (
	ServerSetting   *ServerSettingS
	DatabaseSetting *DatabaseSettingS
)

// global settings
var sections = make(map[string]interface{})

// ServerSettingS is the server config
type ServerSettingS struct {
	port int
}

// DatabaseSettingS is the database config
type DatabaseSettingS struct {
	driverName     string
	dataSourceName string
	dbName         string
}

type Setting struct {
	vp *viper.Viper
}

// NewSetting is the constructor for Setting.
func NewSetting() (*Setting, error) {
	vp := viper.New()
	vp.SetConfigName("conf")
	vp.AddConfigPath("conf")
	vp.SetConfigType("yml")
	err := vp.ReadInConfig()
	if err != nil {
		return nil, err
	}
	s := &Setting{vp: vp}
	return s, nil
}

func (s *Setting) ReadSection(k string) *viper.Viper {
	v := s.vp.Sub(k)
	// if not existed, just set it
	if _, ok := sections[k]; !ok {
		sections[k] = v
	}
	return v
}

func (s *Setting) ReloadAllSection() error {
	for k, _ := range sections {
		s.ReadSection(k)
	}
	return nil
}

func SetupSetting() error {
	s, err := NewSetting()
	if err != nil {
		return err
	}
	dbData := s.ReadSection("database")
	if dbData == nil {
		panic("database configuration not found")
	}
	DatabaseSetting = NewDatabaseSetting(dbData)
	serverData := s.ReadSection("server")
	if serverData == nil {
		panic("server configuration not found")
	}
	ServerSetting = NewServerSetting(serverData)
	log.Printf("DatabaseSetting: %v\n", DatabaseSetting)
	log.Printf("ServerSetting: %v\n", ServerSetting)
	return nil
}

func NewDatabaseSetting(v *viper.Viper) *DatabaseSettingS {
	return &DatabaseSettingS{
		driverName:     v.GetString("driverName"),
		dataSourceName: v.GetString("dataSourceName"),
		dbName:         v.GetString("dbName"),
	}
}

func NewServerSetting(v *viper.Viper) *ServerSettingS {
	return &ServerSettingS{
		port: v.GetInt("port"),
	}
}
