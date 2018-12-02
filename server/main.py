#!/usr/bin/env python
#--coding:utf-8--
import urllib
import os
import json
import time
import pika
import time
import logging
import pymysql.cursors
from flask import Flask, request

app = Flask(__name__)

# 加载配置文件
config = json.load(open("config.json", encoding='utf-8'))

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

connection = pymysql.connect(host=config["dataBase"]["server"], port=config["dataBase"]["port"], user=config["dataBase"]["user"], password=config["dataBase"]["password"], db=config["dataBase"]["name"], charset='utf8mb4', cursorclass=pymysql.cursors.DictCursor)


def clearData (val):
  if val["nickname"] == "已重置":
    return ""
  if 'uid' not in val:
    return ""
  if 'custom_verify' not in val:
    val['custom_verify'] = ''
  if 'enterprise_verify_reason' not in val:
    val['enterprise_verify_reason'] = ''
  if 'signature' not in val:
    val['signature'] = ''
  if 'college_name' not in val:
    val['college_name'] = ''
  # 去除非法字符
  val['nickname'] = val['nickname'].replace('\n', '\\n')
  val['custom_verify'] = val['custom_verify'].replace("'", "’")
  val['enterprise_verify_reason'] = val['enterprise_verify_reason'].replace("'", "’")
  val['nickname'] = val['nickname'].replace("'", "’")
  val['nickname'] = val['nickname'].replace("\\", '\\\\')
  val['signature'] = val['signature'].replace("\\", '\\\\')
  val['signature'] = val['signature'].replace("'", "’")
  val['province'] = val['province'].replace("'", "\\'")
  val['location'] = val['location'].replace("'", "\\'")
  val['district'] = val['district'].replace("'", "\\'")
  val['city'] = val['city'].replace("'", "\\'")
  val['college_name'] = val['college_name'].replace("'", "’")
  val['signature'] = val['signature'].replace('\n', '\\n')
  val['signature'] = val['signature'].replace("(", "\(")
  val['signature'] = val['signature'].replace(")", "\)")
  # print(val)
  if 'college_name' not in val:
    val['college_name'] = ''
  if 'enroll_year' not in val:
    val['enroll_year'] = ''
  if 'school_name' not in val:
    val['school_name'] = ''
  val['school_name'] = val['school_name'].replace("'", "\\'")
  # 如果最是最后一条则拼接以分号结尾的SQL语句
  sqlStr = (
    val['uid'], val['authority_status'], val['avatar_uri'], val['aweme_count'], val['birthday'], val['city'], val['college_name'], val['commerce_user_level'], val['constellation'], val['country'], 
    val['custom_verify'], val['district'], val['dongtai_count'], val['enroll_year'], val['enterprise_verify_reason'], val['favoriting_count'], val['follower_count'], val['following_count'], val['gender'], 
    val['has_orders'], val['hide_location'], val['hide_search'], val['is_ad_fake'], val['is_binded_weibo'], val['is_discipline_member'], val['is_flowcard_member'], val['is_gov_media_vip'], val['is_verified'], val['iso_country_code'], 
    val['language'], val['live_agreement'], val['live_commerce'], 
    val['live_verify'], val['location'], val['mplatform_followers_count'], val['need_recommend'], val['neiguang_shield'], val['nickname'], val['prevent_download'], val['profile_tab_type'], val['province'], 
    
    val['react_setting'], val['realname_verify_status'], val['recommend_reason_relation'], val['reflow_page_gid'], val['reflow_page_uid'], val['region'], val['room_id'], val['school_name'], val['school_poi_id'], val['school_type'], 
    val['secret'], val['share_qrcode_uri'], val['shield_comment_notice'], val['shield_digg_notice'], val['shield_follow_notice'], val['short_id'], val['show_gender_strategy'], val['signature'], val['special_lock'], 
    val['star_use_new_download'], val['status'], val['story_count'], val['total_favorited'], val['unique_id'], val['unique_id_modify_time'], val['user_canceled'],
    val['user_rate'], val['verification_type'], val['verify_info'], val['weibo_name'], val['weibo_schema'], val['weibo_url'],
    val['with_commerce_entry'], val['with_commerce_newbie_task'], val['with_item_commerce_entry'])
  return sqlStr



def saveUser(val, ch, method):
  # 拼接SQL语句一次性插入
  with connection.cursor() as cursor:
    # print(val)
    # 插入数据库
    # print(sqlStr)
    cursor.executemany('INSERT IGNORE INTO `' + config["dataBase"]["table"] +'` VALUES (%s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,    %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,  %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,  %s, %s, %s, %s, %s, %s, %s, %s, %s,  %s, %s, %s, %s, %s, %s, %s, %s, %s,  %s, %s, %s, %s, %s, %s, %s, %s, %s, %s,  %s, %s, %s)', val)
    # 没有设置默认自动提交，需要主动提交，以保存所执行的语句
    connection.commit()
    cursor.close()
    # print('sd')
    # connection.close()
    
  # except:
  #   print('拒绝')
  #   # 拒绝消息
  #   ch.basic_recover(True)  #发送ack消息

@app.route('/monitor', methods=['GET'])
def monitor():
  sendData = {"err": 0, "total": info["gainTotal"]}
  return json.dumps(sendData)

parameters = pika.URLParameters('amqp://admin:admin@127.0.0.1:5672/')
mqConnection = pika.BlockingConnection(parameters)

unCheckChannel = mqConnection.channel()
unCheckChannel.basic_qos(prefetch_size=0, prefetch_count=100, all_channels=True) # 公平消费

tempUserList = []
def callback(ch, method, properties, body):
  # 接收到的数据
  val = json.loads(body.decode('utf-8'))
  # print(val)
  userData = clearData(val)
  if userData != "":
    if len(tempUserList) < 50:
      tempUserList.append(userData)
    else:
      # print(tempUserList)
      saveUser(tempUserList, ch, method)
      tempUserList.clear()
  ch.basic_ack(delivery_tag = method.delivery_tag)  #发送ack消息

  

if __name__ == '__main__':
  
  unCheckChannel.basic_consume(callback, queue='check-id')
  print(' [*] Waiting for messages. To exit press CTRL+C')
  unCheckChannel.start_consuming()    #开始监听 接受消息

  # 启动服务器
  # app.run()