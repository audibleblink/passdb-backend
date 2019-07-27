class ChangePrimaryKeyToBigint < ActiveRecord::Migration
def self.up
  change_column :usernames, :id, :bigint
  change_column :domains, :id, :bigint
  change_column :passwords, :id, :bigint
  change_column :records, :id, :bigint
end

def self.down
end

end
