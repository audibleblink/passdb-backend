class AddDomains < ActiveRecord::Migration
  def self.up
    create_table :domains do |t|
      t.string :domain, null: false
      t.timestamps null: false
    end
    add_index :domains, :domain, unique: true
  end

  def self.down
    drop_table :domains
  end
end
