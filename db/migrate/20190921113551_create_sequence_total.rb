class CreateSequenceTotal < ActiveRecord::Migration[5.2]
  def up
    execute <<-SQL
      CREATE SEQUENCE total;
    SQL
  end

  def down
    execute <<-SQL
      DROP SEQUENCE total;
    SQL
  end
end


