
# CPanel / WHM Exim monitoring

This is the missing piece of software that cPanel have yet to add.
Created to detect malware / virus infected / hacked email accounts that attempt to use a legitimate email and send out spams.
It costed me multiple time, but biggest was USD 90+ on mailgun with 197k of spam emails being billed upon me.
I couldn't wait any longer. Please join me to maintain this :)

## Quick Install (Download Binary)

```bash
# Download latest release
curl -sL https://github.com/u007/eximmon/releases/latest/download/eximmon-linux-amd64.tar.gz | tar xz

# Run installer
chmod +x install.sh && sudo ./install.sh

# Setup config (first time only)
cd /opt/eximmon
API_TOKEN=xxxxx TELEGRAM_BOT_TOKEN=xxx TELEGRAM_ADMIN_IDS=123456 ./eximmon start
# Press Ctrl+C after config saved

# Start service
systemctl start eximmon
```

# How it works?

It scans /var/log/exim_mainlog on interval basis, and logs all email activity being received from authenticated user.
Take note that it only support dovecot_plain at the moment. If you need other type of login, i may add it soon.

# How to operate?

Before starting, please go through this steps to limit the number of emails allowed per hour per domain.

[https://documentation.cpanel.net/display/CKB/How+to+Prevent+Spam+with+Mail+Limiting+Features](https://documentation.cpanel.net/display/CKB/How+to+Prevent+Spam+with+Mail+Limiting+Features)

## Installation (Systemd - Recommended)

### 1. Build for Linux
```bash
make build-linux
```

### 2. Install (as root)
```bash
sudo make install
```

### 3. Setup config (first time only)
```bash
cd /opt/eximmon
API_TOKEN=xxxxx \
TELEGRAM_BOT_TOKEN=xxx \
TELEGRAM_ADMIN_IDS=123456 \
./eximmon start
# Press Ctrl+C after config is saved
```

### 4. Start the service
```bash
systemctl start eximmon
systemctl status eximmon
```

### Useful Commands
```bash
systemctl status eximmon     # Check status
systemctl restart eximmon    # Restart service
systemctl stop eximmon       # Stop service
journalctl -u eximmon -f     # View live logs
./eximmon config             # View current config
./eximmon help               # Show help
```

## Quick Install (rc.local - Alternative)

* login to whm
* development > Manage Api Token
* create an api token, grant "Everything", and copy token (for API_TOKEN)
* add this to /etc/rc.local (on cpanel as root): replace xxxxx, your-email with your configurations

```
cd /root/eximmon &&
API_TOKEN=xxxxx NOTIFY_EMAIL=your-email EXIM_LOG=/var/log/exim_mainlog ./eximmon start > out.log &
```

* to see logs
```
tail -f /root/eximmon/out.log
```

* for helps

```
./eximmon help
```

## Environment variables
* MAX_PER_MIN=4
* MAX_PER_HOUR=100
* NOTIFY_EMAIL=email
* EXIM_LOG=/var/log/exim_mainlog
* WHM_API_HOST=node.servername.com
* PREFER_MODERN_UAPI=true (default: true, set to 'false' to use legacy WHM proxy only)

## API Support

This tool supports both modern UAPI and legacy WHM proxy methods:

| Method | Port | Endpoint | Auth Header |
|--------|------|----------|-------------|
| Modern UAPI (preferred) | 2083 | `/execute/Email/suspend_outgoing` | `cpanel username:TOKEN` |
| Legacy WHM Proxy | 2087 | `/json-api/cpanel?cpanel_jsonapi_...` | `whm root:TOKEN` |

By default, the tool tries modern UAPI first and falls back to legacy WHM proxy if needed. This ensures compatibility with:
- cPanel/WHM v122.x.x and newer (uses modern UAPI)
- Older cPanel/WHM versions (uses legacy fallback)

## Bot Integration

Eximmon supports Telegram and Slack bot integration for real-time notifications and remote control.

### Setup

#### Telegram
1. Create a bot via [@BotFather](https://t.me/botfather) and get the token
2. Get your user ID by messaging [@userinfobot](https://t.me/userinfobot)
3. (Optional) Create a group/channel for notifications and add the bot

#### Slack
1. Create a Slack App with Bot User
2. Install to workspace and get Bot User OAuth Token
3. (Optional) Create a channel for notifications and invite the bot

### Environment Variables

```bash
# Telegram
TELEGRAM_BOT_TOKEN=123456:ABC-DEF
TELEGRAM_ADMIN_IDS=123456789,987654321
TELEGRAM_NOTIFY_CHAT_ID=-100123456789

# Slack
SLACK_BOT_TOKEN=xoxb-123456-abcdef
SLACK_ADMIN_IDS=U12345,U67890
SLACK_NOTIFY_CHANNEL=C12345
```

### Available Commands

| Command | Description | Admin Only |
|---------|-------------|------------|
| `/status` | Check eximmon status | No |
| `/suspend <email>` | Suspend email | Yes |
| `/unsuspend <email>` | Unsuspend email | Yes |
| `/list` | List suspended emails | No |
| `/stats <email|domain>` | View statistics | No |
| `/config` | View configuration | No |
| `/set <key> <value>` | Update threshold | Yes |
| `/whitelist add/remove/list` | Manage whitelist | Yes |

## Available command

* start - continue from last position or start from yesterday, and repeats from last position
* run - continue from last position or start from beginning for one time
* rerun - continue from a specific date
* skip - skip all existing data and repeats for new logs
* reset - reset all data, huh, what?
* suspend - suspend outgoing email
* unsuspend - unsuspend outgoing email
* info - get information of a domain
* test-notify - test send notification mail


# development setup

* copy exim_mainlog to local machine

```
scp root@yourserver.com:/home/user/exim_mainlog ./
```

* clone repository
```
cd $GOPATH/src
git clone git@github.com:u007/eximmon.git

```

* install required dependencies
```
go get
```

* first time run

```
go run main.go start
```

* to build for production

```
GOOS=linux go build -o bin/eximmon main.go

```

# Changes

* 2019 Mar 23 - Fixed dovecot_login, added rerun
* 2018 Sep 25 - First public respository up


# TODO

* count numbers of external relayed recipients per email
* support for other type of dovecot_login methods
* auto delete old data directory by date on every 100 runs
* query email to show list of dates with hour and mins count

# Cleanup

Add this in cronjob, will delete any directory older than 30days.
Remember to change /path-to to your eximmon parent directory of "data"

```
0 1 * * * find /path-to/data/*/ -type f -name '*' -mtime +30 -exec rm -Rf {} \;
```

# Help wanted!

* feel free to report bugs here
* if you need additional features, feel free to add into issue list
* if you need support, contact me @ james@mercstudio.com
