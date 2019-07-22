class AddDomains < ActiveRecord::Migration
  def self.up
    create_table :domains do |t|
      t.string :domain
      t.timestamps
    end
    add_index :domains, :domain
  end

  def self.down
    drop_table :domains
  end
end
