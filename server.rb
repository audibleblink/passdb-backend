require 'sinatra'
require 'json'
require_relative './db/database.rb'
require_relative './models'

def prepare(records)
  records.map do |r|
    {username: r.username.name, password: r.password. password, domain: r.domain.domain}
  end
    .sort_by { |h| h[:username] }
    .to_json
end

def paginated(model, params)
  records = if model && params[:page]
              model.records.page(params[:page]).per(50).without_count
            elsif model
              model.records.page(1).per(50).without_count
            else
              []
            end
  sleep 0.3
  p prepare(records)
end

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
