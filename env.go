package yiigo

import (
	"errors"
	"os"
	"path/filepath"
	"time"

	"github.com/pelletier/go-toml"
	"go.uber.org/zap"
)

// Environment is the interface for config
type Environment interface {
	// Get returns an env value
	Get(key string) EnvValue
}

// EnvValue is the interface for config value
type EnvValue interface {
	// Int returns a value of int64.
	Int(defaultValue ...int64) int64

	// Ints returns a value of []int64.
	Ints(defaultValue ...int64) []int64

	// Float returns a value of float64.
	Float(defaultValue ...float64) float64

	// Floats returns a value of []float64.
	Floats(defaultValue ...float64) []float64

	// String returns a value of string.
	String(defaultValue ...string) string

	// Strings returns a value of []string.
	Strings(defaultValue ...string) []string

	// Bool returns a value of bool.
	Bool(defaultValue ...bool) bool

	// Time returns a value of time.Time.
	// Layout is required when the env value is a string.
	Time(layout string, defaultValue ...time.Time) time.Time

	// Map returns a value of X.
	Map() X

	// Unmarshal attempts to unmarshal the value into a Go struct pointed by dest.
	Unmarshal(dest interface{}) error
}

type config struct {
	tree *toml.Tree
}

func (c *config) Get(key string) EnvValue {
	return &cfgValue{value: c.tree.Get(key)}
}

type cfgValue struct {
	value interface{}
}

func (c *cfgValue) Int(defaultValue ...int64) int64 {
	var dv int64

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	result, ok := c.value.(int64)

	if !ok {
		return 0
	}

	return result
}

func (c *cfgValue) Ints(defaultValue ...int64) []int64 {
	if c.value == nil {
		return defaultValue
	}

	arr, ok := c.value.([]interface{})

	if !ok {
		return []int64{}
	}

	l := len(arr)

	result := make([]int64, 0, l)

	for _, v := range arr {
		if i, ok := v.(int64); ok {
			result = append(result, i)
		}
	}

	if len(result) < l {
		return []int64{}
	}

	return result
}

func (c *cfgValue) Float(defaultValue ...float64) float64 {
	var dv float64

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	result, ok := c.value.(float64)

	if !ok {
		return 0
	}

	return result
}

func (c *cfgValue) Floats(defaultValue ...float64) []float64 {
	if c.value == nil {
		return defaultValue
	}

	arr, ok := c.value.([]interface{})

	if !ok {
		return []float64{}
	}

	l := len(arr)

	result := make([]float64, 0, l)

	for _, v := range arr {
		if f, ok := v.(float64); ok {
			result = append(result, f)
		}
	}

	if len(result) < l {
		return []float64{}
	}

	return result
}

func (c *cfgValue) String(defaultValue ...string) string {
	dv := ""

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	result, ok := c.value.(string)

	if !ok {
		return ""
	}

	return result
}

func (c *cfgValue) Strings(defaultValue ...string) []string {
	if c.value == nil {
		return defaultValue
	}

	arr, ok := c.value.([]interface{})

	if !ok {
		return []string{}
	}

	l := len(arr)

	result := make([]string, 0, l)

	for _, v := range arr {
		if s, ok := v.(string); ok {
			result = append(result, s)
		}
	}

	if len(result) < l {
		return []string{}
	}

	return result
}

func (c *cfgValue) Bool(defaultValue ...bool) bool {
	var dv bool

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	result, ok := c.value.(bool)

	if !ok {
		return false
	}

	return result
}

func (c *cfgValue) Time(layout string, defaultValue ...time.Time) time.Time {
	var dv time.Time

	if len(defaultValue) != 0 {
		dv = defaultValue[0]
	}

	if c.value == nil {
		return dv
	}

	var result time.Time

	switch t := c.value.(type) {
	case time.Time:
		result = t
	case string:
		result, _ = time.Parse(layout, t)
	}

	return result
}

func (c *cfgValue) Map() X {
	if c.value == nil {
		return X{}
	}

	v, ok := c.value.(*toml.Tree)

	if !ok {
		return X{}
	}

	return v.ToMap()
}

func (c *cfgValue) Unmarshal(dest interface{}) error {
	if c.value == nil {
		return nil
	}

	v, ok := c.value.(*toml.Tree)

	if !ok {
		return errors.New("yiigo: invalid env value, expects *toml.Tree")
	}

	return v.Unmarshal(dest)
}

var env Environment

func initEnv() {
	if err := LoadEnvFromFile("yiigo.toml"); err != nil {
		logger.Panic("yiigo: load config file error", zap.Error(err))
	}
}

// LoadEnvFromFile load env from file
func LoadEnvFromFile(path string) error {
	path, err := filepath.Abs(path)

	if err != nil {
		return err
	}

	if _, err := os.Stat(path); err != nil {
		if os.IsNotExist(err) {
			if f, err := os.Create(path); err == nil {
				f.WriteString(defaultEnvContent)
				f.Close()
			}
		} else if os.IsPermission(err) {
			os.Chmod(path, os.ModePerm)
		}
	}

	t, err := toml.LoadFile(path)

	if err != nil {
		return err
	}

	env = &config{tree: t}

	return nil
}

// LoadEnvFromBytes load env from bytes
func LoadEnvFromBytes(b []byte) error {
	t, err := toml.LoadBytes(b)

	if err != nil {
		return err
	}

	env = &config{tree: t}

	return nil
}

// Env returns an env value
func Env(key string) EnvValue {
	return env.Get(key)
}

var defaultEnvContent = `[app]
env = "dev"
debug = true

[db]

    # [db.default]
    # driver = "mysql"
    # dsn = "username:password@tcp(localhost:3306)/dbname?timeout=10s&charset=utf8mb4&collation=utf8mb4_general_ci&parseTime=True&loc=Local"
    # max_open_conns = 20
	# max_idle_conns = 10
	# conn_max_idle_time = 60
    # conn_max_lifetime = 600

[mongo]

	# [mongo.default]
	# dsn = "mongodb://localhost:27017/?connectTimeoutMS=10000&minPoolSize=10&maxPoolSize=20&maxIdleTimeMS=60000&readPreference=primary"

[redis]

	# [redis.default]
	# address = "127.0.0.1:6379"
	# password = ""
	# database = 0
	# connect_timeout = 10
	# read_timeout = 10
	# write_timeout = 10
	# pool_size = 10
	# pool_limit = 20
	# idle_timeout = 60
	# wait_timeout = 10
	# prefill_parallelism = 0

# [nsq]
# lookupd = ["127.0.0.1:4161"]
# nsqd = "127.0.0.1:4150"

[email]

	# [email.default]
	# host = "smtp.exmail.qq.com"
	# port = 25
	# username = ""
	# password = ""

[log]

    [log.default]
    path = "logs/app.log"
    max_size = 500
    max_age = 0
    max_backups = 0
    compress = true
`
