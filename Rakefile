require "yaml"
require "active_record"

task :default => ["query"] 
task :query do
  require_relative 'query'
end

namespace :db do

  task :env do
    require_relative 'database'
  end

  task :admin => :env do
    db_admin = DB_CONF.merge({'database' => 'postgres', 'schema_search_path' => 'public'})
    ActiveRecord::Base.establish_connection(db_admin)
  end

  desc "Create the database"
  task :create => :admin do
    if ENV['RACK_ENV'] == DEV
      file = File.join(File.dirname(__FILE__), "#{DB_CONF['database']}")
      File.new(file, "w") unless File.exists?(file)
    else
      ActiveRecord::Base.connection.create_database(DB_CONF['database'])
    end
    puts "Database created."
  end

  desc "Migrate the database"
  task :migrate => :env do
    ActiveRecord::Migrator.migrate("db/migrate")

    require 'active_record/schema_dumper'
    filename = "db/schema.rb"
    File.open(filename, "w:utf-8") do |file|
      ActiveRecord::SchemaDumper.dump(ActiveRecord::Base.connection, file)
    end

    puts "Database migrated."
  end

  desc "Drop the database"
  task :drop => :admin do
    if ENV['RACK_ENV'] == DEV
      file = File.join(File.dirname(__FILE__), "#{DB_CONF['database']}")
      File.delete(file) if File.exists?(file)
    else
      ActiveRecord::Base.connection.drop_database(DB_CONF['database'])
    end
    puts "Database deleted."
  end

  desc "Reset the database"
  task :reset do
    Rake::Task["db:drop"].invoke
    Rake::Task["db:create"].invoke
    ActiveRecord::Base.establish_connection(DB_CONF)
    Rake::Task["db:migrate"].invoke
  end
end

namespace :g do
  desc "Generate migration"
  task :migration do
    name = ARGV[1] || raise("Specify name: rake g:migration your_migration")
    timestamp = Time.now.strftime("%Y%m%d%H%M%S")
    path = File.expand_path("../db/migrate/#{timestamp}_#{name}.rb", __FILE__)
    migration_class = name.split("_").map(&:capitalize).join

    File.open(path, 'w') do |file|
      file.write <<-EOF
class #{migration_class} < ActiveRecord::Migration
  def self.up
  end
  def self.down
  end
end
      EOF
    end

    puts "Migration #{path} created"
    abort # needed stop other tasks
  end
end
