# passdb

Password dump database normalizer and seeder

super alpha

## DB Setup

Depending on the data drive, add one of the `conf` files from the `db` directory to 
Postgres' `conf.d` dir.

```
cp db/16gb_4cpu_ssd.conf /etc/postgres/10/main/conf.d/dump.conf
systemctl restart postgres@10-main.service
```

Currently averaging around 350K inserts/minute with these settings and table configuration in `db/migrate/`

## Usage

### Seeding

There's a test tar in the `tests` dir

Dump entries should be in the format:

```
email@domain.com:password
```

The parse logic in `seeder` takes a best-effort approach to pulling `domain`, `username`, and
`password` from each line. Some dumps use `;` and `seeder` looks for that too. Apart from that, if
it can't find all three datapoints in the line, it isn't added to the database.

```
# for psql
export RACK_ENV=production

# for sqlite
export RACK_ENV=development

bundle install
bundle exec rake db:reset

#build golang seeder
cd seeder && go build -o seeder main.go

# pushover.net token for mobile progress alerts
export PO_USR=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
export PO_API=yyyyyyyyyyyyyyyyyyyyyyyyyyyyyy
export PG_CONN='postgres://passdb_user:passdb_pass@localhost/passdb'

#macos postgres want this string instead
export PG_CONN='postgres://passdb_user:passdb_pass@localhost/passdb?ssl_mode=disabled'

# seed the
./seeder test_data.tar.gz
```

### Querying

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

## Stats

Run `rake -T` to see all tasks. 

At the time of writing you can pull table sizes, current connection pool utilization 

Stats below taken at 8 million entries:
![](https://i.imgur.com/4ej5HlH.png)


Seeder benchmarks with `bundle exec rake bench:insert`
![](https://i.imgur.com/HGqhUJf.png)
