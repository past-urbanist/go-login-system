/*
TCP terminal handles the HTTP request and search in DB for keys
*/
package TCPTerminal

import (
	"encoding/json"
	"log"
	"net"
	"shopee/entry_task/MySQL"
	"shopee/entry_task/Redis"
	"shopee/entry_task/internal"
	"shopee/entry_task/pool"
)

/*
Main Program: connection with http
*/
func Run() {
	listener, err := net.Listen("tcp", internal.TCPPort)

	if err != nil {
		log.Println(err)
		return
	}
	defer listener.Close()

	for {
		conn, err := listener.Accept()
		if err != nil {
			log.Println(err)
			return
		}
		go handleConn(conn)
	}
}

// handles the connection with http
func handleConn(conn net.Conn) {
	t := pool.NewTransport(conn)
	for {
		msg, err := t.Receive()
		if err != nil {
			log.Println(err)
			break
		}
		// if msg succesfully arrives at the tcp terminal
		jt := &internal.JsonType{}
		err = json.Unmarshal(msg, jt)
		if err != nil {
			log.Println(err)
		}

		var msg2 []byte
		if jt.Type == "login" {
			msg2, err = handleLogin(msg, t)
		} else {
			msg2, err = handleUpdate(msg, t)
		}
		if err != nil {
			log.Println(err)
		}

		err = t.Send(msg2)
		if err != nil {
			log.Println(err)
		}
	}
}

// handle the login requests
func handleLogin(msg []byte, t *pool.Transport) ([]byte, error) {
	login := &internal.JsonLogin{}
	err := json.Unmarshal(msg, login)
	if err != nil {
		return nil, err
	}
	username, password := login.Username, login.Password

	rst := &internal.JsonLoginRst{}
	rst.Type = "login"
	redisOK, passwordOK := Redis.GetUserLogin(username, password)

	// if password is wrong
	if redisOK && !passwordOK {
		// cached in Redis but password is wrong
		rst.Result = false
		rst.Reason = "invalid username or password (redis)."
		buf, _ := json.Marshal(rst)
		return buf, nil
	} else {
		// not cache in Redis: stored in SQL, otherwise wrong answer
		SQLOK, _ := MySQL.GetUserLogin(username, password)
		if !SQLOK {
			rst.Result = false
			rst.Reason = "invalid username or password (mysql)."
			buf, _ := json.Marshal(rst)
			return buf, nil
		}
	}

	// if password is correct
	var userInfo *internal.UserInfo
	userInfo = Redis.GetUserInfo(username)
	if userInfo == nil {
		userInfo = MySQL.GetUserInfo(username)
		go Redis.SetUserLogin(username, password)
		go Redis.SetNickname(userInfo.Username, userInfo.Nickname)
		go Redis.SetProfile(userInfo.Username, userInfo.Profile)
		log.Println("data got from mysql.")
	} else {
		log.Println("data got from redis.")
	}
	rst.Result = (userInfo != nil)
	rst.UserInfo = *userInfo
	buf, _ := json.Marshal(rst)
	return buf, nil
}

// handle the update requests
func handleUpdate(msg []byte, t *pool.Transport) ([]byte, error) {
	update := &internal.JsonUpdate{}
	err := json.Unmarshal(msg, update)
	if err != nil {
		return nil, err
	}

	rst := &internal.JsonUpdateRst{}
	rst.Type = "update"
	var SQLErr, RedisErr error
	// first update in mysql and then delete in redis for consistency
	SQLErr = MySQL.SetNickname(update.Username, update.Nickname)
	if SQLErr != nil {
		rst.Result = false
		rst.Reason = SQLErr.Error()
	}
	SQLErr = MySQL.SetProfile(update.Username, update.Profile)
	if SQLErr != nil {
		rst.Result = false
		rst.Reason = SQLErr.Error()
	}
	RedisErr = Redis.DelUserInfo(update.Username)
	if RedisErr != nil {
		rst.Result = false
		rst.Reason = RedisErr.Error()
	}

	rst.Result = true
	buf, err := json.Marshal(*rst)
	return buf, err
}
