# evethebot

A dumb telegram bot, using TiDB Cloud as backend storage

# Usage

1. Add the bot as your friend: @{your_telegram_bot_id}
2. Ask the bot for your token: /token

Then, you can access RESTful APIs:

1. Post message to your chat:

URL: http://{server addr you run your bot}/post

Request body:

```
{
    "token" : "{your token}"
    "msg" : "{your message}"
}
```


2. Get recent messages you sent to your bot

URL: http://{server addr you run your bot}/message

Request body:

```
{
    "token" : "{your token}"
    "limit" : 100 // 100 by default
}
```

# Example

Add `@eve_is_not_a_bot` and have a try:)

# So? What can we do?

I put some examples in demo directory.

1. A BTC-USD pricing live broadcast bot (with the help of crontab) See: [btcprice.py](https://github.com/c4pt0r/evethebot/blob/main/demo/btcprice.py)
2. A Hackernews bot, notify you when news appears on the HN front page for a specific keyword. See: [hnfilter.py](https://github.com/c4pt0r/evethebot/blob/main/demo/hnfilter.py)


