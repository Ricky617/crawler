package main

import (
	"crypto/md5"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// SERVER 母机地址
const SERVER = "http://127.0.0.1:5000"

// APISERVER 调用Api接口
const APISERVER = "https://api.appsign.vip:2688"

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

func get(requestUrl string) []byte {
	client := &http.Client{}
	fmt.Printf(requestUrl)
	// requestUrl = strings.Replace(requestUrl, "App%20Stroe", "App%2520Stroe", -1)
	// requestUrl = strings.Replace(requestUrl, "iPhone8,1", "iPhone8%2C1", -1)
	req, err := http.NewRequest("GET", requestUrl, nil)
	if err != nil {
		// handle error
	}

	req.Header.Set("User-Agent", "Aweme/2.8.0 (iPhone; iOS 11.0; Scale/2.00)")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		fmt.Println(err.Error())
		return []byte("")
	}
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
func getToken() string {
	url := APISERVER + "/token/douyin/version/2.7.0"
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

// getDevice 请求设备信息
func getDevice() string {
	url := APISERVER + "/douyin/device/new/version/2.7.0"
	res := get(url)
	// DEVICE 设备信息
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
		OsAPISERVER     string
		Screen_width    string
		Device_platform string
	}
	type device struct {
		Success bool
		Data    DEVICE
	}
	var deviceData device
	if err := json.Unmarshal(res, &deviceData); err == nil {
		temp := deviceData.Data
		return `openudid=` + temp.Openudid + `&idfa=` + temp.Idfa + `&vid=` + temp.Vid + `&install_id=` + strconv.Itoa(temp.Install_id) + `&iid=` + strconv.Itoa(temp.Iid) + `&device_id=` + strconv.Itoa(temp.Device_id) + `&new_user=` + strconv.Itoa(temp.New_user) + `&device_type=` + temp.Device_type + `&os_version=` + temp.Os_version + `&osAPISERVER=` + temp.OsAPISERVER + `&screen_width=` + temp.Screen_width + `&device_platform=` + temp.Device_platform
	} else {
		log.Println(err.Error())
	}
	return ""
}

// getSign 获取签名信息
func getSign(token string, device string) string {
	url := APISERVER + "/sign"
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano(), 10)
	query := "user_id=98105997680&offset=0&count=49&source_type=2&max_time=" + timestamp[:10] + "&ac=WIFI&" + device + `&version_code=2.7.0&app_version=2.7.0&channel=App%20Stroe&app_name=aweme&build_number=27014&aid=1128`
	jsonData := `{"token":"` + token + `","query":"` + query + `"}`
	res := post(url, jsonData)
	type SIGN struct {
		Mas string
		As  string
		Ts  string
	}
	type sign struct {
		Success bool
		Data    SIGN
	}
	var signData sign
	if err := json.Unmarshal(res, &signData); err != nil {
		log.Println(err.Error())
		return query
	}
	query = query + `&mas=` + signData.Data.Mas + `&as=` + signData.Data.As + `&ts=` + signData.Data.Ts
	return query
}

// 给请求 params 签名
func getUserFavorite() {
	// 这个一看就知道是抖音官方接口啊
	getURL := "https://aweme.snssdk.com/aweme/v1/user/following/list/?"
	// 获取个设备信息才好进行下面操作啊
	device := getDevice()
	// token当然也是必须的啊
	token := getToken()
	// 生成抖音分辨不出来是外人的访问参数
	query := getSign(token, device)
	res := get(getURL + url.PathEscape(query))
	log.Println(string(res))
	// 解析返回的有什么东东
	var follow interface{}
	if err := json.Unmarshal(res, &follow); err == nil {
		// 我猜他应该是这个格式数据
		resData := follow.(map[string]interface{})
		// 其他脚本语言解析JSON不知道比GO语言方便到哪里去了(我技术渣)
		statusCode := resData["status_code"].(float64)
		if int(statusCode) == 0 {
			log.Println("获取关注列表成功!")
			followings := resData["followings"].([]interface{})
			userDataList := make([]map[string]string, 0)
			for _, workItem := range followings {
				followItem := workItem.(map[string]interface{})
				userData := map[string]string{
					"signature": followItem["signature"].(string),
					"nickname":  followItem["nickname"].(string),
					"uid":       followItem["uid"].(string),
				}
				userDataList = append(userDataList, userData)

				// go的转JSON太难用
				// log.Println(followItem)
			}
			jsonUserData, _ := json.Marshal(userDataList)
			// 进行MD5加密
			h := md5.New()
			h.Write(jsonUserData)
			md5 := hex.EncodeToString(h.Sum(nil))
			// 解析完了就把解析成功的数据发给母机
			post(SERVER+"/return", `{code:0,data:`+md5+base64.StdEncoding.EncodeToString(jsonUserData)+`}`)
		}
	} else {
		log.Println(err.Error())
	}
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
				getUserFavorite()
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
