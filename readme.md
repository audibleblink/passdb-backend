# PassDB

Password-dump database API server. See accompanying 
[blog post](https://sec.alexflor.es/posts/2020/05/password-dump-database-part-2/) 
for more details.

See also [accompanying frontend](https://github.com/audibleblink/passdb-frontend)

## API

```
GET /usernames/{username}
GET /domains/{domain}
GET /passwords/{password}
GET /emails/{email}
# response =>  [{"username": "abc", "domain": "example.com", "password": "p4ssw0rd"}, ...]


# Breach info in which the given email was found
GET /breaches/{email}
# response => [{
  "Title": ...,
  "Domain": ...,
  "Date": ...,
  "Count": ...,
  "Description": ...,
  "LogoPath": ...,
},...]
```

## Seeding

Torrents:
```
# Collection #1
magnet:?xt=urn:btih:b39c603c7e18db8262067c5926e7d5ea5d20e12e&dn=Collection+1

# Collections #2 - #5
magnet:?xt=urn:btih:d136b1adde531f38311fbf43fb96fc26df1a34cd&dn=Collection+%232-%235+%26+Antipublic
```

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
the GCP CLI tool, or from the web portal

This will take a while. You may want to manully upload to GCP Storage and copy in the
data from there because if the upload fails with the GCP CLI, you'll have to start all over,
and burn through more of your bandwidth (and credits).

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
go run main.go [port]
```
