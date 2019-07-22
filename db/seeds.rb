require 'rubygems/package'
require 'zlib'
require 'active_record'

require_relative '../database'
require_relative '../models'

filename = ARGV[0]

unzipped = Zlib::GzipReader.open(filename)
tar_extract = Gem::Package::TarReader.new(unzipped)
tar_extract.rewind

added = 0

tar_extract.each do |file|
  next if file.directory?
  puts "Adding entries from #{file.header.name}"
  begin
    lines = file.read.split(/[\r\n]+/)
  rescue NoMethodError => err
    puts 'Invalid file', err
    next
  end

  lines.each do |line|

    begin
      # might have ':' in the password, so put it back
      password = line.split(/[:]/)[1..-1].join(":") 
      email = line.split(/[:;]/)[0]
      username, domain = email.split("@")
    rescue NoMethodError => err
      puts 'Invalid line', err
    end

    begin
      record = Record.new
      record.username = Username.find_or_create_by!(name: username)
      record.password = Password.find_or_create_by!(password: password)
      record.domain = Domain.find_or_create_by!(domain: domain)
      record.save!
    rescue ActiveRecord::RecordInvalid => err
      puts "Skipping record for #{username} because:", err
      next
    end
    added += 1
    puts "#{added} unique entries added" if added % 1000 == 0
  end
end
