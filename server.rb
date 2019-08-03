require 'sinatra'
require 'json'
require_relative './db/database.rb'
require_relative './models'

PER_PAGE = 50
before do 
  response.headers['Access-Control-Allow-Origin'] = '*'
end

get '/domains/:domain' do
  domain = Domain.find_by(domain: params[:domain])
  paginated(domain, params)
end

get '/usernames/:name' do
  user = Username.find_by(name: params[:name])
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


helpers do

  def prepare(records)
    records
      .map(&:to_hash)
      .sort_by { |h| h[:username] }
      .to_json
  end

  def paginated(model, params)
    records = if model && params[:page]
                page = params[:page].to_i - 1
                offset = PER_PAGE * page
                model.records.offset(offset).limit(PER_PAGE)
              elsif model
                model.records.limit(PER_PAGE)
              else
                []
              end
    p prepare(records)
  end
end
