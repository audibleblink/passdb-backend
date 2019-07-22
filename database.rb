require 'yaml'
require 'active_record'

DEV = 'development'
PROD = 'production'

env = ENV['RACK_ENV'] ||= DEV
DB_CONF = YAML::load(File.open('config/database.yml'))[env]
ActiveRecord::Base.establish_connection(DB_CONF)
