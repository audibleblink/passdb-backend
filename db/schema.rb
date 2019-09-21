# This file is auto-generated from the current state of the database. Instead
# of editing this file, please use the migrations feature of Active Record to
# incrementally modify your database, and then regenerate this schema definition.
#
# Note that this schema.rb definition is the authoritative source for your
# database schema. If you need to create the application database on another
# system, you should be using db:schema:load, not running all the migrations
# from scratch. The latter is a flawed and unsustainable approach (the more migrations
# you'll amass, the slower it'll run and the greater likelihood for issues).
#
# It's strongly recommended that you check this file into your version control system.

ActiveRecord::Schema.define(version: 2019_09_21_113551) do

  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "domains", force: :cascade do |t|
    t.string "domain", null: false
    t.index ["domain"], name: "index_domains_on_domain", unique: true
  end

  create_table "passwords", force: :cascade do |t|
    t.string "password", null: false
    t.index ["password"], name: "index_passwords_on_password", unique: true
  end

  create_table "records", force: :cascade do |t|
    t.bigint "password_id", null: false
    t.bigint "domain_id", null: false
    t.bigint "username_id", null: false
    t.index ["domain_id"], name: "index_records_on_domain_id"
    t.index ["password_id", "domain_id", "username_id"], name: "index_records_on_password_id_and_domain_id_and_username_id", unique: true
    t.index ["password_id"], name: "index_records_on_password_id"
    t.index ["username_id", "domain_id"], name: "index_records_on_username_id_and_domain_id"
    t.index ["username_id", "password_id"], name: "index_records_on_username_id_and_password_id"
    t.index ["username_id"], name: "index_records_on_username_id"
  end

  create_table "usernames", force: :cascade do |t|
    t.string "name", null: false
    t.index ["name"], name: "index_usernames_on_name", unique: true
  end

end
