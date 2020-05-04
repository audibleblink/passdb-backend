class CreateIndexes < ActiveRecord::Migration[5.2]
  def self.up
    add_index :passwords, :password, using: "hash"
    add_index :usernames, :username, using: "hash"
    add_index :domains, :domain, using: "hash"
  end
end
