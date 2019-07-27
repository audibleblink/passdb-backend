# encoding: UTF-8
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

ActiveRecord::Schema.define(version: 20190726170600) do

  # These are extensions that must be enabled in order to support this database
  enable_extension "plpgsql"

  create_table "domains", id: :bigserial, force: :cascade do |t|
    t.string "domain", null: false
  end

  add_index "domains", ["domain"], name: "index_domains_on_domain", unique: true, using: :btree

  create_table "passwords", id: :bigserial, force: :cascade do |t|
    t.string "password", null: false
  end

  add_index "passwords", ["password"], name: "index_passwords_on_password", unique: true, using: :btree

  create_table "records", id: :bigserial, force: :cascade do |t|
    t.integer "password_id", null: false
    t.integer "domain_id",   null: false
    t.integer "username_id", null: false
  end

  add_index "records", ["password_id", "domain_id", "username_id"], name: "index_records_on_password_id_and_domain_id_and_username_id", unique: true, using: :btree

  create_table "usernames", id: :bigserial, force: :cascade do |t|
    t.string "name", null: false
  end

  add_index "usernames", ["name"], name: "index_usernames_on_name", unique: true, using: :btree

end
