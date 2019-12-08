class AddDomains < ActiveRecord::Migration[5.2]
  def self.up
    create_table :domains, unlogged: true do |t|
      t.string :domain, null: false
    end
  end

  def self.down
    drop_table :domains
  end
end
