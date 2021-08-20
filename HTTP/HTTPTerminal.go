/*
HTTP Terminal handles the HTTP request responsible for mainly transmitting message
*/
package HTTPTerminal

import (
	"crypto/md5"
	"encoding/json"
	"fmt"
	"html/template"
	"io"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net/http"
	_ "net/http/pprof"
	"os"
	"shopee/entry_task/internal"
	"shopee/entry_task/pool"
	"strings"
)

var p *pool.Pool

/*
Main Program: Run takes 3 checks:
*/
func Run() {
	// create a pool maintaining by HTTP Terminal
	var err error
	p, err = pool.NewPool(internal.MaxConn, internal.MinConn)
	if err != nil {
		log.Println(err, "error in creating the pool")
	}
	log.Println("connection pool is initialized")
	defer p.Close()

	http.HandleFunc("/", index)
	http.HandleFunc("/login/", login)
	http.HandleFunc("/update/", update)
	http.HandleFunc("/img/", img)

	err = http.ListenAndServe(internal.HTTPPort, nil)
	if err != nil {
		log.Fatal(err)
	}
}

/*
Prat 1: index: for all practices,
get all info and judge the validness first before pass it to the next procedures
*/
func index(w http.ResponseWriter, r *http.Request) {
	t, err := template.ParseFiles(internal.LoginWeb)
	if err != nil {
		log.Fatal(err)
	}
	t.Execute(w, nil)
}

/*
Part 2: login: for login practices,
transfer the info to tcp and get the response from tcp
*/
func login(w http.ResponseWriter, r *http.Request) {
	// get the username, password from the front-end
	username := r.PostFormValue("username")
	password := r.PostFormValue("password")
	password = encoding(password)

	// transfer the info to tcp and get the response
	user, err := loginTry(username, password)

	if err != "" {
		// if failure, the front-end will show the alert
		p := internal.LoginFailure{Success: false, Reason: err}
		t, err := template.ParseFiles(internal.ProfileWeb)

		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}
		t.Execute(w, p)
		log.Printf("%v login fails!", username)
	} else {
		// if success, the front end will show the profile page
		p := internal.LoginSuccess{
			Success:  true,
			Username: user.Username,
			Nickname: user.Nickname,
			Profile:  user.Profile,
		}

		t, err := template.ParseFiles(internal.ProfileWeb)
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}
		t.Execute(w, p)
		log.Printf("%v login succeded! (nickname: %v, profile: %v)", username, p.Nickname, p.Profile)
	}
}

/*
Part 3: update: for update practices
*/
func update(w http.ResponseWriter, r *http.Request) {
	username := r.PostFormValue("username")
	nickname := r.PostFormValue("nickname")
	profile_old := r.PostFormValue("profile_old")

	img, header, _ := r.FormFile("profile")
	log.Println(img == nil, header == nil, profile_old)

	var profile string
	if img != nil && header != nil {
		// if new image is uploaded
		extension, err := imgUpLoading(username, img, header)
		if err != nil {
			fmt.Fprintln(w, err.Error())
			return
		}
		profile = fmt.Sprintf("%s.%s", username, extension)
	} else {
		// if no new image, then the old one
		profile = profile_old
	}
	ok := updateTry(username, nickname, profile)
	if !ok {
		log.Println("refresh fails")
		return
	}

	p := internal.LoginSuccess{
		Success:  true,
		Username: username,
		Nickname: nickname,
		Profile:  profile,
	}

	t, err := template.ParseFiles(internal.ProfileWeb)
	if err != nil {
		fmt.Fprintln(w, err.Error())
		return
	}
	t.Execute(w, p)
	log.Printf("%v updated succeded! (nickname: %v, profile: %v)", username, nickname, profile)
}

// Try to login and establish message communication with tcp
func loginTry(username, password string) (*internal.UserInfo, string) {
	loginMsg := internal.JsonLogin{
		JsonType: internal.JsonType{Type: "login"},
		Username: username,
		Password: password,
	}
	buf, _ := json.Marshal(loginMsg)

	msg, err := p.TimingCall(buf)
	if err != nil {
		log.Println(err)
		return nil, err.Error()
	}

	result := &internal.JsonLoginRst{}
	json.Unmarshal([]byte(msg), result)
	if !result.Result {
		return nil, result.Reason
	}
	return &result.UserInfo, ""
}

// refresh the page when the user info is changed
func updateTry(username, nickname, profile string) bool {
	loginMsg := internal.JsonUpdate{
		JsonType: internal.JsonType{Type: "update"},
		Username: username,
		Nickname: nickname,
		Profile:  profile,
	}
	buf, _ := json.Marshal(loginMsg)

	msg, err := p.TimingCall(buf)
	if err != nil {
		return false
	}

	rst := &internal.JsonUpdateRst{}
	err = json.Unmarshal([]byte(msg), rst)
	if err != nil {
		return false
	}
	return rst.Result
}

// load the image while logging in
func img(w http.ResponseWriter, r *http.Request) {
	path := r.URL.Path
	if path == "/img/" {
		path += "default.png"
	}
	file, err := ioutil.ReadFile(path[1:])
	if err != nil {
		fmt.Fprintf(w, "avatar loaded failure.")
		return
	}
	w.Write(file)
}

// Uploading images in updating
func imgUpLoading(username string, img multipart.File, handle *multipart.FileHeader) (string, error) {
	defer img.Close()
	log.Println("uploaded file's name is", handle.Filename)
	_, err := os.Stat("./img/")
	if err != nil {
		log.Println(err)
		return "", err
	}
	imgType := [...]string{"jpg", "jpeg", "png", "gif"}
	for _, imgT := range imgType {
		os.Remove(fmt.Sprintf("./img/%s.%s", username, imgT))
	}
	temp := strings.Split(handle.Filename, ".")
	extension := strings.ToLower(temp[len(temp)-1])
	saveImg(img, fmt.Sprintf("%s.%s", username, extension))
	return extension, nil
}

// Downloading images in disk and store the path in DB
func saveImg(img multipart.File, name string) {
	file, err := os.Create(fmt.Sprintf("./img/%s", name))
	if err != nil {
		log.Println(err)
		return
	}
	_, err = io.Copy(file, img)
	if err != nil {
		log.Println(err)
		return
	}
	file.Close()
}

func encoding(password string) string {
	encoded := (password + "qwertYUIOP")[:15]
	return fmt.Sprintf("%x", md5.Sum([]byte(encoded)))
}
