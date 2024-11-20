# Backup MYSQL/MariaDB File

## Update
- 21/11/2024
    - [New] project structure
    - [Add] discord notification
    - [Add] backup config : specify tables and ignore tables

#### How to use
### 1. To Configuration
You can config database connection in `config.json` file
```
{
    "name": "configuration name",
    "user": "__user__",
    "password": "__password__",
    "host": "__host__",
    "port": 3306,
    "database": "__database name__"

    "ignored_tables":["db_name.tbl_name1","db_name.tbl_name2"], // ignore dumping some tables
    or
    "tables":["tbl_name3"] // dump specify tables
}
```

### 2. To Setup File Path and Key for Encryption
Create `.env` file first, you can copy `.env.example` and then setup File Path in `.env` file
```
LOG_FOLDER=__Your log folder__
OUTPUT_FOLDER=__Your backup folder__
CONFIG_FILE=config.json
KEY=__Key In Text Form___
DISCORD_WEBHOOK=https://discord.com/api/webhooks/{webhooks.id}/{webhooks.token}
DISCORD_BOT_NAME=__DISCORD_BOT_NAME__
DISCORD_BOT_AVATAR=__DISCORD_BOT_AVATAR (image url)__
```

### 4. To Build
You can use `go build` for build execution file on your Operation System
```
go build
```

### 5. To Backup
```
./db-backup -backup
```

### 5. To Decryption
```
./db-backup -decrypt -file [file path with .enc]
```