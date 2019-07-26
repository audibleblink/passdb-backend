class AddRecords < ActiveRecord::Migration
  def self.up
    create_table :records, unlogged: true do |t|
      t.references :password, null: false
      t.references :domain, null: false
      t.references :username, null: false

      t.timestamps null: false
    end

    add_index :records, [:password_id, :domain_id, :username_id], unique: true
  end

  def self.down
    drop_table :records
  end
end

