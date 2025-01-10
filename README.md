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

    ;tsumego lvl            Show tsumego at level lvl

    ;level                  Show available levels. Shortcut lvl by using first character

    ;solve tsumegoID        Show solution to tsumegoID. Shortcut ;s

    ;theme                  Show available themes

    ;theme name             Select your theme

    ;randomtheme            Set theme to random
```

![plot](./gallery/theme.png)

![plot](./gallery/tsumego.png)

![plot](./gallery/tsumego_spoiler.png)

![plot](./gallery/tsumego_solution.png)

## Create a new discord application
- Create a new application at https://discord.com/developers/applications
- Generate a token
- Check "Message Content Intent"
- Check "bot" in OAuth2 URL Generator 
- Click on the generated link

![plot](./gallery/token.png)

![plot](./gallery/message_content_intent.png)

![plot](./gallery/OAuth2%20URL%20Generator.png)

## Setup

- Paste your token into config.json

## Run
`docker build -t tsumego_bot .`

`docker run -d --restart unless-stopped -v </path/to/discord_tsumego_bot>:/data tsumego_bot`

## Have fun
and get stronger!

## License
[MIT](LICENSE)