#!/usr/bin/env python3

# Usage:
# $ export TOKEN={your token from eve bot}
# $ ./btcprice.py

import os
import sys
import json
from urllib import request


TOKEN = os.getenv('TOKEN')
if TOKEN is None:
    print('please set token')
    sys.exit(-1)

f = request.urlopen('https://api.binance.com/api/v3/ticker/24hr?symbol=BTCUSDT')
body = f.read().decode('utf-8')
f.close()

res = json.loads(body)


output = '*💰BTC Price (To USDT)*\n\n'
for k,v in res.items():
    output += '*' + k + '*' + ': ' + str(v) + '\n'

payload = {'token': TOKEN, 'msg': output}
req = request.Request('http://0xffff.me:8089/post', data=json.dumps(payload).encode('utf-8'))
request.urlopen(req)

