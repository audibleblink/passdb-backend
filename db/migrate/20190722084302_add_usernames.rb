class AddUsernames < ActiveRecord::Migration
  def self.up
    create_table :usernames do |t|
      t.string :name
      t.timestamps
    end

    add_index :usernames, :name
  end

  def self.down
    drop_table :usernames
  end
end

