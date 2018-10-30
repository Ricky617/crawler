import json
import hashlib
import base64
from flask import Flask
app = Flask(__name__)

# md5 = hashlib.md5(b'123').hexdigest()
# print(md5)

# 加密
def encryption (original):
  md5Data = hashlib.md5(original).hexdigest()
  base64Data = base64.b64encode(original)
  return md5Data + str(base64Data, encoding = "utf8")

jobList = ['1212', '121111']

@app.route('/getJob')
def index():
  sendData = {"err": 0, "data": jobList}
  return encryption(json.dumps(sendData).encode())

if __name__ == '__main__':
  app.run(debug=True)