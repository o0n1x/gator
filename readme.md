
# Aggre**Gator**

aggregator is a cli RSS aggregator. it uses PostGresSQL and Go to aggregate RSS feeds and store them for future browsing via CLI.

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



### Usage:

commands that can be used

