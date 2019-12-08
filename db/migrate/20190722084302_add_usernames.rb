class AddUsernames < ActiveRecord::Migration[5.2]
  def self.up
    create_table :usernames, unlogged: true do |t|
      t.string :username, null: false
    end
  end

  def self.down
    drop_table :usernames
  end
end

