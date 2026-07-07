package config

import (
    "os"
    "flag"
    "bufio"
    "strings"
)



type Config struct {
    ServerPort    string
    ContextPath   string
    MySQLDSN      string
    RedisAddr     string
    RedisUsername  string
    RedisPassword string
    RedisDB       int
    SessionName   string
    SessionSecret string
    SessionMaxAge int
}

func Load() (*Config, error) {
    // 1. 读取 .env 文件中的环境变量
    baseEnv := currentEnvKeys()
    envName := detectEnvName()
    loadEnvFiles(envName, baseEnv)

    // 2. 解析命令行参数
    serverPortFlag := flag.String("server-port", "", "server port")
    contextPathFlag := flag.String("context-path", "", "context path")
    mysqlDSNFlag := flag.String("mysql-dsn", "", "mysql dsn")
    redisAddrFlag := flag.String("redis-addr", "", "redis addr")
    redisUsernameFlag := flag.String("redis-username", "", "redis username")
    redisPasswordFlag := flag.String("redis-password", "", "redis password")
    redisDBFlag := flag.Int("redis-db", 0, "redis db")
    sessionNameFlag := flag.String("session-name", "", "session name")
    sessionSecretFlag := flag.String("session-secret", "", "session secret")
    sessionMaxAgeFlag := flag.Int("session-max-age-seconds", 0, "session max age seconds")
    flag.Parse()

    // 3. 按优先级组装配置
    return &Config{
        ServerPort:  pickString(*serverPortFlag, os.Getenv("SERVER_PORT"), "8123"),
        ContextPath: pickString(*contextPathFlag, os.Getenv("SERVER_CONTEXT_PATH"), "/api"),
        MySQLDSN:    pickString(*mysqlDSNFlag, os.Getenv("MYSQL_DSN"), ""),
        RedisAddr:   pickString(*redisAddrFlag, os.Getenv("REDIS_ADDR"), ""),
        RedisUsername: pickString(*redisUsernameFlag, os.Getenv("REDIS_USERNAME"), ""),
        RedisPassword: pickString(*redisPasswordFlag, os.Getenv("REDIS_PASSWORD"), ""),
        RedisDB:       *redisDBFlag,
        SessionName:   pickString(*sessionNameFlag, os.Getenv("SESSION_NAME"), ""),
        SessionSecret: pickString(*sessionSecretFlag, os.Getenv("SESSION_SECRET"), ""),
        SessionMaxAge: *sessionMaxAgeFlag,
        
    }, nil
}

func detectEnvName() string {
	envName := strings.TrimSpace(os.Getenv("APP_ENV"))
	if envName == "" {
		return "dev"
	}
	return envName
}

func loadEnvFiles(envName string, baseEnv map[string]struct{}) {
	paths := []string{
		".env",
		".env." + envName,
	}
	for _, path := range paths {
		loadEnvFile(path, baseEnv)
	}
}

func loadEnvFile(path string, baseEnv map[string]struct{}) {
	file, err := os.Open(path)
	if err != nil {
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		parts := strings.SplitN(line, "=", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		value := strings.TrimSpace(parts[1])
		value = strings.Trim(value, `"'`)
		if key == "" {
			continue
		}
		// Do not override real process-level env,
		// but allow later .env files to override earlier .env files.
		if _, exists := baseEnv[key]; exists {
			continue
		}
		_ = os.Setenv(key, value)
	}
}

func currentEnvKeys() map[string]struct{} {
	result := make(map[string]struct{})
	for _, item := range os.Environ() {
		parts := strings.SplitN(item, "=", 2)
		if len(parts) == 0 {
			continue
		}
		result[parts[0]] = struct{}{}
	}
	return result
}

func pickString(flagValue, envValue, defaultValue string) string {
    if flagValue != "" {
        return flagValue
    }
    if envValue != "" {
        return envValue
    }
    return defaultValue
}
