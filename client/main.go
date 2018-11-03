package main

import (
	"crypto/md5"
	"database/sql"
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

	_ "github.com/denisenkom/go-mssqldb"
)

// encrypt 数据是否加密传输
const encrypt = false

// server 母机地址
const server = "http://127.0.0.1:8000"

// 调用Api接口
const apiServer = "https://api.appsign.vip:2688"

// 还没有扫描用户列表
var unknownUserList = []string{"98448654828"}

// 缓存时间
var cacheTime int64 = 1041218962781626500
var cacheToken = ""
var cacheDevice = ""

// 错误次数
var errorNumber = 1
var dbConnect *sql.DB

// 当然是请求任务啦
func getJob() (string, int) {
	resp, err := http.Get(server)
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

// Get请求数据
func get(requestURL string) []byte {
	client := &http.Client{}
	// fmt.Printf(requestURL)
	req, err := http.NewRequest("GET", requestURL, nil)
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

// Post请求数据
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
	url := apiServer + "/token/douyin/version/2.7.0"
	res := get(url)
	type token struct {
		Success bool
		Token   string
	}
	var tokenData token
	if err := json.Unmarshal(res, &tokenData); err != nil {
		log.Println(err.Error())
		return ""
	}
	return tokenData.Token
}

// getDevice 请求设备信息
func getDevice() string {
	url := apiServer + "/douyin/device/new/version/2.7.0"
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
	if err := json.Unmarshal(res, &deviceData); err != nil {
		log.Println(err.Error())
		return ""
	}
	temp := deviceData.Data
	return `openudid=` + temp.Openudid + `&idfa=` + temp.Idfa + `&vid=` + temp.Vid + `&install_id=` + strconv.Itoa(temp.Install_id) + `&iid=` + strconv.Itoa(temp.Iid) + `&device_id=` + strconv.Itoa(temp.Device_id) + `&new_user=` + strconv.Itoa(temp.New_user) + `&device_type=` + temp.Device_type + `&os_version=` + temp.Os_version + `&osAPISERVER=` + temp.OsAPISERVER + `&screen_width=` + temp.Screen_width + `&device_platform=` + temp.Device_platform
}

// getSign 获取签名信息
func getSign(token string, device string, userID string) string {
	url := apiServer + "/sign"
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()+3600000000000, 10)
	query := "user_id=" + userID + "&offset=0&count=49&source_type=2&max_time=" + timestamp[:10] + "&ac=WIFI&" + device + `&version_code=2.7.0&app_version=2.7.0&channel=App%20Stroe&app_name=aweme&build_number=27014&aid=1128`
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

// 生成请求参数
func getQuery(userID string) string {
	device := cacheDevice
	token := cacheToken
	timestamp := time.Now().UTC().UnixNano()
	// 如果超过了1000秒重新获取设备信息和Token
	if timestamp > cacheTime+1000000000000 {
		// 获取个设备信息才好进行下面操作啊
		device = getDevice()
		// token当然也是必须的啊
		token = getToken()
		// 刷新缓存
		cacheTime = timestamp
		cacheDevice = device
		cacheToken = token
	} else {
		log.Println("use cache device and token")
	}
	// 生成访问参数
	query := getSign(token, device, userID)
	return url.PathEscape(query)
}

func getUserFavoriteList(userID string) []map[string]string {
	userDataList := make([]map[string]string, 0)
	// 这个一看就知道是抖音官方接口啊
	getURL := "https://aweme.snssdk.com/aweme/v1/user/following/list/?"
	query := getQuery(userID)
	res := get(getURL + query)
	// log.Println(string(res))
	// 解析返回的有什么东东
	var follow interface{}
	if err := json.Unmarshal(res, &follow); err == nil {
		// 我猜他应该是这个格式数据
		resData := follow.(map[string]interface{})
		statusCode := resData["status_code"].(float64)
		if int(statusCode) == 0 {
			// 获取完当前用户的关注信息后将当前用户移出用户池
			unknownUserList = unknownUserList[1:]
			// 清洗数据
			followings := resData["followings"].([]interface{})
			// 计算数据库插入时间
			t1 := time.Now()
			for _, workItem := range followings {
				followItem := workItem.(map[string]interface{})
				// 如果在数据库中已经存在的则不进行保存
				if !checkUserSaved(followItem["uid"].(string)) {
					userData := map[string]string{
						"signature": followItem["signature"].(string),
						"nickname":  followItem["nickname"].(string),
						"uid":       followItem["uid"].(string),
						"birthday":  followItem["birthday"].(string),
					}
					userDataList = append(userDataList, userData)
					// 限制用户池大小防止溢出 上限200
					if len(unknownUserList) < 200 {
						// 添加到还没有扫描列表
						unknownUserList = append(unknownUserList, followItem["uid"].(string))
					}
				}
			}
			time := time.Now().Sub(t1)
			// 输出当前用户池
			log.Println(time)
			log.Println("------------------------- users pool -------------------------")
			log.Println(unknownUserList)
			log.Println("--------------------------------------------------------------")
		}
	} else {
		log.Println(err.Error())
	}
	return userDataList
}

func work(workList string, userID string) {
	decodedData := []byte(workList)
	// 如果设置了加密需要先解密
	if encrypt {
		md5Data := workList[:32]
		data := workList[32:]
		// fmt.Print(data)
		decodedData, err := base64.StdEncoding.DecodeString(data)
		// 解析失败了那0.01%是伪造的请求, 99.9%是服务端挂了,那还执行毛任务啊歇着吧
		if err == nil {
			has := md5.Sum(decodedData)
			md5Check := fmt.Sprintf("%x", has)
			if md5Data == md5Check {
				fmt.Printf("数据MD5校验失败!")
				return
			}
		} else {
			fmt.Printf(err.Error())
			return
		}
	}
	// 开始解析服务器发来什么玩意
	var message interface{}
	if err := json.Unmarshal(decodedData, &message); err == nil {
		resData := message.(map[string]interface{})
		// fmt.Println(reflect.TypeOf(resData["data"]))
		workList := resData["data"].([]interface{})
		for _, workItem := range workList {
			fmt.Println(workItem)
		}
	}
}

// 获取抖音指定用户关注列表
func getDouYinFavorite(userID string) {
	// 获取到关注用户列表
	userFavoriteList := getUserFavoriteList(userID)
	favoriteList, _ := json.Marshal(userFavoriteList)
	sendData := string(favoriteList)
	if encrypt {
		// fmt.Print(string(sendData))
		// 进行MD5加密
		h := md5.New()
		h.Write(favoriteList)
		md5Data := hex.EncodeToString(h.Sum(nil))
		sendData = md5Data + base64.StdEncoding.EncodeToString(favoriteList)
	}
	log.Println("save user number:" + strconv.Itoa(len(userFavoriteList)))
	// print(`{"err":0,"data":` + sendData + `}`)
	// 解析完了就把解析成功的数据发给母机
	post(server+"/return", `{"err":0,"data":`+sendData+`}`)
}

// 获取工作的函数
func getWork(userID string) {
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
				work(data, userID)
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
		work(data, userID)
	}
}

func checkUserSaved(douyinID string) bool {
	var count int
	err := dbConnect.QueryRow("select isnull((select top(1) 1 from DouYin.dbo.SIMPLE where DOUYIN_ID = '" + douyinID + "'), 0)").Scan(&count)
	if err != nil {
		log.Fatal(err)
	}
	if count > 0 {
		return true
	}
	return false
}

func main() {
	// 连接数据库
	conn, err := sql.Open("mssql", "server=localhost;user id=PUGE;password=mmit7750;")
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	dbConnect = conn
	// 请求100次服务器都返回错误 100%是我的垃圾服务挂了 洗洗睡吧
	if errorNumber > 100 {
		// 问我为什么是中文输出(这样岂不是很没有逼格?), 我的回答是:要照顾每个使用者的智商(我其实也不懂英文啊!!!)
		fmt.Println("与服务器建立连接失败!")
	}
	// 干不完不准休息
	for len(unknownUserList) > 0 {
		getDouYinFavorite(unknownUserList[0])
	}
}
