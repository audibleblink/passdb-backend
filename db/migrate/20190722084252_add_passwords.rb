class AddPasswords < ActiveRecord::Migration
  def self.up
    create_table :passwords do |t|
      t.string :password, null: false
      t.timestamps null: false
    end

    add_index :passwords, :password, unique: true
  end

  def self.down
    drop_table :passwords
  end
end
