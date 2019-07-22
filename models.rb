require 'active_record'

class Password < ActiveRecord::Base
  has_many :records
  has_many :usernames, through: :records
  has_many :domains, through: :records

  validates :password, uniqueness: true, presence: true
end

class Domain < ActiveRecord::Base
  has_many :records
  has_many :passwords, through: :records
  has_many :usernames, through: :records

  validates :domain, uniqueness: true, presence: true
end

class Username < ActiveRecord::Base
  has_many :records
  has_many :passwords, through: :records
  has_many :domains, through: :records

  validates :name, uniqueness: true, presence: true
end

class Record < ActiveRecord::Base
  belongs_to :domain
  belongs_to :password
  belongs_to :username

  validates :username, uniqueness: { scope: [:domain, :password] }

  # needed?
  # validates :password, uniqueness: { scope: [:domain, :username] }
  # validates :domain, uniqueness: { scope: [:username, :password] }
end

