package cache

import (
	"encoding/json"
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/gomodule/redigo/redis"
)

var (
	pool *redis.Pool
	mu   sync.Mutex
)

// Config Redis 连接参数
type Config struct {
	Host     string
	Port     string
	Password string
	DB       int
}

// Init 初始化 Redis 连接池
func Init(cfg Config) {
	addr := fmt.Sprintf("%s:%s", cfg.Host, cfg.Port)

	// ===== 先用一次性的直连测试，避免 pool.Get() 在连接失败时阻塞 =====
	testConn, err := redis.Dial("tcp", addr,
		redis.DialConnectTimeout(3*time.Second),
		redis.DialReadTimeout(3*time.Second),
		redis.DialWriteTimeout(3*time.Second),
	)
	if err != nil {
		log.Printf("⚠️  Redis 不可用: %v（将回退到内存缓存）", err)
		pool = nil
		return
	}
	defer testConn.Close()

	// 认证
	if cfg.Password != "" {
		if _, err := testConn.Do("AUTH", cfg.Password); err != nil {
			log.Printf("⚠️  Redis AUTH 失败: %v（将回退到内存缓存）", err)
			pool = nil
			return
		}
	}
	// 选择 DB
	if cfg.DB != 0 {
		if _, err := testConn.Do("SELECT", cfg.DB); err != nil {
			log.Printf("⚠️  Redis SELECT 失败: %v（将回退到内存缓存）", err)
			pool = nil
			return
		}
	}
	// PING 验证
	if _, err := testConn.Do("PING"); err != nil {
		log.Printf("⚠️  Redis PING 失败: %v（将回退到内存缓存）", err)
		pool = nil
		return
	}

	// ===== 测试通过，创建连接池 =====
	pool = &redis.Pool{
		MaxIdle:     10,
		MaxActive:   50,
		IdleTimeout: 300 * time.Second,
		Wait:        true,
		Dial: func() (redis.Conn, error) {
			conn, err := redis.Dial("tcp", addr,
				redis.DialConnectTimeout(3*time.Second),
				redis.DialReadTimeout(5*time.Second),
				redis.DialWriteTimeout(5*time.Second),
			)
			if err != nil {
				return nil, err
			}
			if cfg.Password != "" {
				if _, err := conn.Do("AUTH", cfg.Password); err != nil {
					conn.Close()
					return nil, err
				}
			}
			if cfg.DB != 0 {
				if _, err := conn.Do("SELECT", cfg.DB); err != nil {
					conn.Close()
					return nil, err
				}
			}
			return conn, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			if time.Since(t) < time.Minute {
				return nil
			}
			_, err := c.Do("PING")
			return err
		},
	}

	log.Println("✅ Redis 连接成功:", addr)
}

// Close 关闭 Redis 连接池
func Close() {
	if pool != nil {
		pool.Close()
	}
}

// ==================== 内存缓存（Redis 不可用时回退） ====================

var memStore = make(map[string]memItem)

type memItem struct {
	Value    any
	ExpireAt time.Time
}

// ==================== 缓存操作（全部带 panic recover 保护） ====================

// getConn 安全地从池中获取连接（带 recover）
func getConn() redis.Conn {
	defer func() {
		if r := recover(); r != nil {
			log.Printf("⚠️  Redis pool.Get() 异常: %v，切换回内存缓存", r)
			pool = nil
		}
	}()
	if pool != nil {
		return pool.Get()
	}
	return nil
}

// Set 写入缓存
func Set(key string, value any, ttl time.Duration) {
	conn := getConn()
	if conn != nil {
		defer conn.Close()
		// 字符串直接存，避免 json.Marshal 加上多余的双引号
		var data []byte
		if str, ok := value.(string); ok {
			data = []byte(str)
		} else {
			var err error
			data, err = json.Marshal(value)
			if err != nil {
				log.Printf("❌ 缓存序列化失败 key=%s: %v", key, err)
				return
			}
		}
		ttlSeconds := int(ttl.Seconds())
		if ttlSeconds > 0 {
			_, err := conn.Do("SET", key, data, "EX", ttlSeconds)
			if err != nil {
				log.Printf("❌ Redis 写入失败 key=%s: %v（回退内存）", key, err)
				pool = nil
				goto memFallback
			}
		} else {
			_, err := conn.Do("SET", key, data)
			if err != nil {
				log.Printf("❌ Redis 写入失败 key=%s: %v（回退内存）", key, err)
				pool = nil
				goto memFallback
			}
		}
		return
	}

memFallback:
	mu.Lock()
	memStore[key] = memItem{
		Value:    value,
		ExpireAt: time.Now().Add(ttl),
	}
	mu.Unlock()
}

// Get 读取缓存（返回原始值，与旧版兼容）
func Get(key string) (any, bool) {
	conn := getConn()
	if conn != nil {
		defer conn.Close()
		data, err := redis.Bytes(conn.Do("GET", key))
		if err != nil {
			return nil, false
		}
		return data, true
	}

	mu.Lock()
	defer mu.Unlock()
	item, ok := memStore[key]
	if !ok || time.Now().After(item.ExpireAt) {
		return nil, false
	}
	return item.Value, true
}

// Del 删除缓存
func Del(key string) {
	conn := getConn()
	if conn != nil {
		defer conn.Close()
		conn.Do("DEL", key)
		return
	}

	mu.Lock()
	delete(memStore, key)
	mu.Unlock()
}

// ==================== 带类型反序列化的便捷方法 ====================

// GetJSON 读取缓存并自动 JSON 反序列化到目标结构体
// dest 必须是指针
func GetJSON(key string, dest any) bool {
	conn := getConn()
	if conn != nil {
		defer conn.Close()
		data, err := redis.Bytes(conn.Do("GET", key))
		if err != nil {
			return false
		}
		if err := json.Unmarshal(data, dest); err != nil {
			log.Printf("❌ 缓存反序列化失败 key=%s: %v", key, err)
			return false
		}
		return true
	}

	mu.Lock()
	item, ok := memStore[key]
	mu.Unlock()
	if !ok || time.Now().After(item.ExpireAt) {
		return false
	}

	data, err := json.Marshal(item.Value)
	if err != nil {
		return false
	}
	return json.Unmarshal(data, dest) == nil
}

// GetString 读取缓存中的字符串值
func GetString(key string) (string, bool) {
	conn := getConn()
	if conn != nil {
		defer conn.Close()
		s, err := redis.String(conn.Do("GET", key))
		if err != nil {
			return "", false
		}
		return s, true
	}

	mu.Lock()
	item, ok := memStore[key]
	mu.Unlock()
	if !ok || time.Now().After(item.ExpireAt) {
		return "", false
	}
	s, ok := item.Value.(string)
	return s, ok
}

// Exists 检查缓存键是否存在
func Exists(key string) bool {
	conn := getConn()
	if conn != nil {
		defer conn.Close()
		n, err := redis.Int(conn.Do("EXISTS", key))
		if err != nil {
			return false
		}
		return n > 0
	}

	mu.Lock()
	item, ok := memStore[key]
	mu.Unlock()
	if !ok || time.Now().After(item.ExpireAt) {
		return false
	}
	return true
}
