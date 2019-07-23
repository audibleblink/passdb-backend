# passdb

password dump database normalizer and seeder

super alpha


## Usage


### Seeding

Theres a test tar int `tests` dir

```
bundle install
bundle exec rake db:reset
bundle exec ruby db/seeds.rb <dump.tar.gz>
# contents of tar
# some-folder
# |-- dumpfile1.txt
# |-- dumpfile2.txt
# --- dumpfile3.txt
```

### Querying

Associations are set in the ORM such that pivotting on any of `username`, `password`, or `domain`
is possible

```
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

