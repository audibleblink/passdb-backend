class AddPasswords < ActiveRecord::Migration
  def self.up
    create_table :passwords do |t|
      t.string :password
      t.timestamps
    end

    add_index :passwords, :password
  end

  def self.down
    drop_table :passwords
  end
end
