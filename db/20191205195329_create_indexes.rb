class CreateIndexes < ActiveRecord::Migration
  def self.up
    add_index :records, [:password_id, :domain_id, :username_id], unique: true
    add_index :records, [:username_id, :domain_id]
    add_index :records, [:username_id, :password_id]

    add_index :passwords, :password, unique: true
    add_index :usernames, :username, unique: true
    add_index :domains, :domain, unique: true
  end
end
