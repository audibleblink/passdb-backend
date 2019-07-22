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

```
bundle exec rake


[1] pry(main)> Password.first
D, [2019-07-22T19:11:56.615253 #10090] DEBUG -- :   Password Load (0.1ms)  SELECT  "passwords".* FROM "passwords"  ORDER BY "passwords"."id" ASC LIMIT 1
=> #<Password:0x00005590259a2580
 id: 1,
 password: "p@ssword1!",
 created_at: 2019-07-22 23:11:47 UTC,
 updated_at: 2019-07-22 23:11:47 UTC>
[2] pry(main)>
```

