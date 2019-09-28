require 'yaml'
require 'active_record'

DEV = 'development'
PROD = 'production'

env = ENV['RACK_ENV'] ||= DEV
conf = ERB.new(File.read('config/database.yml')).result

DB_CONF = YAML::load(conf)[env]

if env == PROD
  DB_CONF.each do |key, val|
    raise "Key #{key} must be set" unless val
  end
end

ActiveRecord::Base.establish_connection(DB_CONF)
