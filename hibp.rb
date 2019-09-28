raise "HIBP_API_KEY not set." unless ENV['HIBP_API_KEY']
require 'httparty'

class HIBP
  include HTTParty
  base_uri 'https://haveibeenpwned.com/api/v3'

  def initialize(token)
    @options = {
      headers: { "hibp-api-key" => token, "user-agent" => 'script' }
    }
  end

  def breach(account)
    opts = @options.merge( query: { truncateResponse: false } )
    self.class.get("/breachedaccount/#{account}", opts)
  end
end
