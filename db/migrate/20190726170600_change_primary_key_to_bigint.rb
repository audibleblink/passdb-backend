class ChangePrimaryKeyToBigint < ActiveRecord::Migration
def self.up
  change_column :usernames, :id, :bigint
  change_column :domains, :id, :bigint
  change_column :passwords, :id, :bigint

  change_column :records, :id, :bigint
  change_column :records, :username_id, :bigint
  change_column :records, :password_id, :bigint
  change_column :records, :domain_id, :bigint
end

def self.down
end

end
