#!/usr/bin/env python
#--coding:utf-8--
import urllib
import os
import json
import time
import pyodbc
import logging
from http.server import BaseHTTPRequestHandler, HTTPServer

if not os.path.exists("./log/"):
  os.mkdir("./log/")
# Log输出配置
logFileName = str(int(time.time())) + '.log'
file = open('./log/' + logFileName , 'w' ,encoding='utf-8')
file.close()
logging.basicConfig(
  level=logging.INFO,
  format='%(asctime)s: %(message)s',
  handlers=[logging.FileHandler('./log/'+ logFileName, 'w', 'utf-8')]
)

jobList = ['1212', '121111']
class testHTTPServer_RequestHandler(BaseHTTPRequestHandler):
  # GET
  def do_GET(self):
    sendData = {"err": 0, "data": jobList}
    self.outputtxt(json.dumps(sendData))

  def do_POST(self):
    content_length = int(self.headers['Content-Length']) # <--- Gets the size of data
    post_data = self.rfile.read(content_length) # <--- Gets the data itself
    logging.debug("POST request,\nPath: %s\nHeaders:\n%s\n\nBody:\n%s\n", str(self.path), str(self.headers), post_data.decode('utf-8'))
    # print(resData)
    # 接收到的数据
    resData = json.loads(post_data.decode('utf-8'))
    # 解析出用户列表数据
    userList = resData["data"]
    if (resData["err"] == 0 and len(userList) > 0):
      # 连接数据库
      conn = pyodbc.connect(r'DRIVER={SQL Server Native Client 11.0};SERVER=localhost;DATABASE=Douyin;UID=PUGE;PWD=mmit7750')
      # 获取数据库指针
      c = conn.cursor()
      print('receive ' + str(len(userList)) +' user info!')
      # logging.info(userList)
      # 拼接SQL语句一次性插入
      sqlStr = 'INSERT INTO SIMPLE (DOUYIN_ID, NAME, SIGNA, BIRTHDAY, GET_TIME) VALUES '
      # 获取当前时间戳
      Time = time.time()
      saveUserList = {}
      # 数据清洗 清洗重复ID
      for value in userList:
        saveUserList[value["uid"]] = value
      for ind, uid in enumerate(saveUserList):
        val = saveUserList[uid]
        # 取出用户ID
        userId = val['uid']
        # print(val)
        # 去除非法字符
        val['nickname'] = val['nickname'].replace('\n', '')
        val['nickname'] = val['nickname'].replace("'", "''")
        val['signature'] = val['signature'].replace('\n', '')
        val['signature'] = val['signature'].replace("'", "''")
        # 如果最是最后一条则拼接以分号结尾的SQL语句
        if (len(saveUserList) - 1 == ind):
          sqlStr += "(" + userId +", '" + val['nickname'] +"', '" + val['signature'] +"', '" + val['birthday'] +"', " + str(int(Time)) +" );"
        else:
          sqlStr += "(" + userId +", '" + val['nickname'] +"', '" + val['signature'] +"', '" + val['birthday'] +"', " + str(int(Time)) +" ),"
      # 插入数据库
      # logging.info(sqlStr)
      c.execute(sqlStr)
      conn.commit()
      conn.close()
    self.outputtxt(json.dumps({"err": 0}))
 
  def outputtxt(self, content):
    self.send_response(200)
    self.send_header('Content-type', 'application/json')
    self.end_headers()
    self.wfile.write(bytes(content, "utf-8"))

def run():
  port = 8000
  print('starting server, port', port)

  # Server settings
  server_address = ('0.0.0.0', port)
  httpd = HTTPServer(server_address, testHTTPServer_RequestHandler)
  print('running server...')
  httpd.serve_forever()

if __name__ == '__main__':
  run()