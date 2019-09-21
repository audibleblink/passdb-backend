class AddPasswords < ActiveRecord::Migration[5.2]
  def self.up
    create_table :passwords, unlogged: true do |t|
      t.string :password, null: false
    end

    add_index :passwords, :password, unique: true
  end

  def self.down
    drop_table :passwords
  end
end
