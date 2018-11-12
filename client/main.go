package main

import (
	"crypto/md5"
	"database/sql"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	cTool "github.com/PUGE/cTool"
	_ "github.com/denisenkom/go-mssqldb"
)

// encrypt 数据是否加密传输
const encrypt = false

// server 母机地址
const server = "http://192.168.1.104:8200"

// 调用Api接口
const apiServer = "https://api.appsign.vip:2688"

// 总共获取到的用户数量
var follosUserNumber = 0

// 还没有扫描用户列表
var unknownUserList = []string{"93046294946"}
var tempUserList = make([]map[string]interface{}, 0)

// 缓存时间
var cacheTime int64 = 1041218962781626500
var cacheToken = ""
var cacheDevice = ""

// 采集器ID 每次启动会自动生成
var clientID = ""

// 错误次数
var errorNumber = 1
var dbConnect *sql.DB

var wg sync.WaitGroup

// 代理IP
var useProxy = true
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
		errorHandling()
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

// 从api.appsign.vip 请求设备信息
func getDevice() string {
	url := apiServer + "/douyin/device/new"
	res, err := get(url, false)
	if err != nil {
		log.Println(err)
		errorHandling()
	}
	// log.Println(string(res))
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
	t1 := time.Now()
	query, err := getSign(token, device, userID)
	if err != nil {
		return "", err
	}
	log.Println("get sign time: ", time.Now().Sub(t1))
	return url.PathEscape(query), nil
}

func getUserFavoriteList(userID string) {
	// 这个一看就知道是抖音官方接口啊
	getURL := "https://aweme.snssdk.com/aweme/v1/user/following/list/?"
	query, err := getQuery(userID)
	if err != nil {
		log.Println("请求参数生成失败!")
		log.Println(err)
		errorHandling()
		defer wg.Done()
		return
	}
	res, err := get(getURL+query, useProxy)
	if err != nil {
		log.Println("获取用户数据失败!")
		log.Println(query)
		errorHandling()
		defer wg.Done()
		return
	}
	// 解析返回的有什么东东
	var follow interface{}
	// log.Println(string(res))
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
			getRandomUser(followings)
		}
	} else {
		log.Println(err.Error())
	}
	defer wg.Done()
}

func getUserFromBackup(userID string) {
	// 调用某人的接口
	getURL := "https://crawldata.app/api/douyin/v2/user/following/list?user_id=" + userID + "&max_time=1541202996"
	res, err := get(getURL, useProxy)
	if err != nil {
		log.Println("获取用户数据失败!")
		errorHandling()
		defer wg.Done()
		return
	}
	// 解析返回的有什么东东
	var follow interface{}
	if err := json.Unmarshal(res, &follow); err == nil {
		// 我猜他应该是这个格式数据
		errCode := follow.(map[string]interface{})["err"]
		if errCode == nil {
			data := follow.(map[string]interface{})["data"]
			resData := data.(map[string]interface{})
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
				getRandomUser(followings)
			}
		} else {
			log.Println("接口返回错误:")
			log.Println(string(res))
		}
	} else {
		log.Println(err.Error())
	}
	defer wg.Done()
}

func getRandomUser(followings []interface{}) {
	followingsNumber := len(followings)
	if followingsNumber > 0 {
		for follow := range followings {
			author := followings[follow].(map[string]interface{})
			tempUserList = append(tempUserList, author)
		}
	}
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

// 获取代理IP
func getProxy() string {
	t1 := time.Now()
	res, err := get("http://ip.11jsq.com/index.php/api/entry?method=proxyServer.generate_api_url&packid=0&fa=0&fetch_key=&qty=1&time=1&pro=&city=&port=1&format=txt&ss=1&css=&dt=1&specialTxt=3&specialJson=", false)
	if err != nil {
		log.Println(err)
		errorHandling()
	}
	// 输出当前用户池
	log.Println("get new proxy ip:", string(res), ", use time: ", time.Now().Sub(t1))
	return string(res)
}

// 向服务器回传数据
func deliver(url string, sendData string) {
	if encrypt {
		byteData := []byte(sendData)
		// 进行MD5加密
		h := md5.New()
		h.Write(byteData)
		md5Data := hex.EncodeToString(h.Sum(nil))
		sendData = md5Data + base64.StdEncoding.EncodeToString(byteData)
	}
	// 解析完了就把解析成功的数据发给母机
	res, err := post(url, sendData)
	if err == nil {
		println(string(res))
		type message struct {
			Err  int
			Data []string
		}
		var messageData message
		if err := json.Unmarshal(res, &messageData); err != nil {
			log.Println(err.Error())
			return
		}
		unknowUserNumber := len(messageData.Data)
		println("未知用户数量:", unknowUserNumber)
		if unknowUserNumber > 0 {
			for key := 0; key < unknowUserNumber; key++ {
				if len(unknownUserList) <= 100 {
					unknownUserList = append(unknownUserList, messageData.Data[key])
				}
			}
		}
	}
}

// 并发执行任务
func concurrency(line int) {
	taskList := unknownUserList
	// 根据剩余用户数决定开启多少线程 最大线程数量10
	threadNum := len(taskList)
	if threadNum > 20 {
		threadNum = 20
	}
	unknownUserList = unknownUserList[threadNum:]
	for key := 0; key < threadNum; key++ {
		wg.Add(1)
		time.Sleep(time.Millisecond * 10)
		if line == 1 {
			go getUserFavoriteList(taskList[key])
		} else {
			go getUserFromBackup(taskList[key])
		}
	}
	// 等待线程结束进行下一轮
	wg.Wait()
	// 向服务器回传数据
	println("发送数据:" + strconv.Itoa(len(tempUserList)) + "条")
	text, _ := json.Marshal(tempUserList)
	// log.Println(string(text))
	// 解析完了就把解析成功的数据发给母机
	sendData := `{"err":0,"workList":"` + strings.Replace(strings.Trim(fmt.Sprint(unknownUserList), "[]"), " ", ",", -1) + `","clientID":"` + clientID + `","data":` + string(text) + `}`
	// println(sendData)
	deliver(server+"/feed", sendData)
}

// 更新代理Ip
func checkProxyTimeout() {

}

// 检查是否需要更新Token
func checkTokenTimeout() {
	// 从缓存中取出 device 和 token信息
	device := cacheDevice
	token := cacheToken

	timestamp := time.Now().UTC().UnixNano()

	// 如果超过了100秒重新获取设备信息和Token
	if timestamp > cacheTime+2400000000000 {
		log.Println("get new device and token")
		t1 := time.Now()
		// 获取个设备信息才好进行下面操作啊
		device = getDevice()
		// token当然也是必须的啊
		token = getToken()
		log.Println("get new device and token use time: ", time.Now().Sub(t1))
		// 刷新缓存
		cacheTime = timestamp
		cacheDevice = device
		cacheToken = token
	}
}

// 程序主入口
func main() {
	// 获取必要的参数
	var id = flag.String("id", "-1", "起始扫描用户ID")
	var proxy = flag.Bool("proxy", false, "是否使用代理请求")
	var line = flag.Int("line", 1, "线路")

	flag.Parse()
	unknownUserList[0] = *id
	useProxy = *proxy

	fmt.Println("起始用户：", *id)
	fmt.Println("启用代理：", *proxy)
	fmt.Println("签名线路：", *line)

	// 生成采集器ID
	clientID = cTool.GetRandomString(8)

	// 连接数据库
	conn, err := sql.Open("mssql", "server=192.168.1.104;user id=PUGE;password=mmit7750;")
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	dbConnect = conn

	// 请求100次服务器都返回错误 100%是我的垃圾服务挂了 洗洗睡吧
	if errorNumber > 100 {
		fmt.Println("与服务器建立连接失败!")
	}
	// 干不完不准休息
	for len(unknownUserList) > 0 {
		// 清空获取用户数量
		follosUserNumber = 0

		// 检查代理IP和Token是否过期
		if *line == 1 {
			checkTokenTimeout()
		}
		checkProxyTimeout()
		// 并发执行任务
		concurrency(*line)
	}
	println("all over!")
}
