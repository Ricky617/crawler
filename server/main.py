import json
import hashlib
from flask import Flask
app = Flask(__name__)

md5 = hashlib.md5(b'123').hexdigest()
print(md5)



jobList = ['1212', '121111']

@app.route('/getJob')
def index():
  return json.dumps(jobList)

if __name__ == '__main__':
  app.run(debug=True)