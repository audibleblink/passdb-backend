# PassDB

Password-dump database API server. See accompanying 
[blog post](https://sec.alexflor.es/posts/2020/05/password-dump-database-part-2/) 
for more details.

### Seeding

Dump entries should be in the format:

```
username,domain,password

# where
test@example.com:p4$$w0rd

# becomes
test,example.com,p4$$w0rd
```

Feel free to do this manually, though I had great success using GCP Dataprep

Once in the proper format, you can create the table and import the csv using the GCP Console,
the GCP CLI tool, or the included rake commands.

```bash
bundle exec rake db:create
bundle exec rake db:load[dumps.csv]
```

This will take a while. You may want to manully upload to GCP Storage and copy in the
data from there because if the upload fails with Rake, you'll have to start all over,
and burn through more of your bandwidth.

## Usage

The following enivironment varilables are necessary

```bash
# Project Name
GOOGLE_CLOUD_PROJECT=

# Format: $project.$dataset.$tablename
GOOGLE_BIGQUERY_TABLE=

# Obtained from the GCP Auth Console
GOOGLE_APPLICATION_CREDENTIALS=./credentials.json

# Have I Been Pwned API key
HIBP_API_KEY=
```

Run:

```bash
source .env
go run main.go
```
