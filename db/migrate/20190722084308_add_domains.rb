class AddDomains < ActiveRecord::Migration[5.2]
  def self.up
    create_table :domains, unlogged: true do |t|
      t.string :domain, null: false
    end
    add_index :domains, :domain, unique: true
  end

  def self.down
    drop_table :domains
  end
end
