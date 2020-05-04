class CreateIndexes < ActiveRecord::Migration[5.2]
  def self.up
    add_index :passwords, :password, unique: true
    add_index :usernames, :username, unique: true
    add_index :domains, :domain, unique: true
  end
end
