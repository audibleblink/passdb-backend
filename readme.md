# PassDB

Password-dump database API server. See accompanying blog post for details
on setting up GCP Dataprep, Storage, and BigQuery

### Seeding

Dump entries should be in the format:

```
username,domain,password

# where
test@example.com:p4$$w0rd

# becomes
test,example.com,p4$$w0rd
```

Feel free to do this manually, though I had great success and enjoyment using GCP Dataprep

Once in the proper format, you can create the table and import the csv using the GCP Console,
the, GCP CLI tool, or the included rake commands.

```bash
bundle exec rake db:create
bundle exec rake db:load[dumps.csv]
```

This will take a while. You may want to manully upload to GCP storage and copy in the
data from there because should the upload fail with Rake, you'll have to start all over,
and burn through more of your bandwidth.

## Usage

The following enivironment varilables are necessary

```bash
# Project Name
export GOOGLE_CLOUD_PROJECT=

# Format: $project.$dataset.$tablename
export GOOGLE_BIGQUERY_TABLE=

# Obtained from GCP Console
export GOOGLE_CLOUD_KEYFILE_JSON=./credentials.json

# Have I Been Pwned API key
export HIBP_API_KEY=xxxxxxxxxxxxxxxxxxxxxxxxxxxxxxxx
```

Install the project deps with:

```bash
gem install bundle
bundle install
```

`server.rb` starts a JSON API for use with the passdb-frontend. 

```bash
bundle exec ruby server.rb 
```

If you don't have a ruby environment set, using docker will be less of a headache.

```
# build the image
docker build -t passdb-server .

# run the container, passing the necessary environment variables, port maps, and volume mounts
docker run --env-file .env -p 4567:4567 -v $FOLDER_CONTAINING_DUMPS:/app passdb-server bash
```

### Stats

Run `bundle exec rake -T` to see all tasks. 

You can pull table sizes and counts using:

```bash
$ bundle exec rake db:stats
```
