
# Aggre*Gator*

Gator is a cli RSS aggregator. it uses PostGresSQL and Go to aggregate RSS feeds and store them for future browsing via CLI.

# Installation and Usage:

### requirements setup:

This CLI program needs PostgresSQL and Go installed.

steps to installed it in Linux(Debian):

install go version 1.25+ using webi
```
curl -sS https://webi.sh/golang | sh
```
run `go version` to make sure the installation worked.

install PostgreSQL using apt
```
sudo apt update
sudo apt install postgresql postgresql-contrib
```

run `psql --version` to make sure the installation worked.

update postgres password (optional for fresh installs only)

```
sudo passwd postgres
```



### Program setup:

install aggreGator using go

```
go install github.com/o0n1x/aggreGator@latest
```

#### create a database for gator:
enter psql shell
```
sudo -u postgres psql
```
Create new gator database
```
CREATE DATABASE gator;
```
connect to the new database
```
\c gator
```
set the user password for the gator database
```
ALTER USER postgres PASSWORD 'postgres';
```

### config file setup:

we need to setup a config file to communicate with PostgreSQL.


create config file in the home directory

```
touch ~/.gatorconfig.json
```
open the config file and paste the following

```
{
  "db_url": "postgres://postgres:username@localhost:5432/gator?sslmode=disable", 
  "current_user_name": "username"
}
```
change 'username' to the name that you chose for postgresSQL (by default its postgres)

### Usage:

gator is a CLI tool and its usage is as follows

```
gator (command) [args]
```

### list of commands:

| command   | args               | usage                                                                             |
|-----------|--------------------|-----------------------------------------------------------------------------------|
| login     | name               | login with given username                                                         |
| register  | name               | register a username                                                               |
| users     |                    | list usernames                                                                    |
| agg       | time_between_reqs* | start aggregate loop that aggregates every time t (timee_between_reqs)            |
| addfeed   | name , url         | add a new rss feed with given name and url. logged user auto follows the new feed |
| feeds     |                    | get rss feed for logged user                                                      |
| follow    | url                | follow an existing rss feed with a given url                                      |
| following |                    | list followed feeds of logged user                                                |
| unfollow  | url                | unfollow a feed for logged user                                                   |
| browse    | limit (default: 2) | list the latest n Posts from followed feeds                                       |

*=time_between_reqs ex: 1s , 1m, 1h , etc..