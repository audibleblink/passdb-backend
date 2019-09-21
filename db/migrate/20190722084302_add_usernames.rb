class AddUsernames < ActiveRecord::Migration[5.2]
  def self.up
    create_table :usernames, unlogged: true do |t|
      t.string :name, null: false
    end

    add_index :usernames, :name, unique: true
  end

  def self.down
    drop_table :usernames
  end
end

