# passdb

Password-dump database normalizer, seeder, and API server.

## DB Setup

Depending on the data drive, add one of the `conf` files from the `db` directory to 
Postgres' `conf.d` dir.

```
cp db/16gb_4cpu_ssd.conf /etc/postgres/10/main/conf.d/dump.conf
systemctl restart postgres@10-main.service
```

Currently averaging around 350K inserts/minute with these settings.

## Usage

The following enivironment varilables are necessary

```
# Required for rake and API
PG_HOST=localhost
DB_NAME=passdb
DB_USER=passdb_user
DB_PASS=passdb_pass

# Required for API server
HIBP_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx

# Required for seeder
PG_CONN=postgres://${DB_USER}:${DB_PASS}@${PG_HOST}/${DB_NAME}
PO_USR=ii5299VZk7DCXcntrnkxXMalEb2Hph # pushover.net for mobile notifications
PO_API=abtnpi5eht2qckz67s6j5xfpoe83tg # pushover.net for mobile notifications
```

### Seeding

There's a test tar in the `tests` dir

Dump entries should be in the format:

```
email@domain.com:password
```

The parsing logic in `seeder` takes a best-effort approach to pulling `domain`, `username`, and
`password` from each line. Some dumps use `;` instead of `:` as a field seperator. The `seeder`
will account for that. Apart from that, if it can't find all three datapoints in the line, it isn't
added to the database.

```
# for psql
export RACK_ENV=production

# setup the db
bundle install
bundle exec rake db:reset

# build golang seeder
cd seeder && go build main.go

# macos' postgres wants this string instead
export PG_CONN='postgres://passdb_user:passdb_pass@localhost/passdb?ssl_mode=disabled'

# seed the database
./seeder test_data.tar.gz
```

### Manual Querying

Associations are set in the ORM such that pivotting on any of `username`, `password`, or `domain`
is possible

```
# to start the query interface
bundle exec rake


# start with a domain
yahoo = Domain.find_by(domain: "yahoo.com")

# find all passwords by yahoo mail users
yahoo.passwords

# find all yahoo mail users
yahoo.usernames

# find all password of a particular yahoo mail user
yahoo.usernames.first.passwords



# start with a user
eric = Usernames.find_by(name: "eric1990")

# see all passwords belonging to eric
eric.passwords

# see all email account for eric
eric.domains



# starting with a password
pass = Password.find_by(password: "P@ssw0rd!")

# see the users that share this password
pass.usernames
```

### The API
`server.rb` starts a JSON API for use with the passdb-frontend. If you have a ruby environment set
up, simply `bundle exec ruby server.rb`. If not, using docker will be less of a headache.

```
# build the image
docker build -t passdb-server .

# run the container, passing the necessary environment variables
docker run --env-file .env passdb-server
```

## Stats

Run `rake -T` to see all tasks. 

You can pull table sizes, current connection pool utilization 

Stats below taken at 8 million entries:
![](https://i.imgur.com/4ej5HlH.png)


Seeder benchmarks with `bundle exec rake bench:insert`
![](https://i.imgur.com/HGqhUJf.png)
