require "google/cloud/bigquery"

class Record 
  attr_reader :username, :domain, :password

  @@bq = Google::Cloud::Bigquery.new 
  @@table = ENV['GOOGLE_BIGQUERY_TABLE']

  def initialize(username:, domain:, password:)
    @username = username
    @domain   = domain
    @password = password
  end

  def self.find_by(hash={})
    column, value = hash.first
    query = "SELECT DISTINCT * FROM #{table} WHERE #{column} = @#{column}"
    query_and_build(query, column => value)
  end

  def self.find_by_email(username, domain)
    query = "SELECT DISTINCT * FROM #{table} WHERE username = @username AND domain = @domain"
    query_and_build(query, username: username, domain: domain)
  end

  def to_raw
    "#{username}@#{domain}:#{password}"
  end

  def to_hash
    { username: username, domain: domain, password: password }
  end

  def self.query_and_build(query, params)
    data = @@bq.query(query, params: params)
    data.map { |d| new(d) }
  end

  def self.table
    @@table
  end
end

