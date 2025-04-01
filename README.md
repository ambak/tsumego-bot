# Discord tsumego-bot

![go](https://img.shields.io/badge/go-1.23-green)
![python](https://img.shields.io/badge/python-3.12-green)
![docker](https://img.shields.io/badge/docker-blue
)

### Tsumego-bot is a Discord bot that will help you improve your tsumego solving skills.

Uses [sgf2image](https://github.com/noword/sgf2image) to generate images.

## Usage
```
Available commands:
    ;tsumego                Show random tsumego. Shortcut ;t

    ;tsumego LVL            Show tsumego at level LVL

    ;level                  Show available levels. Shortcut LVL by using first character

    ;theme                  Show available themes

    ;theme name             Select your theme

    ;randomtheme            Set theme to random

    ;subscribe              Subscribe to get daily tsumego

    ;subscribe LVL          Subscribe to get daily tsumego at level LVL

    ;unsubscribe            Unsubscribe daily tsumego

    ;link OGS_USERNAME      Link your OSG account and your discord account

Example:
    ;t e                    Show elementary tsumego

    ;theme cartoon          Set theme to "cartoon"

    ;subscribe a            Subscribe to daily tsumego at advanced level
    
```

![plot](./gallery/theme.png)

![plot](./gallery/tsumego_spoiler.png)

![plot](./gallery/tsumego_solution.png)

![plot](./gallery/subscribe.png)

## Create a new discord application
- Create a new application at https://discord.com/developers/applications
- Generate a token
- Check "Message Content Intent"
- Check "bot" in OAuth2 URL Generator / Scopes
- Check "Send Messages" and "Manage Roles" in OAuth2 URL Generator / Bot Permissions
- Click on the generated link

![plot](./gallery/token.png)

![plot](./gallery/message_content_intent.png)

![plot](./gallery/OAuth2%20URL%20Generator.png)

## Create a new OGS account
- Create a new account at https://online-go.com/

## Setup

- Paste your token into config.json 
- Paste your tsumego-bot OGS account data into config.json

## Run
`docker build -t tsumego-bot .`

`docker run -d --restart unless-stopped -v </path/to/tsumego-bot>:/data tsumego-bot`

## Have fun
and get stronger!

## License
[MIT](LICENSE)