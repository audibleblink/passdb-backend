class AddRecords < ActiveRecord::Migration[5.2]
  def self.up
    create_table :records, unlogged: true do |t|
      t.references :password, null: false
      t.references :domain, null: false
      t.references :username, null: false
    end
  end

  def self.down
    drop_table :records
  end
end

