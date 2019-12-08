require 'sinatra'
require 'json'

require_relative './db/database.rb'
require_relative './models'
require_relative './hibp'

HIBP_CLIENT = HIBP.new(ENV['HIBP_API_KEY'])
DEFAULT_PER_PAGE = 50

before do
  response.headers['Access-Control-Allow-Origin'] = '*'
end

get '/domains/:domain' do
  domain = Domain.find_by(domain: params[:domain])
  paginated(domain, params)
end

get '/usernames/:name' do
  user = Username.find_by(username: params[:name])
  paginated(user, params)
end

get '/passwords/:password' do
  password = Password.find_by(password: params[:password])
  paginated(password, params)
end

get '/emails/:email' do
  user, domain = params[:email].split('@')
  emails = Record.joins(:username)
    .where("usernames.name = ?", user)
    .where("domains.domain = ?", domain)
  prepare(emails)
end

get '/breaches/:email' do
  response = HIBP_CLIENT.breach(params[:email])
  response.map do |br|
    {
      Title: br['Title'],
      Domain: br['Domain'],
      Date: br['BreachDate'],
      Count: br['PwnCount'],
      Description: br['Description'],
      LogoPath: br['LogoPath'],
    }
  end.to_json
end

helpers do
  def prepare(records)
    records
      .map(&:to_hash)
      .sort_by { |h| h[:username] }
      .to_json
  end

  def paginated(model, params)
    limit = params[:per_page] ? params[:per_page].to_i : DEFAULT_PER_PAGE
    records = if model && params[:page]
                page = params[:page].to_i - 1
                offset = limit * page
                model.records.offset(offset).limit(limit)
              elsif model
                model.records.limit(limit)
              else
                []
              end
    prepare(records)
  end
end
