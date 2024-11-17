# Backup MYSQL/MariaDB File

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
}
```

### 2. To Setup File Path and Key for Encryption
Create `.env` file first, you can copy `.env.example` and then setup File Path in `.env` file
```
LOG_FOLDER=__Your log folder__
OUTPUT_FOLDER=__Your backup folder__
CONFIG_FILE=config.json
KEY=__Key In Text Form___
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