# evethebot

A handy telegram bot for building other bots

# Usage

1. Add the bot as friend: `@{your_telegram_bot_id}`
2. Ask the bot for your token: `/token`

Then you can access APIs (with your token):

1. Post message to your chat:

URL: http://{server_addr_you_run_your_bot}/post

Request body:

```
{
    "token" : "{your token}"
    "msg" : "{your message}"
}
```

Example:

`$ curl -X POST http://0xffff.me:8089/post -d '{"token":"{your token}","msg":"*Hello* World"}'`

2. Get recent messages you sent to your bot

URL: http://{server_addr_you_run_your_bot}/message

Request body:

```
{
    "token" : "{your token}"
    "limit" : 100 // by default
}
```
Example:

`$ curl -X GET http://0xffff.me:8089/message -d '{"token":"{your token}"}' | jq .`

# Example

Add `@eve_is_not_a_bot` and have a try:)

*@eve_is_not_a_bot is using [TiDB Cloud](https://tidbcloud.com) (Free Tier) as backend storage*


# So? What can we do?

I put 2 examples in `demo` directory:

1. A BTC-USD pricing live broadcast bot (with the help of crontab) See: [btcprice.py](https://github.com/c4pt0r/evethebot/blob/main/demo/btcprice.py)
2. A Hackernews bot, notify you when news appears on the HN front page for a specific keyword. See: [hnfilter.py](https://github.com/c4pt0r/evethebot/blob/main/demo/hnfilter.py)


