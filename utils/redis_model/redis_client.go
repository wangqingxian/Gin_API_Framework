package redis_model

import (
    "github.com/garyburd/redigo/redis"
    "time"
    "log"
)

// Wraps the Redis client to meet the Cache interface.
type RedisClient struct {
    pool              *redis.Pool
    defaultExpiration time.Duration
}

// until redigo supports sharding/clustering, only one host will be in hostList
func NewRedisCache(host string, password string, defaultExpiration time.Duration) *RedisClient {
    var pool = &redis.Pool{
        MaxIdle:     5,
        IdleTimeout: 0 * time.Second,
        Dial: func() (redis.Conn, error) {
            // the redis protocol should probably be made sett-able
            c, err := redis.Dial("tcp", host)
            if err != nil {
                return nil, err
            }
            if len(password) > 0 {
                if _, err := c.Do("AUTH", password); err != nil {
                    c.Close()
                    return nil, err
                }
            } else {
                // check with PING
                if _, err := c.Do("PING"); err != nil {
                    c.Close()
                    return nil, err
                }
            }
            return c, err
        },
        // custom connection test method
        TestOnBorrow: func(c redis.Conn, t time.Time) error {
            if _, err := c.Do("PING"); err != nil {
                return err
            }
            return nil
        },
    }
    return &RedisClient{pool, defaultExpiration}
}

/* Queue Function */
func (c *RedisClient) LPush(key string,  value []byte) error {
    conn := c.pool.Get()
    defer conn.Close()
    raw, err := conn.Do("LPUSH", key , value)
    if raw == nil {
        return nil
    }
    item, err := redis.Bytes(raw, err)
    if item !=nil {
        log.Println(item)
    }
    return err
}


func (c *RedisClient) BRpop(key string) ( []interface {}, error) {
    conn := c.pool.Get()
    defer conn.Close()
    raw, err := conn.Do("BRPOP", key , 0)
    //log.Println("[Redis Client Raw]",raw)
    if err != nil {
        log.Println("[Redis Client BRpop]",err)
    }

    item, ierr := redis.Values(raw, err)
    if item !=nil {
        log.Println("[Redis Client BRpop]",ierr)
    }

    //log.Println("[Redis Client item]",item)
    return item, err
}
