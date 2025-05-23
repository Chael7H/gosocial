package settings

import (
	"fmt"

	"github.com/fsnotify/fsnotify"
	"github.com/spf13/viper"
)

// Conf 全局变量，用来保存程序的所有配置信息
var Conf = new(AppConfig)

type AppConfig struct {
	Name         string `mapstructure:"name"`
	Mode         string `mapstructure:"mode"`
	Version      string `mapstructure:"version"`
	StartTime    string `mapstructure:"start_time"`
	MachineID    uint16 `mapstructure:"machine_id"`
	Port         int    `mapstructure:"port"`
	*LogConfig   `mapstructure:"log"`
	*MySQLConfig `mapstructure:"mysql"`
	*RedisConfig `mapstructure:"redis"`
}

type LogConfig struct {
	Level      string `mapstructure:"level"`
	FileName   string `mapstructure:"filename"`
	MaxSize    int    `mapstructure:"max_size"`
	MaxAge     int    `mapstructure:"max_age"`
	MaxBackups int    `mapstructure:"max_backups"`
}

type MySQLConfig struct {
	Host         string `mapstructure:"host"`
	User         string `mapstructure:"user"`
	PassWord     string `mapstructure:"password"`
	Dbname       string `mapstructure:"dbname"`
	Port         int    `mapstructure:"port"`
	MaxOpenConns int    `mapstructure:"max_open_conns"`
	MaxIdleConns int    `mapstructure:"max_idle_conns"`
}

type RedisConfig struct {
	Host     string `mapstructure:"host"`
	Password string `mapstructure:"password"`
	Port     int    `mapstructure:"port"`
	DB       int    `mapstructure:"db"`
	PoolSize int    `mapstructure:"pool_size"`
}

func Init() (err error) {
	//方式1：直接指定文件路径(相对路径或者绝对路径)
	//viper.SetConfigFile("./conf/config.yaml") // ---相对路径，一般项目使用较多
	//方式2，指定配置文件名和配置文件的位置，viper可自行查找可用的配置文件
	//配置文件名不需要带后缀

	viper.SetConfigName("config") //指定配置文件名称(不需要带后缀)
	//viper.SetConfigType("yaml")   //指定配置文件类型(专用于从远程获取配置信息时指定配置文件类型的)
	//配置文件可配置多个
	viper.AddConfigPath(".")      //指定查找配置文件的路径(相对路径)
	viper.AddConfigPath("./conf") //指定查找配置文件的路径(相对路径)
	err = viper.ReadInConfig()    //读取配置信息
	if err != nil {
		//读取配置信息失败
		fmt.Printf("viper.ReadInConfig() failed, err: %s\n", err)
		return
	}
	//把读取到的配置信息反序列化到Conf中
	if err := viper.Unmarshal(Conf); err != nil {
		fmt.Printf("viper.Unmarshal() failed, err: %v\n", err)
	}
	//配置热加载
	viper.WatchConfig()
	viper.OnConfigChange(func(in fsnotify.Event) {
		fmt.Println("配置文件修改了... ")
		//配置更新后，把读取到的最新配置信息反序列化到Conf中
		if err := viper.Unmarshal(Conf); err != nil {
			fmt.Printf("viper.Unmarshal() failed, err: %s\n", err)
		}
	})
	return
}
