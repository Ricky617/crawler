#!/usr/bin/env python
#--coding:utf-8--
import urllib
import json
import time
import pyodbc
import logging
from http.server import BaseHTTPRequestHandler, HTTPServer

logging.basicConfig(level=logging.INFO, format='%(asctime)s %(filename)s:%(lineno)d %(threadName)s:%(funcName)s %(levelname)s] %(message)s')

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
      logging.info('receive ' + str(len(userList)) +' user info!')
      logging.debug(userList)
      # 拼接SQL语句一次性插入
      sqlStr = 'INSERT INTO SIMPLE (DOUYIN_ID, NAME, SIGNA, BIRTHDAY, GET_TIME) VALUES '
      # 获取当前时间戳
      Time = time.time()
      for ind, val in enumerate(userList):
        # print(val)
        # 去除非法字符
        val['nickname'] = val['nickname'].replace('\n', '')
        val['nickname'] = val['nickname'].replace("'", "''")
        val['signature'] = val['signature'].replace('\n', '')
        val['signature'] = val['signature'].replace("'", "''")
        if (len(userList) - 1 == ind):
          sqlStr += "(" + val['uid'] +", '" + val['nickname'] +"', '" + val['signature'] +"', '" + val['birthday'] +"', " + str(int(Time)) +" );"
        else:
          sqlStr += "(" + val['uid'] +", '" + val['nickname'] +"', '" + val['signature'] +"', '" + val['birthday'] +"', " + str(int(Time)) +" ),"
      # print(sqlStr)
      # 插入数据库
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