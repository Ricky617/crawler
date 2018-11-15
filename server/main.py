#!/usr/bin/env python
#--coding:utf-8--
import urllib
import os
import json
import time
import pyodbc
import logging
from flask import Flask, request

app = Flask(__name__)

# 日志文件存储
# if not os.path.exists("./log/"):
#   os.mkdir("./log/")
# # Log输出配置
# logFileName = str(int(time.time())) + '.log'
# file = open('./log/' + logFileName , 'w' ,encoding='utf-8')
# file.close()
logging.basicConfig(
  level=logging.INFO,
  format='%(asctime)s: %(message)s'
  # handlers=[logging.FileHandler('./log/'+ logFileName, 'w', 'utf-8')]
)

info = {
  'gainTotal': 0,
  'clientList': {}
}

def saveUser(userList, conn):
  # 获取数据库指针
  c = conn.cursor()
  print('save ' + str(len(userList)) +' user info!')
  # logging.info(userList)
  # 拼接SQL语句一次性插入
  sqlStr = 'INSERT INTO [dbo].[USER] (uid, short_id, nickname, gender, signature, birthday, is_verified, follow_status, hide_search, constellation, hide_location, weibo_verify, custom_verify, unique_id, bind_phone, special_lock, need_recommend, is_binded_weibo, weibo_name, weibo_schema, weibo_url, story_open, room_id, live_verify, authority_status, verify_info, shield_follow_notice, shield_digg_notice, shield_comment_notice, school_name, school_poi_id, school_type, with_commerce_entry, verification_type, enterprise_verify_reason, is_ad_fake, region, account_region, commerce_user_level, live_agreement, secret, has_orders, prevent_download, unique_id_modify_time, ins_id, google_account, youtube_channel_id, youtube_channel_title, apple_account, with_fusion_shop_entry, is_phone_binded, accept_private_policy, twitter_id, twitter_name, user_canceled, has_email, is_gov_media_vip, live_agreement_time, status, create_time, avatar_uri, follower_status, neiguang_shield, comment_setting, duet_setting, reflow_page_gid, reflow_page_uid, user_rate, download_setting, download_prompt_ts, react_setting, live_commerce, language, get_time) VALUES '
  # 获取当前时间戳
  Time = time.time()
  for ind, val in enumerate(userList):
    # print(val)
    # 去除非法字符
    val['nickname'] = val['nickname'].replace('\n', '\\n')
    val['custom_verify'] = val['custom_verify'].replace("'", "’")
    val['enterprise_verify_reason'] = val['enterprise_verify_reason'].replace("'", "’")
    val['nickname'] = val['nickname'].replace("'", "’")
    val['signature'] = val['signature'].replace("'", "’")
    val['signature'] = val['signature'].replace('\n', '\\n')
    val['signature'] = val['signature'].replace("(", "\(")
    val['signature'] = val['signature'].replace(")", "\)")
    # print(val['language'])
    if 'language' not in val:
      val['language'] = ""
    # 如果最是最后一条则拼接以分号结尾的SQL语句
    sqlStr += "('" + val['uid'] + "', '" + val['short_id'] + "', '" + val['nickname'] + "', " + str(val['gender']) + ", '" + val['signature'] + "', '" + val['birthday'] + "', " + str(val['is_verified']) + ", " + str(val['follow_status']) + ", " + str(val['hide_search']) + ", " + str(val['constellation']) + ", " + str(val['hide_location']) + ", '" + val['weibo_verify'] + "', '" + val['custom_verify'] + "', '" + val['unique_id'] + "', '" + val['bind_phone'] + "', " + str(val['special_lock']) + ", " + str(val['need_recommend']) + ", " + str(val['is_binded_weibo']) + ", '" + val['weibo_name'] + "', '" + val['weibo_schema'] + "', '" + val['weibo_url'] + "', " + str(val['story_open']) + ", " + str(val['room_id']) + ", " + str(val['live_verify']) + ", " + str(val['authority_status']) + ", '" + val['verify_info'] + "', " + str(val['shield_follow_notice']) + ", " + str(val['shield_digg_notice']) + ", " + str(val['shield_comment_notice']) + ", '" + val['school_name'] + "', '" + val['school_poi_id'] + "', " + str(val['school_type']) + ", " + str(val['with_commerce_entry']) + ", " + str(val['verification_type']) + ", '" + val['enterprise_verify_reason'] + "', " + str(val['is_ad_fake']) + ", '" + val['region'] + "', '" + val['account_region'] + "', " + str(val['commerce_user_level']) + ", " + str(val['live_agreement']) + ", " + str(val['secret']) + ", " + str(val['has_orders']) + ", " + str(val['prevent_download']) + ", " + str(val['unique_id_modify_time']) + ", '" + val['ins_id'] + "', '" + val['google_account'] + "', '" + val['youtube_channel_id'] + "', '" + val['youtube_channel_title'] + "', " + str(val['apple_account']) + ", " + str(val['with_fusion_shop_entry']) + ", " + str(val['is_phone_binded']) + ", " + str(val['accept_private_policy']) + ", '" + val['twitter_id'] + "', '" + val['twitter_name'] + "', " + str(val['user_canceled']) + ", " + str(val['has_email']) + ", " + str(val['is_gov_media_vip']) + ", " + str(val['live_agreement_time']) + ", " + str(val['status']) + ", " + str(val['create_time']) + ", '" + val['avatar_uri'] + "', " + str(val['follower_status']) + ", " + str(val['neiguang_shield']) + ", " + str(val['comment_setting']) + ", " + str(val['duet_setting']) + ", " + str(val['reflow_page_gid']) + ", " + str(val['reflow_page_uid']) + ", " + str(val['user_rate']) + ", " + str(val['download_setting']) + ", " + str(val['download_prompt_ts']) + ", " + str(val['react_setting']) + ", " + str(val['live_commerce']) + ", '" + val['language'] + "', " + str(int(Time)) + ")"
    if (len(userList) - 1 == ind):
      sqlStr += ";"
    else:
      sqlStr += ","
  # 插入数据库
  # logging.info(sqlStr)
  sqlStr = sqlStr.replace('False', '0')
  sqlStr = sqlStr.replace('True', '1')
  # print(sqlStr)
  c.execute(sqlStr)
  conn.commit()
def saveSimple(userList, conn):
  # 获取数据库指针
  c = conn.cursor()
  print('save ' + str(len(userList)) +' simple user info!')
  # logging.info(userList)
  # 拼接SQL语句一次性插入
  sqlStr = 'INSERT INTO SIMPLE (DOUYIN_ID, NAME, SIGNA, BIRTHDAY, GET_TIME) VALUES '
  # 获取当前时间戳
  Time = time.time()
  saveUserList = {}
  # 保存添加到数据库数据的信息
  info["gainTotal"] += len(userList)
  # 数据清洗 清洗重复ID
  for value in userList:
    saveUserList[value["uid"]] = value
  for ind, uid in enumerate(saveUserList):
    val = saveUserList[uid]
    # 取出用户ID
    userId = val['uid']
    # print(val)
    # 去除非法字符
    val['nickname'] = val['nickname'].replace('\n', '\\n')
    val['nickname'] = val['nickname'].replace("'", "''")
    val['signature'] = val['signature'].replace('\n', '\\n')
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

@app.route('/monitor', methods=['GET'])
def monitor():
  sendData = {"err": 0, "total": info["gainTotal"]}
  return json.dumps(sendData)

@app.route('/push', methods=['POST'])
def push():
  
  # 开始处理时间
  start =time.clock()

  # 接收到的数据
  resData = json.loads(request.data)
  # 连接数据库
  conn = pyodbc.connect(r'DRIVER={SQL Server Native Client 11.0};SERVER=localhost;DATABASE=Douyin;UID=PUGE;PWD=mmit7750')
  # 获取数据库指针
  c = conn.cursor()
  unknowIdList = []
  # 检查重复键
  unknowUserList = []
  simpleUnknowUserList = []
  # 解析出用户列表数据
  # print(resData)
  userList = resData["data"]

  # 数据去重
  tempIdList = []
  newUserList = []
  for ind, val in enumerate(userList):
    if val["uid"] not in tempIdList:
      tempIdList.append(val["uid"])
      newUserList.append(val)
  # logging.info('receive ' + str(len(userList)) + ' user info!')
  for ind, val in enumerate(newUserList):
    # print(val)
    # 查询简单用户信息库
    c.execute("select isnull((select top(1) 1 from DouYin.dbo.SIMPLE where DOUYIN_ID = '" + val["uid"] + "'), 0)")
    row = c.fetchone()
    if (row[0] == 0):
      simpleUnknowUserList.append(val)
      unknowIdList.append(val["uid"])
    # 查询详细用户信息库
    c.execute("select isnull((select top(1) 1 from [dbo].[USER] where uid = '" + val["uid"] + "'), 0)")
    row = c.fetchone()
    if (row[0] == 0):
      unknowUserList.append(val)
  # 关闭数据库连接
  conn.commit()
  if (len(simpleUnknowUserList) > 0):
    saveSimple(simpleUnknowUserList, conn)
  if len(unknowUserList) > 0:
    saveUser(unknowUserList, conn)
  sendData = json.dumps({"err": 0, "data": unknowIdList})
  # logging.info('send data:' + sendData)
  conn.close()
  # 处理结束时间
  end = time.clock()
  print('Running time: %s Seconds'%(end - start))
  return sendData

if __name__ == '__main__':
  # 程序启动时向数据库查询数据总条数
  conn = pyodbc.connect(r'DRIVER={SQL Server Native Client 11.0};SERVER=localhost;DATABASE=Douyin;UID=PUGE;PWD=mmit7750')
  c = conn.cursor()
  startCount = c.execute('SELECT COUNT(*) from DouYin.dbo.SIMPLE').fetchone()
  print('Database data volume: ' + str(startCount[0]))
  info["gainTotal"] = startCount[0]
  conn.commit()
  conn.close()
  # 启动服务器
  app.run()