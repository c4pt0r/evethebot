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

Add @eve_is_not_a_bot  :)

# So? What can we do?

I put some examples in demo directory.

1. A BTC-USD pricing live broadcast bot (with the help of crontab)
2. A Hackernews bot, notify you when news appears on the HN front page for a specific keyword.


