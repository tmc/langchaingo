package memory

import (
	"errors"
	"fmt"                       // nolint:goimports,gci,gofumpt
	"github.com/go-redis/redis" // nolint:goimports,gci
	"sync"                      // nolint:gofumpt
	"time"                      // nolint:gci
)

var ErrAddressEmpty = errors.New("redis address is empty")

type redisClientManager struct {
	sync.Once
	mu      sync.RWMutex
	clients map[string]*redis.Client
}

var ( // nolint:gofumpt
	redisClientIns = new(redisClientManager) //nolint:gochecknoglobals
)

type RedisConfOptions struct {
	Address      string
	Password     string
	DB           int
	ReadTimeout  int
	WriteTimeout int
	IdleTimeout  int
	PoolTimeout  int
	PoolSize     int
	TTl          int
	SessionID    string
	KeyPrefix    string
}

func (manager *redisClientManager) readOptions(
	redisConfOptions RedisConfOptions) (
	*redis.Options, error) {
	redisOptions := &redis.Options{}
	if redisConfOptions.Address == "" {
		return redisOptions, ErrAddressEmpty
	}
	redisOptions.Addr = redisConfOptions.Address
	if redisConfOptions.Password != "" {
		redisOptions.Password = redisConfOptions.Password
	}
	redisOptions.DB = redisConfOptions.DB
	if redisConfOptions.ReadTimeout > 0 {
		redisOptions.ReadTimeout = time.Millisecond * time.Duration(redisConfOptions.ReadTimeout)
	}
	if redisConfOptions.WriteTimeout > 0 {
		redisOptions.WriteTimeout = time.Millisecond * time.Duration(redisConfOptions.WriteTimeout)
	}
	if redisConfOptions.IdleTimeout > 0 {
		redisOptions.IdleTimeout = time.Millisecond * time.Duration(redisConfOptions.IdleTimeout)
	}
	if redisConfOptions.PoolTimeout > 0 {
		redisOptions.PoolTimeout = time.Millisecond * time.Duration(redisConfOptions.PoolTimeout)
	}
	if redisConfOptions.PoolSize > 0 {
		redisOptions.PoolSize = redisConfOptions.PoolSize
	}
	return redisOptions, nil
}

func (manager *redisClientManager) Release(address string, db int) {
	manager.mu.Lock()
	keyName := fmt.Sprintf("%s%d", address, db)
	delete(manager.clients, keyName)
	manager.mu.Unlock()
}

func (manager *redisClientManager) createClient(redisConfOptions RedisConfOptions) (*redis.Client, error) {
	options, err := manager.readOptions(redisConfOptions)
	if err != nil {
		return nil, err
	}
	client := redis.NewClient(options)
	return client, nil
}

func (manager *redisClientManager) GetClient(redisConfOptions RedisConfOptions) (*redis.Client, error) {
	keyName := fmt.Sprintf("%s%d", redisConfOptions.Address, redisConfOptions.DB)
	manager.mu.RLock()
	client, exist := manager.clients[keyName]
	manager.mu.RUnlock()
	// if client not exist
	if !exist {
		newClient, err := manager.createClient(redisConfOptions)
		if err != nil {
			return nil, err
		}
		manager.mu.Lock()
		if client, exist = manager.clients[keyName]; !exist {
			manager.clients[keyName] = newClient
			client = newClient
		}
		manager.mu.Unlock()
		if client != newClient {
			_ = newClient.Close()
		}
	}
	return client, nil
}

func (manager *redisClientManager) ReleaseAll() {
	manager.mu.Lock()
	for name, client := range manager.clients {
		_ = client.Close()
		delete(manager.clients, name)
	}
	manager.mu.Unlock()
}
