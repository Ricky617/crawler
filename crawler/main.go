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

	_ "github.com/go-sql-driver/mysql"
	"github.com/gosuri/uilive"
	"github.com/streadway/amqp"
)

// DB 数据库连接池
var DB *sql.DB

// 总共获取到的用户数量
var getNumber = 0

// 插入延迟
var insertDelay time.Duration

// Config 配置项
var Config map[string]interface{}

// 消息队列
var mqConn *amqp.Connection
var mqChannel *amqp.Channel

// 缓存时间
var proxyCacheTime int64 = 1041218962781626500
var cacheTime int64 = 1041218962781626500
var cacheToken = ""
var cacheDevice = ""

// 错误次数
var errorNumber = 1

var wg sync.WaitGroup

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

func errorHandling() {
	// 重新获取Token
	cacheTime = 0
	proxyCacheTime = 0
	if errorNumber > 5 {
		log.Println("发生错误次数过多,休息一会, 10分钟后重试")
		time.Sleep(time.Second * 600)
	} else if errorNumber > 7 {
		log.Println("发生错误次数过多,休息一会, 20分钟后重试")
		time.Sleep(time.Second * 1200)
	} else {
		log.Println("请求发生错误,休息一会, 10秒后重试")
		time.Sleep(time.Second * 10)
	}

}

// 注意方法名大写，就是public
func initDB() {
	t1 := time.Now()
	//构建连接："用户名:密码@tcp(IP:端口)/数据库?charset=utf8"
	path := strings.Join([]string{Config["dbUserName"].(string), ":", Config["dbPassword"].(string), "@tcp(", Config["dbIP"].(string), ":", Config["dbPort"].(string), ")/", Config["dbName"].(string), "?charset=utf8"}, "")

	//打开数据库,前者是驱动名，所以要导入： _ "github.com/go-sql-driver/mysql"
	DB, _ = sql.Open("mysql", path)
	//设置数据库最大连接数
	DB.SetConnMaxLifetime(100)
	//设置上数据库最大闲置连接数
	DB.SetMaxIdleConns(10)
	//验证连接
	if err := DB.Ping(); err != nil {
		fmt.Println("opon database fail")
		return
	}
	fmt.Println("connnect success:", time.Now().Sub(t1))
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

// 检查是否需要更新Token
func checkTokenTimeout() {
	// 从缓存中取出 device 和 token信息
	device := cacheDevice
	token := cacheToken

	timestamp := time.Now().UTC().UnixNano()

	// 如果超过了100秒重新获取设备信息和Token
	if timestamp > cacheTime+2400000000000 {
		// log.Println("get new device and token")
		// t1 := time.Now()
		// 获取个设备信息才好进行下面操作啊
		device = getDevice()
		// token当然也是必须的啊
		token = getToken()
		// log.Println("get new device and token use time: ", time.Now().Sub(t1))
		// 刷新缓存
		cacheTime = timestamp
		cacheDevice = device
		cacheToken = token
	}
}

func getWork() string {
	// 从队列获取消息
	msg, _, err := mqChannel.Get(
		Config["sourceQueue"].(string), // queue name
		true,                           // auto-ack
	)
	if nil != err {
		log.Fatalf("basic.consume source: %s", err)
	}
	return string(msg.Body)
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

// 发送信息
func sendMessage(message string, queue string) {
	// 发布消息
	err := mqChannel.Publish(
		"",
		queue,
		false,
		false,
		amqp.Publishing{
			ContentType: "text/plain",
			Body:        []byte(message),
		})

	// 错误处理
	if err != nil {
		log.Fatal("Failed to publish a message:", err.Error())
	}
}

// 将解析好的用户数据发至指定队列
func getRandomUser(followings []interface{}) {
	// log.Println(followings)
	followingsNumber := len(followings)
	if followingsNumber > 0 {
		for follow := range followings {
			author := followings[follow].(map[string]interface{})
			uid := author["uid"].(string)

			_, err := DB.Exec(`INSERT INTO user  (uid) VALUES (` + uid + `)`)
			// 如果插入失败证明数据库中已经存在了
			if err != nil {
				// fmt.Printf("insert data error: %v\n", err)
				return
			}
			// 如果不存在的话
			sendMessage(uid, "uncheck-id")
			sendMessage(uid, "douyin-unknow-id")
			getNumber = getNumber + 1

		}
	}
}

// 获取用户关注列表
func concurrency() {
	// 向数据库中插入数据
	t1 := time.Now()
	defer wg.Done()
	userID := getWork()
	if userID != "" {
		// 这个一看就知道是抖音官方接口啊
		query, err := getQuery(userID)
		// fmt.Println(getURL + query)
		if err != nil {
			log.Println("请求参数生成失败!")
			log.Println(err)
			errorHandling()
			return
		}
		res, err := get("https://aweme.snssdk.com/aweme/v1/user/following/list/?" + query)
		// fmt.Println(res)
		if err != nil {
			log.Println("获取用户数据失败!")
			log.Println(err)
			errorHandling()
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
				log.Println(follow)
				errorHandling()
				return
			}
			// maxTime := resData["max_time"].(float64)
			statusCode := resData["status_code"].(float64)
			if int(statusCode) == 0 {
				// 清洗数据
				followings := resData["followings"].([]interface{})
				// 保存数据
				// t1 := time.Now()
				getRandomUser(followings)
				// fmt.Println(time.Now().Sub(t1))
			}
		} else {
			log.Println(err.Error())
		}
	}
	insertDelay = time.Now().Sub(t1)
}

// 注册消息队列
func rabbit() {
	var err error
	// 注册消息队列
	mqConn, err := amqp.Dial(Config["sourceMQ"].(string))
	if err != nil {
		log.Fatal("Open connection failed:", err.Error())
	}
	mqChannel, err = mqConn.Channel()
	if err != nil {
		log.Fatal("Failed to open a channel:", err.Error())
	}
}

// 程序主入口
func main() {

	// 加载配置项
	configFile, err := ioutil.ReadFile("config.json")
	if err != nil {
		fmt.Print(err)
	}
	if err := json.Unmarshal(configFile, &Config); err != nil {
		log.Println("配置文件格式错误!")
		log.Println(err.Error())
	}
	fmt.Println("配置项加载成功!")

	// 链接数据库
	initDB()
	// 注册消息队列
	rabbit()

	writer := uilive.New()
	writer.Start()
	// 干不完不准休息
	for true {
		// 检查代理IP和Token是否过期
		checkTokenTimeout()
		for key := 0; key < int(Config["threadNum"].(float64)); key++ {
			wg.Add(1)
			// 并发执行任务
			go concurrency()
		}
		fmt.Fprintf(writer, "\n获取数量: %d 条\n插入延迟: %s", getNumber, insertDelay)
		wg.Wait()
	}
	writer.Stop()
	println("all over!")
}
