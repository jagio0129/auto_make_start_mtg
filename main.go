package main

import (
	"encoding/json"
	"fmt"
	"github.com/BurntSushi/toml"
	"golang.org/x/net/context"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
	"google.golang.org/api/calendar/v3"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"os"
	"os/user"
	"path/filepath"
	"strings"
	"time"
)

type Config struct {
	User UserConfig
	Env  EnvConfig
}

type UserConfig struct {
	CanteraName string
	UserName    string
}

type EnvConfig struct {
	CalendarID        string
	AuthorizationCode string
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config, myConf Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config, myConf)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config, myConf Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	//fmt.Printf("Go to the following link in your browser then type the "+
	//	"authorization code: \n%v\n", authURL)
	if myConf.Env.AuthorizationCode == "" {
		fmt.Println("url.txtファイルに記載されいるURLにアクセスし、表示されるコードをconfig.tomlのauthorizationCodeに記入してください。")
		file, err := os.OpenFile("./url.txt", os.O_RDWR|os.O_CREATE, 0666)
		if err != nil {
			log.Fatal(err)
		}
		defer file.Close()

		fmt.Fprintln(file, authURL) //書き込み
		time.Sleep(5 * time.Second)
		log.Fatal("終了")
	}

	tok, err := config.Exchange(oauth2.NoContext, myConf.Env.AuthorizationCode)
	if err != nil {
		fmt.Println("config.tomlのauthorizationCodeが正しく設定されていません。")
		fmt.Println("url.txtファイルに記載されいるURLにアクセスし、表示されるコードをconfig.tomlのauthorizationCodeに記入してください。")
		time.Sleep(3 * time.Second)
		log.Fatalf("Unable to retrieve token from web %v", err)
	}
	return tok
}

// tokenCacheFile generates credential file path/filename.
// It returns the generated credential path/filename.
func tokenCacheFile() (string, error) {
	usr, err := user.Current()
	if err != nil {
		return "", err
	}
	tokenCacheDir := filepath.Join(usr.HomeDir, ".credentials")
	os.MkdirAll(tokenCacheDir, 0700)
	return filepath.Join(tokenCacheDir,
		url.QueryEscape("calendar-go-quickstart.json")), err
}

// tokenFromFile retrieves a Token from a given file path.
// It returns the retrieved Token and any read error encountered.
func tokenFromFile(file string) (*oauth2.Token, error) {
	f, err := os.Open(file)
	if err != nil {
		return nil, err
	}
	t := &oauth2.Token{}
	err = json.NewDecoder(f).Decode(t)
	defer f.Close()
	return t, err
}

// saveToken uses a file path to create a file and store the
// token in it.
func saveToken(file string, token *oauth2.Token) {
	fmt.Printf("Saving credential file to: %s\n", file)
	f, err := os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		log.Fatalf("Unable to cache oauth token: %v", err)
	}
	defer f.Close()
	json.NewEncoder(f).Encode(token)
}

func rmSpace(str string) (summary string) {
	return strings.Replace(str, " ", "", -1)
}

func getSameMem(events *calendar.Events, myStart string) []string {
	var sameMem []string
	for _, i := range events.Items {
		if i.Start.DateTime == myStart {
			sameMem = append(sameMem, i.Summary)
		}
	}
	return sameMem
}

func mkFile(filePath string, fileName string, txt string) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintln(file, txt) //書き込み
}

func mkFileName(filePath string, canteraName string, myStart string) string {
	t, _ := time.Parse("2006-01-02T15:04:05+09:00", myStart)
	newT := t.Format("02-15-04")
	fileName := filePath + "\\Desktop\\" + newT + "-" + canteraName + ".md"
	return fileName
}

// define file content
func mkTxt(membList []string) string {
	var txt string

	txt = "# メンバ\n"
	for _, v := range membList {
		txt += "- " + v + "\n"
	}

	txt += "\n"
	txt += "# タスク内容\n"
	for _, v := range membList {
		txt += "- " + v + "\n"
		txt += "  - \n"
	}

	txt += "\n"
	txt += "# 共有事項\n"

	return txt
}

func main() {
	var myConf Config
	_, err := toml.DecodeFile("./config.toml", &myConf)
	if err != nil {
		log.Fatalf("%v", err)
	}

	ctx := context.Background()

	b, err := ioutil.ReadFile("client_secret.json")
	if err != nil {
		log.Fatalf("Unable to read client secret file: %v", err)
	}

	// If modifying these scopes, delete your previously saved credentials
	// at ~/.credentials/calendar-go-quickstart.json
	config, err := google.ConfigFromJSON(b, calendar.CalendarReadonlyScope)
	if err != nil {
		log.Fatalf("Unable to parse client secret file to config: %v", err)
	}
	client := getClient(ctx, config, myConf)

	srv, err := calendar.New(client)
	if err != nil {
		log.Fatalf("Unable to retrieve calendar Client %v", err)
	}

	// 日付を指定
	year, month, day := time.Now().Date()
	today_start := time.Date(year, month, day, 00, 00, 00, 0, time.Local).Format(time.RFC3339)
	today_end := time.Date(year, month, day, 23, 59, 59, 0, time.Local).Format(time.RFC3339)

	// 指定したカレンダーIDの情報取得
	events, err := srv.Events.
		List(myConf.Env.CalendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(today_start).
		TimeMax(today_end).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}

	usr, err := user.Current()
	if err != nil {
		log.Fatal(err)
	}

	userName := myConf.User.UserName
	canteraName := myConf.User.CanteraName
	userHome := usr.HomeDir
	hasPartTime := false

	var event *calendar.Event
	for _, v := range events.Items {
		if strings.Contains(rmSpace(v.Summary), rmSpace(userName)) {
			hasPartTime = true
			event = v
		}
	}
	if hasPartTime == false {
		fmt.Println("本日のシフトはありません")
	} else {
		fmt.Println("start meetingファイルを作成します")
		mem := getSameMem(events, event.Start.DateTime)
		fileName := mkFileName(userHome, canteraName, event.Start.DateTime)
		txt := mkTxt(mem)
		mkFile(userHome, fileName, txt)
	}
	time.Sleep(3 * time.Second)
}
