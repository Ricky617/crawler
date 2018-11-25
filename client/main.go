package main

import (
	"database/sql"
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
	"github.com/streadway/amqp"
)

type config struct {
	server   string // 服务器地址
	encrypt  bool   // 启用加密
	clientID string // 采集器ID
	thread   int    // 最大线程数量
}

var clientConfig config

// 总共获取到的用户数量
var follosUserNumber = 0

// 还没有扫描用户列表
var unknownUserList = []string{"85488042163"}

// 缓存时间
var proxyCacheTime int64 = 1041218962781626500
var cacheTime int64 = 1041218962781626500
var cacheToken = ""
var cacheDevice = ""

// 消息队列
var mqConn *amqp.Connection
var mqChannel *amqp.Channel

// 错误次数
var errorNumber = 1
var dbConnect *sql.DB

var wg sync.WaitGroup

// 代理IP
var proxyURL = "http://42.6.43.198:13267"

// Get请求数据
func get(requestURL string) ([]byte, error) {
	client := &http.Client{}

	// fmt.Printf(requestURL)
	req, err := http.NewRequest("GET", requestURL, nil)
	if err != nil {
		return []byte(""), err
	}

	req.Header.Set("User-Agent", "Aweme")
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
func post(requestURL string, data string) ([]byte, error) {
	client := &http.Client{}
	req, err := http.NewRequest("POST", requestURL, strings.NewReader(data))
	if err != nil {
		log.Println(err.Error())
		return []byte(""), err
	}

	req.Header.Set("Content-Type", "application/json")
	resp, err := client.Do(req)
	if err != nil {
		log.Println(err.Error())
		return []byte(""), err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		log.Println(err.Error())
		return []byte(""), err
	}
	return body, nil
}

// 获取新的设备信息:有效期60分钟永久
func getToken() string {
	url := "https://api.appsign.vip:2688/token/douyin/version/2.7.0"
	res, err := get(url)
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
	// 重新获取Token
	cacheTime = 0
	proxyCacheTime = 0
	if errorNumber > 5 {
		log.Println(strings.Replace(strings.Trim(fmt.Sprint(unknownUserList), "[]"), " ", ",", -1))
		log.Println("发生错误次数过多,休息一会, 10分钟后重试")
		time.Sleep(time.Second * 600)
	} else if errorNumber > 7 {
		log.Println(strings.Replace(strings.Trim(fmt.Sprint(unknownUserList), "[]"), " ", ",", -1))
		log.Println("发生错误次数过多,休息一会, 20分钟后重试")
		time.Sleep(time.Second * 1200)
	} else {
		log.Println("请求发生错误,休息一会, 10秒后重试")
		time.Sleep(time.Second * 10)
	}

}

// 从api.appsign.vip 请求设备信息
func getDevice() string {
	url := "https://api.appsign.vip:2688/douyin/device/new"
	res, err := get(url)
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
	return `&openudid=` + temp.Openudid + `&idfa=` + temp.Idfa + `&vid=` + temp.Vid + `&install_id=` + strconv.Itoa(temp.Install_id) + `&iid=` + strconv.Itoa(temp.Iid) + `&device_id=` + strconv.Itoa(temp.Device_id) + `&new_user=` + strconv.Itoa(temp.New_user) + `&device_type=` + temp.Device_type + `&screen_width=` + temp.Screen_width + `&device_platform=` + temp.Device_platform
}

// getSign 获取签名信息
func getSign(token string, device string, userID string) (string, error) {
	url := "http://127.0.0.1:8100/"
	timestamp := strconv.FormatInt(time.Now().UTC().UnixNano()+3600000000000, 10)
	query := "_rticket=1542368731370032509&ac=wifi&aid=1128&app_name=aweme&channel=360&count=49&device_brand=OnePlus&dpi=420&language=zh&manifest_version_code=169&max_time=" + timestamp[:10] + "&os_api=27&os_version=8.1.0&resolution=1080%2A1920&retry_type=no_retry&ssmix=a&update_version_code=1692&user_id=" + userID + "&uuid=615720636968612&version_code=169&version_name=1.6.9" + device
	res, err := post(url, query)
	if err != nil {
		return "", err
	}
	return string(res), nil
}

// 生成请求参数
func getQuery(userID string) (string, error) {
	device := cacheDevice
	token := cacheToken
	// 生成访问参数
	// t1 := time.Now()
	query, err := getSign(token, device, userID)
	if err != nil {
		return "", err
	}
	// log.Println("get sign time: ", time.Now().Sub(t1))
	return url.PathEscape(query), nil
}

func getUserFavoriteList(userID string) {
	// 这个一看就知道是抖音官方接口啊
	getURL := "http://api.amemv.com/aweme/v1/user/?"
	query, err := getQuery(userID)
	if err != nil {
		log.Println("请求参数生成失败!")
		log.Println(err)
		errorHandling()
		defer wg.Done()
		return
	}
	res, err := get(getURL + query)
	if err != nil {
		log.Println("获取用户数据失败!")
		log.Println(err)
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
		// // 待优化
		// if resData["max_time"] == nil {
		// 	log.Println(follow)
		// 	errorHandling()
		// 	defer wg.Done()
		// 	return
		// }
		// maxTime := resData["max_time"].(float64)
		statusCode := resData["status_code"].(float64)
		if int(statusCode) == 0 {
			// 清洗数据
			userData := resData["user"].(map[string]interface{})
			println(userData)
			getRandomUser(userData)
		} else {
			log.Println(follow)
			errorHandling()
			//defer wg.Done()
			return
		}
	} else {
		log.Println(err.Error())
	}
	// defer wg.Done()
}

func getRandomUser(followings map[string]interface{}) {
	// println(followings)
	text, _ := json.Marshal(followings)
	println(string(text))
	// // println(string(text))
	sendMessage(string(text))
}

var msg amqp.Delivery

func getWork() {
	queue, err := mqChannel.QueueDeclare("douyin-unknow-id-10000000000", false, false, false, false, nil)
	if err != nil {
		log.Fatalf("queue.declare source: %s", err)
	}
	unknownUserList = unknownUserList[0:0]
	msg, _, err = mqChannel.Get(
		queue.Name, // queue name
		false,      // auto-ack
	)
	if nil != err {
		log.Fatalf("basic.consume source: %s", err)
	}
	if len(msg.Body) != 0 {
		unknownUserList = append(unknownUserList, string(msg.Body))
	}
	// println(unknownUserList)
	// time.Sleep(time.Second * 1)
}

// 并发执行任务
func concurrency() {
	taskList := unknownUserList
	// 根据剩余用户数决定开启多少线程 最大线程数量10
	threadNum := len(taskList)
	if threadNum > clientConfig.thread {
		threadNum = clientConfig.thread
	}
	unknownUserList = unknownUserList[threadNum:]
	for key := 0; key < threadNum; key++ {
		wg.Add(1)
		time.Sleep(time.Millisecond * 10)
		go getUserFavoriteList(taskList[key])
	}
	// 等待线程结束进行下一轮
	wg.Wait()
	getWork()
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

func output() {
	fmt.Printf("采集器ID: %s\n\r线程数量: %d\n\r", clientConfig.clientID, clientConfig.thread)
	time.Sleep(time.Second * 10)
}

func rabbit() {
	var err error
	// 注册消息队列
	mqConn, err = amqp.Dial("amqp://admin:admin@127.0.0.1:5672/")
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}

	mqChannel, err = mqConn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err.Error())
	}

	_, err = mqChannel.QueueDeclare(
		"check-id-10000000000",
		false,
		false,
		false,
		false,
		nil,
	)
	if err != nil {
		log.Fatal("Failed to declare a queue:", err.Error())
	}
}

func sendMessage(message string) {
	err := mqChannel.Publish(
		"",
		"check-id-10000000000",
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})
	if err != nil {
		log.Fatal("Failed to publish a message:", err.Error())
	} else {
		msg.Ack(false)
	}
}

// 程序主入口
func main() {
	rabbit()
	for true {
		getWork()
		checkTokenTimeout()
		getUserFavoriteList(unknownUserList[0])
	}
	// // // 获取必要的参数
	// var id = flag.String("id", "-1", "起始扫描用户ID")
	// var threadNum = flag.Int("thread", 10, "线程数量")
	// flag.Parse()
	// unknownUserList[0] = *id

	// // defer mqConn.Close()
	// // defer mqChannel.Close()
	// // 注册消息队列
	// rabbit()

	// clientConfig.server = "http://127.0.0.1:5000"
	// clientConfig.thread = *threadNum
	// clientConfig.encrypt = false
	// // 生成采集器ID
	// clientConfig.clientID = cTool.GetRandomString(8)
	// output()
	// fmt.Println("起始用户：", *id)
	// // 连接数据库
	// conn, err := sql.Open("mssql", "server=127.0.0.1;user id=PUGE;password=mmit7750;")
	// if err != nil {
	// 	log.Fatal("Open connection failed:", err.Error())
	// }
	// dbConnect = conn

	// // 请求100次服务器都返回错误 100%是我的垃圾服务挂了 洗洗睡吧
	// if errorNumber > 100 {
	// 	fmt.Println("与服务器建立连接失败!")
	// }
	// // 干不完不准休息
	// for len(unknownUserList) > 0 {
	// 	// 清空获取用户数量
	// 	follosUserNumber = 0

	// 	// 检查代理IP和Token是否过期
	// 	checkTokenTimeout()
	// 	// 并发执行任务
	// 	concurrency()
	// }
	// println("all over!")
}
