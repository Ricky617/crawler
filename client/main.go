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
	"sync"
	"time"

	_ "github.com/denisenkom/go-mssqldb"
)

// encrypt 数据是否加密传输
const encrypt = false

// server 母机地址
const server = "http://127.0.0.1:8000"

// 调用Api接口
const apiServer = "https://api.appsign.vip:2688"

// 总共获取到的用户数量
var follosUserNumber = 0

// 还没有扫描用户列表
var unknownUserList = []string{"79789629139"}

// 待发送用户列表
var tempUserList = []map[string]string{}

// 缓存时间
var cacheTime int64 = 1041218962781626500
var cacheToken = ""
var cacheDevice = ""

// 错误次数
var errorNumber = 1
var dbConnect *sql.DB

var wg sync.WaitGroup

// 代理IP
var useProxy = false
var proxyURL = "http://183.166.132.137:38747"

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
func get(requestURL string, useProxyGet bool) ([]byte, error) {
	client := &http.Client{}
	if useProxyGet {
		// fmt.Println(proxyURL)
		proxy, _ := url.Parse(proxyURL)
		tr := &http.Transport{
			Proxy: http.ProxyURL(proxy),
		}
		client = &http.Client{
			Transport: tr,
			Timeout:   time.Second * 15, //超时时间
		}
	}

	// fmt.Printf(requestURL)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return []byte(""), err
	}

	req.Header.Set("User-Agent", "Aweme/2.8.0 (iPhone; iOS 11.0; Scale/2.00)")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	resp, err := client.Do(req)
	if err != nil {
		return []byte(""), err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	return body, nil
}

// Post请求数据
func post(url string, data string) ([]byte, error) {
	resp, err := http.Post(url,
		"application/x-www-form-urlencoded",
		strings.NewReader(data))
	if err != nil {
		fmt.Println(err)
		return nil, err
	}

	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		// handle error
	}
	return body, nil
}

// 获取新的设备信息:有效期60分钟永久
func getToken() string {
	url := apiServer + "/token/douyin/version/2.7.0"
	res, err := get(url, false)
	if err != nil {
		log.Println(err)
		fmt.Println(strings.Replace(strings.Trim(fmt.Sprint(unknownUserList), "[]"), " ", ",", -1))
		log.Println("请求发生错误,休息一会, 10秒后重试")
		cacheTime = 0
		time.Sleep(time.Second * 10)
	}
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

func errorHandling() {
	log.Println(strings.Replace(strings.Trim(fmt.Sprint(unknownUserList), "[]"), " ", ",", -1))
	// 重新获取Token
	cacheTime = 0
	log.Println("请求发生错误,休息一会, 10秒后重试")
	time.Sleep(time.Second * 10)
}

// getDevice 请求设备信息
func getDevice() string {
	url := apiServer + "/douyin/device/new/version/2.7.0"
	res, err := get(url, false)
	if err != nil {
		log.Println(err)
		errorHandling()
	}
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
func getSign(token string, device string, userID string) (string, error) {
	url := apiServer + "/sign"
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()+3600000000000, 10)
	query := "user_id=" + userID + "&offset=0&count=49&source_type=2&max_time=" + timestamp[:10] + "&ac=WIFI&" + device + `&version_code=2.7.0&app_version=2.7.0&channel=App%20Stroe&app_name=aweme&build_number=27014&aid=1128`
	jsonData := `{"token":"` + token + `","query":"` + query + `"}`
	res, err := post(url, jsonData)
	if err != nil {
		return "", err
	}
	// log.Println(string(res))
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
		return query, nil
	}
	query = query + `&mas=` + signData.Data.Mas + `&as=` + signData.Data.As + `&ts=` + signData.Data.Ts
	return query, nil
}

// 生成请求参数
func getQuery(userID string) (string, error) {
	device := cacheDevice
	token := cacheToken
	// 生成访问参数
	query, err := getSign(token, device, userID)
	if err != nil {
		return "", err
	}
	return url.PathEscape(query), nil
}

func getUserFavoriteList(userID string) {
	// 这个一看就知道是抖音官方接口啊
	getURL := "https://aweme.snssdk.com/aweme/v1/user/following/list/?"
	query, err := getQuery(userID)
	if err != nil {
		return
	}
	res, err := get(getURL+query, useProxy)
	if err != nil {
		log.Println(err)
		errorHandling()
		defer wg.Done()
		return
	}
	// log.Println(string(res))
	// 解析返回的有什么东东
	var follow interface{}
	if err := json.Unmarshal(res, &follow); err == nil {
		// 我猜他应该是这个格式数据
		resData := follow.(map[string]interface{})
		// 待优化
		if resData["max_time"] == nil {
			log.Println("Incorrect return format")
			errorHandling()
			defer wg.Done()
			return
		}
		// maxTime := resData["max_time"].(float64)
		statusCode := resData["status_code"].(float64)
		if int(statusCode) == 0 {
			// 清洗数据
			followings := resData["followings"].([]interface{})
			// 计算数据库插入时间
			// t1 := time.Now()
			// log.Println("get user info number:" + strconv.Itoa(len(followings)))
			// 统计总共获取到的用户数量
			follosUserNumber += len(followings)
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
					tempUserList = append(tempUserList, userData)
					// fmt.Println(tempUserList)
					// 限制用户池大小防止溢出 上限200
					if len(unknownUserList) < 200 {
						// 添加到还没有扫描列表
						// fmt.Println(followItem["uid"].(string))
						unknownUserList = append(unknownUserList, followItem["uid"].(string))
					}
				}
			}
			// time := time.Now().Sub(t1)
			// 输出当前用户池
			// log.Println(time)
			// log.Println("------------------------- users pool -------------------------")
			// log.Println(unknownUserList)
			// log.Println("--------------------------------------------------------------")
		}
	} else {
		log.Println(err.Error())
	}
	defer wg.Done()
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

// 从数据库中查找用户是否存在
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

// 获取代理IP
func getProxy() string {
}

// 向服务器回传数据
func deliver(sendData string) {
	// 解析完了就把解析成功的数据发给母机
	post(server+"/return", sendData)
}

// 格式化关注列表 将获取到的关注列表包装成返回数据格式
func formatUserList() {
	needSendNumber := len(tempUserList)
	fmt.Println("send user number: " + strconv.Itoa(needSendNumber))
	if follosUserNumber > 0 {
		fmt.Println("work efficiency: " + strconv.Itoa((needSendNumber*100)/follosUserNumber) + "%")
	}
	// 有数据才发送向服务端发回数据
	if needSendNumber > 0 {
		favoriteList, _ := json.Marshal(tempUserList)
		sendData := string(favoriteList)
		if encrypt {
			// fmt.Print(string(sendData))
			// 进行MD5加密
			h := md5.New()
			h.Write(favoriteList)
			md5Data := hex.EncodeToString(h.Sum(nil))
			sendData = md5Data + base64.StdEncoding.EncodeToString(favoriteList)
		}
		// 解析完了就把解析成功的数据发给母机
		deliver(`{"err":0,"data":` + sendData + `}`)
	}
}

// 并发执行任务
func concurrency() {
	// 清空获取用户数量
	follosUserNumber = 0
	taskList := unknownUserList
	if len(taskList) > 10 {
		unknownUserList = unknownUserList[10:]
		for key := 0; key < 10; key++ {
			wg.Add(1)
			time.Sleep(time.Millisecond * 500)
			go getUserFavoriteList(taskList[key])
		}
	} else {
		unknownUserList = unknownUserList[0:0]
		for key := range taskList {
			wg.Add(1)
			time.Sleep(time.Millisecond * 500)
			go getUserFavoriteList(taskList[key])
		}
	}
	// 等待线程结束进行下一轮
	wg.Wait()
	// 任务完成后将用户关注信息传回服务端
	formatUserList()
}

// 检查是否需要更新Token
func checkTimeout() {
	// 从缓存中取出 device 和 token信息
	device := cacheDevice
	token := cacheToken
	// 清除待发送用户列表数据
	tempUserList = tempUserList[0:0]
	timestamp := time.Now().UTC().UnixNano()
	// 超过40秒更换新的代理IP
	if (timestamp > cacheTime+40000000000) && useProxy {
		proxyURL = "http://" + getProxy()
	}
	// 如果超过了100秒重新获取设备信息和Token
	if timestamp > cacheTime+100000000000 {
		log.Println("get new device and token")
		// 获取个设备信息才好进行下面操作啊
		device = getDevice()
		// token当然也是必须的啊
		token = getToken()
		// 刷新缓存
		cacheTime = timestamp
		cacheDevice = device
		cacheToken = token
	}
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
		// 定期获取Token和代理IP
		checkTimeout()
		// 并发执行任务
		concurrency()
	}
	println("all over!")
}
