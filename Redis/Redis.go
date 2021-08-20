/*
Handle the operation with Redis
*/
package Redis

import (
	"fmt"
	"log"
	internal "shopee/entry_task/internal"

	"github.com/go-redis/redis/v7"
)

const (
	redisClient   = ":6379"
	redisPassword = ""
)

var client *redis.Client = nil

// initiate the Redis and sync with part of the MySQL data through channel
func init() {
	client = redis.NewClient(&redis.Options{
		Addr:     redisClient,
		Password: redisPassword,
		DB:       0,
	})
	_, err := client.Ping().Result()
	if err != nil {
		log.Println(err)
	}
	log.Println("redis connected.")
}

// login: Authenticate the username and password in Redis
func GetUserLogin(username, password string) (errOK bool, passwordOK bool) {
	res, err := client.Get(fmt.Sprintf("user:%s", username)).Result()
	if err != nil {
		return false, true
	}
	if password != res {
		return true, false
	}
	return true, true
}

// login: if not stored in Redis, cached into Redis from MySQL
func SetUserLogin(username, password string) error {
	_, err := client.Set(fmt.Sprintf("user:%s", username), password, 0).Result()
	if err != nil {
		log.Println(err)
	}
	return err
}

// update: get user info from Redis
func GetUserInfo(username string) *internal.UserInfo {
	nickname, err := client.HGet(fmt.Sprintf("info:%s", username), "nickname").Result()
	if err != nil {
		log.Println(err)
		return nil
	}
	profile, err := client.HGet(fmt.Sprintf("info:%s", username), "profile").Result()
	if err != nil {
		log.Println(err)
		return nil
	}
	return &internal.UserInfo{Username: username, Nickname: nickname, Profile: profile}
}

// update: set user info
func SetNickname(username, nickname string) error {
	if nickname != "" {
		_, err := client.HMSet(fmt.Sprintf("info:%s", username), "username", username, "nickname", nickname).Result()
		return err
	}
	return nil
}

func SetProfile(username, profile string) error {
	if profile != "" {
		_, err := client.HMSet(fmt.Sprintf("info:%s", username), "username", username, "profile", profile).Result()
		return err
	}
	return nil
}

func DelUserInfo(username string) error {
	_, err := client.Del(fmt.Sprintf("info:%s", username), "username").Result()
	return err
}
