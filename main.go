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
}

type UserConfig struct {
	CalendarID  string
	CanteraName string
	UserName    string
}

// getClient uses a Context and Config to retrieve a Token
// then generate a Client. It returns the generated Client.
func getClient(ctx context.Context, config *oauth2.Config) *http.Client {
	cacheFile, err := tokenCacheFile()
	if err != nil {
		log.Fatalf("Unable to get path to cached credential file. %v", err)
	}
	tok, err := tokenFromFile(cacheFile)
	if err != nil {
		tok = getTokenFromWeb(config)
		saveToken(cacheFile, tok)
	}
	return config.Client(ctx, tok)
}

// getTokenFromWeb uses Config to request a Token.
// It returns the retrieved Token.
func getTokenFromWeb(config *oauth2.Config) *oauth2.Token {
	authURL := config.AuthCodeURL("state-token", oauth2.AccessTypeOffline)
	fmt.Printf("Go to the following link in your browser then type the "+
		"authorization code: \n%v\n", authURL)

	var code string
	if _, err := fmt.Scan(&code); err != nil {
		log.Fatalf("Unable to read authorization code %v", err)
	}

	tok, err := config.Exchange(oauth2.NoContext, code)
	if err != nil {
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

func showMembersOfDay(events *calendar.Events, when string) {
	fmt.Println("showMembersOfDay()")
	for _, i := range events.Items {
		// If the DateTime is an empty string the Event is an all-day Event.
		// So only Date is available.
		if i.Start.DateTime != "" {
			when = i.Start.DateTime
		} else {
			when = i.Start.Date
		}
		fmt.Printf("event: (%s): %q\n", when, rmSpace(i.Summary))
	}
}

/*ファイルの作成

params
	filePath: ファイルの保存場所
	fileName: ファイルの名前(ex 09-13-00-saga.md)
	txt:      ファイルの内容
*/
func mkFile(filePath string, fileName string, txt string) {
	file, err := os.OpenFile(fileName, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		log.Fatal(err)
	}
	defer file.Close()

	fmt.Fprintln(file, txt) //書き込み
}

/*ファイル名の作成

params
	cateraName: 所属(ex saga, kyukodai)
	myStart   : ユーザのシフト開始時間
return
	fileName  : ファイルの名前(ex 09-13-00-saga.md)
*/
func mkFileName(canteraName string, myStart string) string {
	t, _ := time.Parse("2006-01-02T15:04:05+09:00", myStart)
	newT := t.Format("02-15-04")
	fileName := newT + "-" + canteraName + ".md"
	fmt.Println("File name :", fileName)
	return fileName
}

func main() {
	var conf Config
	_, err := toml.DecodeFile("./config.toml", &conf)
	if err != nil {
		log.Fatalf("%v", err)
	}

	fmt.Println("名前:", conf.User.UserName)
	fmt.Println("所属:", conf.User.CanteraName)

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
	client := getClient(ctx, config)

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
		List(conf.User.CalendarID).
		ShowDeleted(false).
		SingleEvents(true).
		TimeMin(today_start).
		TimeMax(today_end).
		OrderBy("startTime").
		Do()
	if err != nil {
		log.Fatalf("Unable to retrieve next ten of the user's events. %v", err)
	}

	var when string
	name := conf.User.UserName
	canteraName := conf.User.CanteraName
	flag := false
	var event *calendar.Event

	for _, i := range events.Items {
		// If the DateTime is an empty string the Event is an all-day Event.
		// So only Date is available.
		if i.Start.DateTime != "" {
			when = i.Start.DateTime
		} else {
			when = i.Start.Date
		}
		if strings.Contains(rmSpace(i.Summary), rmSpace(name)) {
			flag = true
			event = i
		}
	}

	if flag == true {
		fmt.Printf("event: (%s): %q\n", when, event.Summary)
	} else {
		fmt.Println("本日のシフトはありません")
	}

	//mem := getSameMem(events, event.Start.DateTime)
	fileName := mkFileName(canteraName, event.Start.DateTime)
	fmt.Println(fileName)

	//showMembersOfDay(events,when)
}
