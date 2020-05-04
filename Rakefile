require "./models"

desc 'Environment check'
task :env do
    _, dataset, table = Record.table.split('.')
    unless dataset && table
      raise("Is your environment set?")
    end
    @dataset = Record.bq.dataset(dataset)
    @table   = @dataset.table(table)
end

namespace :db do

  desc "create the database"
  task :create => :env do
    @dataset.create_table @table do |t|
      t.string "username"
      t.string "domain"
      t.string "password"
    end
  end

  # TODO enable uploads from GCP storage
  desc "Load the database"
  task :load, [:filename] => [:env] do |t, args|
    puts "[+] Really load file #{args[:filename]}? [y|n]"
    response = $stdin.gets.chomp

    exit unless response.downcase == 'y'
    puts "Uploading #{args[:filename]}"
    file = File.open(filename)

    initial_count = @table.rows_count
    @table.load(file)

    insert_count = @table.rows_count = initial_count
    puts "Inserted #{insert_count} records"
  end

  desc "Drop the database"
  task :drop => :env do
    # TODO
  end

  # TODO might be too much for local
  # would be cool to export to GCP storage
  desc "Export the database"
  task :export, [:filename] => [:env] do |t, args|
    data = @table.data
    File.open(args[:filename], 'w+') do |f|
      data.each { |r| f.puts r.values.join(',') }
    end
  end

  desc "Get table sizes for the database"
  task :stats => :env do
    query = "
      SELECT 
        COUNT(DISTINCT username) AS u_count,
        COUNT(DISTINCT domain) AS d_count,
        COUNT(DISTINCT password) AS p_count
      FROM #{Record.table}
    "

    uniques = Record.bq.query(query)[0]

    # require 'pry'; binding.pry
    puts "Stats for #{@dataset.gapi.id}:"
    puts "========================================"
    puts "Bytes:   #{@table.bytes_count}"
    puts "Rows:    #{@table.rows_count}"
    puts "Unique:"
    puts "  Usernames: #{uniques[:u_count]}"
    puts "  Domain:    #{uniques[:d_count]}"
    puts "  Password:  #{uniques[:p_count]}"
  end
end
