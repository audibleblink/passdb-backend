class AddRecords < ActiveRecord::Migration
  def self.up
    create_table :records do |t|
      t.references :password
      t.references :domain
      t.references :username

      t.timestamps
    end

    
  end

  def self.down
    drop_table :records
  end
end

