require 'rubygems/package'
require 'zlib'
require 'active_record'
require 'pushover'


require_relative '../database'
require_relative '../models'

filename = ARGV[0]
puts filename


unzipped = Zlib::GzipReader.open(filename)
tar_extract = Gem::Package::TarReader.new(unzipped)
tar_extract.rewind

added = 0
dupes = 0

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

      if email
        email = email == "" ? "email" : email.downcase
      end
      if domain
        domain = domain == "" ? "domain" : domain.downcase
      end
      if username
        username = username == "" ? "username" : username.downcase
      end

    rescue NoMethodError => err
      nil
    end

    begin
      record = Record.new
      record.username = Username.find_or_create_by!(name: username)
      record.domain = Domain.find_or_create_by!(domain: domain)
      record.password = Password.find_or_create_by!(password: password)
      record.save!
    rescue ActiveRecord::RecordNotUnique => err
      #puts "Skipping record for #{username} because:", err
      dupes += 1
      next
    rescue ArgumentError => err
      next
    rescue PG::NotNullViolation => err
      next
    rescue ActiveRecord::StatementInvalid => err
      next
    end
    added += 1
    puts "#{added} unique entries added" if added % 10000 == 0
  end

  dup_msg = "#{dupes} duplicate entries omitted"
  total = Record.count
  msg = "Finished: #{file.header.name}.\n#{total} added so far\n#{dup_msg}"
end
