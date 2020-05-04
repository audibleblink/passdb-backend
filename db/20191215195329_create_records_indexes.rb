class CreateRecordsIndexes < ActiveRecord::Migration[5.2]
  def self.up
    add_index :records, [:username_id, :domain_id]
    add_index :records, [:username_id, :password_id]
  end
end
