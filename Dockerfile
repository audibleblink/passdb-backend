FROM ruby:2.6.3

WORKDIR /app
COPY Gemfile* ./
RUN bundle install

COPY . .
CMD ["bundle", "exec", "ruby", "server.rb"]

# $ docker build -t passdb-server .
# $ docker run --env-file .env passdb-server
