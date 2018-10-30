package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"strconv"
	"strings"
	"time"
)

// SERVER 服务器地址
const SERVER = "http://127.0.0.1:5000"

const _API = "https://api.appsign.vip:2688"

// APPINFO 应用信息
var APPINFO = map[string]string{
	"version_code": "2.7.0",
	"app_version":  "2.7.0",
	"channel":      "App%20Stroe",
	"app_name":     "aweme",
	"build_number": "27014",
	"aid":          "1128",
}

type DEVICE struct {
	Openudid        string
	Idfa            string
	Vid             string
	Install_id      int
	Iid             int
	Device_id       int
	New_user        int
	Device_type     string
	Os_version      string
	Os_api          string
	Screen_width    string
	Device_platform string
}
type SIGN struct {
	Mas string
	As  string
	Ts  string
}

var errorNumber = 1

// 当然是请求任务啦
func getJob() (string, int) {
	resp, err := http.Get(SERVER + "/getJob")
	if err != nil {
		return err.Error(), 1
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return err.Error(), 1
	}
	return string(body), 0
}

func get(url string) []byte {
	client := &http.Client{}
	fmt.Println(url)
	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		// handle error
	}

	req.Header.Set("User-Agent", "Aweme/2.8.0 (iPhone; iOS 11.0; Scale/2.00)")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	// req.Header.Set("Cookie", "name=anny")

	resp, err := client.Do(req)

	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	return body
}

func post(url string, data string) []byte {
	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	return body
}

// 获取新的设备信息:有效期60分钟永久
func get_token() string {
	url := _API + "/token/douyin"
	res := get(url)
	type token struct {
		Success bool
		Token   string
	}
	var tokenData token
	if err := json.Unmarshal(res, &tokenData); err == nil {
		log.Println(tokenData.Token)
		return tokenData.Token
	} else {
		log.Println(err.Error())
	}
	return ""
}

func get_device() DEVICE {
	url := _API + "/douyin/device/new"
	res := get(url)
	type device struct {
		Success bool
		Data    DEVICE
	}
	var deviceData device
	if err := json.Unmarshal(res, &deviceData); err == nil {
		return deviceData.Data
	} else {
		log.Println(err.Error())
	}
	return deviceData.Data
}

func get_sign(token string, device DEVICE) (SIGN, string) {
	url := _API + "/sign"
	query := `openudid=` + device.Openudid + `&idfa=` + device.Idfa + `&vid=` + device.Vid + `&install_id=` + strconv.Itoa(device.Install_id) + `&iid=` + strconv.Itoa(device.Iid) + `&device_id=` + strconv.Itoa(device.Device_id) + `&new_user=` + strconv.Itoa(device.New_user) + `&device_type=` + device.Device_type + `&os_version=` + device.Os_version + `&os_api=` + device.Os_api + `&screen_width=` + device.Screen_width + `&device_platform=` + device.Device_platform + `&version_code=2.7.0&app_version=2.7.0&channel=App%20Stroe&app_name=aweme&build_number=27014&aid=1128`
	jsonData := `{"token":"` + token + `","query":"` + query + `"}`
	// log.Println(jsonData)
	res := post(url, jsonData)
	type sign struct {
		Success bool
		Data    SIGN
	}
	var signData sign
	if err := json.Unmarshal(res, &signData); err == nil {
		query = query + `&mas=` + signData.Data.Mas + `&as=` + signData.Data.As + `&ts=` + signData.Data.Ts
		return signData.Data, query
	} else {
		log.Println(err.Error())
	}
	return signData.Data, query
	// log.Println(string(b), token, err)
}

// 给请求 params 签名
func get_signed_params() {
	device := get_device()
	token := get_token()
	sign, query := get_sign(token, device)
	url := "https://aweme.snssdk.com/aweme/v1/user/following/list/?"
	t := time.Now()
	timestamp := strconv.FormatInt(t.UTC().UnixNano(), 10)
	log.Println(sign, url+"user_id=98105997680&offset=0&count=49&source_type=2&ac=WIFI&max_time="+timestamp[:10]+"&"+query)
	res := get(url + "user_id=98105997680&offset=0&count=49&source_type=2&ac=WIFI&max_time=" + timestamp[:10] + "&" + query)
	log.Println(string(res))
}

// 抖音的签名请求函数
func curl() {
	get_signed_params()
}

func work(workList string) {
	md5Data := workList[:32]
	data := workList[32:]
	// fmt.Print(data)
	decodedData, err := base64.StdEncoding.DecodeString(data)
	// 解析失败了那0.01%是伪造的请求, 99.9%是服务端挂了,那还执行毛任务啊歇着吧
	if err == nil {
		has := md5.Sum(decodedData)
		md5Check := fmt.Sprintf("%x", has)
		if md5Data == md5Check {
			// 开始解析服务器发来什么玩意
			var message interface{}
			if err := json.Unmarshal(decodedData, &message); err == nil {
				resData := message.(map[string]interface{})
				// fmt.Println(reflect.TypeOf(resData["data"]))
				workList := resData["data"].([]interface{})
				for _, workItem := range workList {
					fmt.Println(workItem)
				}
				curl()
			}
		}
	} else {
		fmt.Printf(err.Error())
	}
}

// 获取工作的函数
func getWork() {
	// 没有找到工作要你有毛用啊 继续找
	data, err := getJob()
	if err != 0 {
		for true {
			time.Sleep(time.Duration(1) * time.Second)
			data, err := getJob()
			if err != 0 {
				fmt.Println("第" + strconv.Itoa(errorNumber) + "次尝试与服务器建立连接失败,1分钟后重试!")
				errorNumber++
			} else {
				work(data)
			}
			// 请求100次服务器都返回错误 100%是我的垃圾服务挂了 洗洗睡吧
			if errorNumber > 10 {
				// 问我为什么是中文输出(这样岂不是很没有逼格?), 我的回答是:要照顾每个使用者的智商(我其实也不懂英文啊!!!)
				fmt.Println("与服务器建立连接失败!")
				break
			}
			// 没有找到工作要你有毛用啊 继续找
		}
	} else {
		work(data)
	}
}

func main() {
	// 请求100次服务器都返回错误 100%是我的垃圾服务挂了 洗洗睡吧
	if errorNumber > 100 {
		// 问我为什么是中文输出(这样岂不是很没有逼格?), 我的回答是:要照顾每个使用者的智商(我其实也不懂英文啊!!!)
		fmt.Println("与服务器建立连接失败!")
	}
	getWork()
}
