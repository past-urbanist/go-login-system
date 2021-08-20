/*
Handle the operation with MySQL
*/

package MySQL

import (
	"database/sql"
	"fmt"
	"log"
	internal "shopee/entry_task/internal"
	"time"

	_ "github.com/go-sql-driver/mysql"
)

const (
	username = "root"
	password = "Zke_1997"
	hostname = "127.0.0.1:3306"
	dbname   = "user_info"
)

var db *sql.DB = nil
var MqLogin chan internal.Login = make(chan internal.Login, 1024)
var MqUsers chan internal.UserInfo = make(chan internal.UserInfo, 1024)

// Connect to MySQL Database
func init() {
	name := fmt.Sprintf("%s:%s@tcp(%s)/%s", username, password, hostname, dbname)
	var err error
	db, err = sql.Open("mysql", name)
	if err != nil {
		log.Printf("Error %s when opening DB\n", err)
		return
	}
	// defer db.Close()

	db.SetMaxOpenConns(20)
	db.SetMaxIdleConns(20)
	db.SetConnMaxLifetime(time.Minute * 5)

	// chanCache(db)
	log.Println("mysql connected.")
}

// login: Authenticate the username and password in MySQL
func GetUserLogin(username, password string) (bool, string) {
	row := db.QueryRow("SELECT password FROM user where username=?", username)
	var rst string
	err := row.Scan(&rst)
	if err != nil {
		log.Println(err)
		return false, err.Error()
	}

	if password != rst {
		log.Println(password, rst)
		return false, "invalid username or password"
	}
	return true, ""
}

// update: get user info from MySQL
func GetUserInfo(username string) *internal.UserInfo {
	row := db.QueryRow("SELECT nickname, url FROM user where username=?", username)
	var nickname, profile sql.NullString
	err := row.Scan(&nickname, &profile)
	if err != nil {
		log.Println(err)
		return nil
	}
	return &internal.UserInfo{Username: username, Nickname: nickname.String, Profile: profile.String}
}

// update: set user info in MySQL
func SetNickname(username, nickname string) error {
	if nickname != "" {
		update, _ := db.Prepare("UPDATE user SET nickname=? WHERE username=?")
		_, err := update.Exec(nickname, username)
		return err
	}
	return nil
}

func SetProfile(username, profile string) error {
	if profile != "" {
		update, _ := db.Prepare("UPDATE user SET url=? WHERE username=?")
		_, err := update.Exec(profile, username)
		return err
	}
	return nil
}
