package container

import "github.com/ilyakaznacheev/cleanenv"

type config struct {
	Wasm struct {
		MaxPages uint32 `env:"WASM_MAX_PAGES" env-default:"2048"`                    // 128 mb
		StoreDir string `env:"WASM_STORE_DIR" env-default:"/var/lib/mini-faas/wasm"` // where .wasm binaries are stored
	}
	Docker struct {
		RegistryPath string `env:"DOCKER_REGISTRY_PATH"` // localhost:5000
	}
	ContainerdSockPath string `env:"CONTAINERD_SOCKET_PATH"` // /run/containerd/containerd.sock
	ApiGatewayBaseUrl  string `env:"APIGATEWAY_URL_PATH" env-default:"http://localhost:8080"`
}

func NewConfig() (*config, error) {
	var cfg config
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		return nil, err
	}
	return &cfg, nil
}
