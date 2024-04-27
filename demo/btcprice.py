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

f = request.urlopen('https://api.coindesk.com/v1/bpi/currentprice/USD.json')
body = f.read().decode('utf-8')
f.close()

res = json.loads(body)

float_price = res['bpi']['USD']['rate_float']

payload = {'token': TOKEN, 'msg': f"BTC2USD: {float_price}"}
req = request.Request('http://aws.0xffff.me:8089/post', data=json.dumps(payload).encode('utf-8'))
request.urlopen(req)