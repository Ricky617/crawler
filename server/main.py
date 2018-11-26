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

connection = pymysql.connect(host=config["dataBase"]["server"], port=config["dataBase"]["port"], user=config["dataBase"]["user"], password=config["dataBase"]["password"], db=config["dataBase"]["name"], charset='utf8mb4', cursorclass=pymysql.cursors.DictCursor)

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

def saveUser(val):
  # print('save ' + str(len(userList)) +' user info!')
  # logging.info(userList)
  # 拼接SQL语句一次性插入
  with connection.cursor() as cursor:
    sqlStr = 'INSERT IGNORE INTO `1` VALUES '
    # print(val)
    if 'uid' not in val:
      return
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
    # 如果最是最后一条则拼接以分号结尾的SQL语句
    sqlStr += "(%s, %d, '%s', %d, %d, '%s', %d, '%s', '%s', '%s',  %d, %d, %d, '%s', '%s', '%s', %d, %d, %d, %d,  %d, '%s', '%s', %d, %d, %d, %d, %d, %d, %d,  %d, %d, %d, %d, %d, %d, %d, %d, %d, %d,  %d, %d, %d, %d, %d, '%s', '%s', %d, %d, %d,  %d, '%s', %d, %d, %d, %d, '%s', %d, %d, '%s',  %d, %d, '%s', %d, %d, '%s', %d, '%s', '%s', %d,  %d, '%s', %d, %d, %d, '%s', %d, %d, '%s', %d,  %d, %d, %d, %d, %d, %d, '%s', %d, %d, %d,   %d, %d, %d, '%s', '%s', %d, '%s', '%s', '%s', '%s',  %d, %d, %d, %d, %d, %d, %d, %d );" % (val['uid'], val['accept_private_policy'], val['account_region'], val['apple_account'], val['authority_status'], val['avatar_uri'], val['aweme_count'], val['birthday'], val['city'], val['college_name'],

    val['comment_setting'], val['commerce_user_level'], val['constellation'], val['country'], val['custom_verify'], val['district'], val['dongtai_count'], val['dou_plus_share_location'], val['download_prompt_ts'], val['download_setting'], 

    val['duet_setting'], val['enroll_year'], val['enterprise_verify_reason'], val['favoriting_count'], val['fb_expire_time'], val['follow_status'], val['follower_count'], val['follower_status'], val['following_count'], val['gender'], 
    
    val['has_activity_medal'], val['has_email'], val['has_insights'], val['has_orders'], val['hide_location'], val['hide_search'], val['is_ad_fake'], val['is_binded_weibo'], val['is_block'], val['is_discipline_member'], 
    
    val['is_effect_artist'], val['is_flowcard_member'], val['is_gov_media_vip'], val['is_phone_binded'], val['is_verified'], val['iso_country_code'], val['language'], val['live_agreement'], val['live_agreement_time'], val['live_commerce'], 
    
    val['live_verify'], val['location'], val['login_platform'], val['mplatform_followers_count'], val['need_recommend'], val['neiguang_shield'], val['nickname'], val['prevent_download'], val['profile_tab_type'], val['province'], 
    
    val['react_setting'], val['realname_verify_status'], val['recommend_reason_relation'], val['reflow_page_gid'], val['reflow_page_uid'], val['region'], val['room_id'], val['school_name'], val['school_poi_id'], val['school_type'], 
    
    val['secret'], val['share_qrcode_uri'], val['shield_comment_notice'], val['shield_digg_notice'], val['shield_follow_notice'], val['short_id'], val['show_gender_strategy'], val['show_image_bubble'], val['signature'], val['special_lock'], 
    
    val['star_use_new_download'], val['status'], val['story_count'], val['story_open'], val['sync_to_toutiao'], val['total_favorited'], val['unique_id'], val['unique_id_modify_time'], val['user_canceled'], val['user_mode'], 
    
    val['user_period'], val['user_rate'], val['verification_type'], val['verify_info'], val['video_icon_virtual_URI'], val['watch_status'], val['weibo_name'], val['weibo_schema'], val['weibo_url'], val['weibo_verify'], 
    
    val['with_commerce_entry'], val['with_commerce_newbie_task'], val['with_dou_entry'], val['with_douplus_entry'], val['with_fusion_shop_entry'], val['with_item_commerce_entry'], val['with_new_goods'], val['with_shop_entry'])

    # 插入数据库
    # logging.info(sqlStr)
    # sqlStr = sqlStr.replace('False', '0')
    # sqlStr = sqlStr.replace('True', '1')
    # print(sqlStr)
    cursor.execute(sqlStr)
    # 没有设置默认自动提交，需要主动提交，以保存所执行的语句
    connection.commit()

@app.route('/monitor', methods=['GET'])
def monitor():
  sendData = {"err": 0, "total": info["gainTotal"]}
  return json.dumps(sendData)

parameters = pika.URLParameters('amqp://admin:admin@127.0.0.1:5672/')
mqConnection = pika.BlockingConnection(parameters)

unCheckChannel = mqConnection.channel()
unCheckChannel.basic_qos(prefetch_size=0, prefetch_count=20, all_channels=True) # 公平消费
unCheckChannel.queue_declare(queue='douyin-uncheck')

def callback(ch, method, properties, body):
  '''回调函数,处理从rabbitmq中取出的消息'''
  # print(" [x] Received %r" % body)
  # time.sleep(1)
  
  # 接收到的数据
  val = json.loads(body.decode('utf-8'))
  saveUser(val)

  ch.basic_ack(delivery_tag = method.delivery_tag)  #发送ack消息

  

if __name__ == '__main__':
  # 程序启动时向数据库查询数据总条数
  # conn = pyodbc.connect(r'DRIVER={SQL Server Native Client 11.0};SERVER=localhost;DATABASE=Douyin;UID=PUGE;PWD=mmit7750')
  # c = conn.cursor()
  # startCount = c.execute('SELECT COUNT(*) from DouYin.dbo.SIMPLE').fetchone()
  # print('Database data volume: ' + str(startCount[0]))
  # info["gainTotal"] = startCount[0]
  # conn.commit()
  # conn.close()
  
  unCheckChannel.basic_consume(callback, queue='check-id-10000000000', no_ack=False)
  print(' [*] Waiting for messages. To exit press CTRL+C')
  unCheckChannel.start_consuming()    #开始监听 接受消息
  conn.close()

  # 启动服务器
  # app.run()