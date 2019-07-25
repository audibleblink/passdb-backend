require 'active_record'
require 'pry-byebug'

require_relative 'database'
require_relative '../models'

ActiveRecord::Base.logger = Logger.new(STDOUT)
binding.pry
puts 'Begin Query Interface'
