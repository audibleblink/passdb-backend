require "yaml"
require "active_record"

task :default => ["query"] 
task :query do
  require_relative 'db/query'
end

namespace :db do

  task :env do
    require_relative 'db/database'
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
    Rake::Task["db:vacuum"].invoke
  end

  desc "Disable vacuum"
  task :vacuum => :env do
    if ENV['RACK_ENV'] == DEV
      puts 'PROD only'
    else
      %w(records passwords domains usernames).each do |table|
        q = "ALTER TABLE #{table} SET (autovacuum_enabled = false)"
        ActiveRecord::Base.connection.exec_query(q)
      end
      puts "Setting applied."
    end
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

  desc "Get table sizes for the database"
  task :stat => :env do
    ActiveSupport::Deprecation.behavior = :silence
    $stderr.reopen(File.new('/dev/null', 'w'))

    if ENV['RACK_ENV'] == DEV
      puts 'PRODUCTION only'
    else
      table = ActiveRecord::Base.connection.exec_query <<-EOF
      SELECT nspname || '.' || relname AS "relation",
          pg_size_pretty(pg_total_relation_size(C.oid)) AS "total_size"
      FROM pg_class C
      LEFT JOIN pg_namespace N ON (N.oid = C.relnamespace)
      WHERE nspname NOT IN ('pg_catalog', 'information_schema')
        AND C.relkind <> 'i'
        AND nspname !~ '^pg_toast'
      ORDER BY pg_total_relation_size(C.oid) DESC
      LIMIT 20;
      EOF
      table.each do |row|
        puts row['total_size'] + "\t\t" + row['relation']
      end
    end

  end

  desc "Get connection states"
  task :conn => :env do
    ActiveSupport::Deprecation.behavior = :silence
    $stderr.reopen(File.new('/dev/null', 'w'))
    if ENV['RACK_ENV'] == DEV
      puts 'PRODUCTION only'
    else
      table = ActiveRecord::Base.connection.exec_query <<-EOF
      SELECT max_conn,used,res_for_super,max_conn-used-res_for_super res_for_normal
      FROM
        (SELECT count(*) used FROM pg_stat_activity) t1,
        (SELECT setting::int res_for_super FROM pg_settings WHERE name=$$superuser_reserved_connections$$) t2,
        (SELECT setting::int max_conn FROM pg_settings WHERE name=$$max_connections$$) t3
      EOF
      table.each do |row|
        row.each_pair do |k, v|
          puts "#{v}\t#{k}"
        end
      end
    end
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

namespace :bench do
  desc "Benchmark Insertion Rate"
  task :insert do
    require 'active_record'
    require_relative 'db/database'
    require_relative 'models'

    sleep_time = 10
    per_min = 60 / sleep_time
    txCQ = "SELECT xact_commit FROM pg_stat_database WHERE datname = 'passdb';"
    txRQ = "SELECT xact_rollback FROM pg_stat_database WHERE datname = 'passdb';"

    orig_commit = ActiveRecord::Base.connection.exec_query(txCQ).rows[0][0].to_i
    puts "Benchmarking seeder progress.  Starting Record Count: #{orig_commit}"
    while true do

      orig_commit = ActiveRecord::Base.connection.exec_query(txCQ).rows[0][0].to_i
      orig_rollbk = ActiveRecord::Base.connection.exec_query(txRQ).rows[0][0].to_i
      sleep sleep_time
      new_commit = ActiveRecord::Base.connection.exec_query(txCQ).rows[0][0].to_i
      new_rollbk = ActiveRecord::Base.connection.exec_query(txRQ).rows[0][0].to_i

      if new_commit == orig_commit
        puts "No new records #{new_commit}"
      else
        comm_delta = new_commit - orig_commit
        rlbk_delta = new_rollbk - orig_rollbk
        puts "TXs: #{new_commit} - Rate: #{comm_delta * per_min}/min - #{comm_delta/sleep_time}/sec - Skipped: #{rlbk_delta}"
      end
    end
  end
end

