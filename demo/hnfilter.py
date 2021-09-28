#!/usr/bin/env python3

import os
import sys
import json
import sqlite3
import time
from urllib import request
from concurrent.futures import as_completed
from concurrent.futures import ThreadPoolExecutor
from requests_futures.sessions import FuturesSession

HN_API_TOPSTORIES = 'https://hacker-news.firebaseio.com/v0/topstories.json'
HN_API_GETITEM = 'https://hacker-news.firebaseio.com/v0/item/%d.json'

TOKEN = os.getenv('TOKEN')
if TOKEN is None:
    print('please set token')
    sys.exit(-1)

DB_PATH = os.getenv('DB') 
if DB_PATH is None:
    DB_PATH = '.hnfilter.db'


if len(sys.argv) != 2:
    print('usage: ./hnfilter.py [keyword]')
    print('filter the stories with [keyword] in title or url')
    print('example: ./hnfilter.py github')
    sys.exit(-1)


con = sqlite3.connect(DB_PATH)
def init_db():
    try:
        cur = con.cursor()
        cur.execute('''CREATE TABLE stories (id INTEGER PRIMARY KEY, title TEXT, url TEXT, time INTEGER)''')
        cur.execute('''CREATE INDEX idx_id ON stories (id);''')
        cur.execute('''CREATE INDEX idx_time ON stories (time);''')
    except:
        pass

init_db()

def insert_story(story_id, title, url, timestamp):
    cur = con.cursor()
    cur.execute("insert into stories values (?, ?, ?, ?)", (story_id, title, url, timestamp))
    con.commit()


def is_story_seen(story_id):
    cur = con.cursor()
    cur.execute("select * from stories WHERE id = %d" % story_id)
    ret = cur.fetchall()
    if len(ret) > 0:
        return True
    return False

session = FuturesSession(executor=ThreadPoolExecutor(max_workers=10))
def fetch_top_stories(story_filter):
    f = request.urlopen(HN_API_TOPSTORIES)
    body = f.read().decode('utf-8')
    f.close()
    res = json.loads(body)

    new_matches = []

    print("loading top stories...")
    futures=[]
    for story_id in res:
        future = session.get(HN_API_GETITEM % story_id)
        future.id = story_id
        futures.append(future)

    print("loading top stories...done, found %d" % len(res))
    for future in as_completed(futures):
        resp = future.result()
        if not is_story_seen(future.id):
            s = resp.json()
            if s['type'] == 'story' and 'url' in s and story_filter(s):
                print("find match story")
                print(s)
                insert_story(s['id'], s['title'], s['url'], s['time'])
                new_matches.append(s)
    return new_matches

def send_message(stories):
    def chunks(lst, n):
        for i in range(0, len(lst), n):
            yield lst[i:i + n]

    for c in chunks(stories, 10):
        output = '👀💗🤖 Top HN Stories with Keyword: %s\n\n' % keyword
        for s in c: 
            output += '👾 [%s](%s)\n\n' % (s['title'], s['url'])
        payload = {'token': TOKEN, 'msg': output}
        req = request.Request('http://0xffff.me:8089/post', data=json.dumps(payload).encode('utf-8'))
        request.urlopen(req)
        time.sleep(1)
        
keyword = sys.argv[1]
new_matches = fetch_top_stories(lambda s: keyword in s['url'].lower() or keyword in s['title'].lower())
send_message(new_matches)

