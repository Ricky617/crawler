package main

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"strconv"
	"time"
)

// 消息结构体
type message struct {
	err int
	msg string
}

const SERVER = "http://127.0.0.1:5000"

var errorNumber int = 1

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

func main() {
	// 请求100次服务器都返回错误 100%是我的垃圾服务挂了 洗洗睡吧
	if errorNumber > 100 {
		// 问我为什么是中文输出(这样岂不是很没有逼格?), 我的回答是:要照顾每个使用者的智商(我其实也不懂英文啊!!!)
		fmt.Println("与服务器建立连接失败!")
	}
	// 没有找到工作要你有毛用啊 继续找
	data, err := getJob()
	if err != 0 {
		for true {
			time.Sleep(time.Duration(1) * time.Second)
			data, err := getJob()
			if err == 0 {
				println("第" + strconv.Itoa(errorNumber) + "次尝试与服务器建立连接失败,1分钟后重试!")
				errorNumber++
				fmt.Printf(data, err)
				break
			}
			// 请求100次服务器都返回错误 100%是我的垃圾服务挂了 洗洗睡吧
			if errorNumber > 10 {
				// 问我为什么是中文输出(这样岂不是很没有逼格?), 我的回答是:要照顾每个使用者的智商(我其实也不懂英文啊!!!)
				fmt.Println("与服务器建立连接失败!")
				break
			}
			// 没有找到工作要你有毛用啊 继续找
		}
	}
	fmt.Printf("sdds")
	fmt.Printf(data, err)
}
