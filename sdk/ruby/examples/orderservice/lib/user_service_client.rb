# frozen_string_literal: true

require "net/http"
require "json"
require "uri"

# UserServiceClient is a client for the UserService API.
# This is an example of a real client that would be used in your application.
class UserServiceClient
  class NotFoundError < StandardError; end
  class ApiError < StandardError; end

  attr_reader :base_url

  def initialize(base_url:)
    @base_url = base_url
  end

  # Fetches a user by ID.
  # @param id [Integer] the user ID
  # @return [Hash] the user data
  # @raise [NotFoundError] if the user is not found
  # @raise [ApiError] if the API returns an error
  def get_user(id)
    uri = URI("#{@base_url}/users/#{id}")
    response = Net::HTTP.get_response(uri)

    case response.code.to_i
    when 200
      JSON.parse(response.body, symbolize_names: true)
    when 404
      raise NotFoundError, "User #{id} not found"
    else
      raise ApiError, "API error: #{response.code}"
    end
  end

  # Creates a new user.
  # @param name [String] the user's name
  # @param email [String] the user's email
  # @return [Hash] the created user data
  def create_user(name:, email:)
    uri = URI("#{@base_url}/users")
    http = Net::HTTP.new(uri.host, uri.port)

    request = Net::HTTP::Post.new(uri)
    request["Content-Type"] = "application/json"
    request.body = JSON.generate(name: name, email: email)

    response = http.request(request)

    case response.code.to_i
    when 201
      JSON.parse(response.body, symbolize_names: true)
    else
      raise ApiError, "API error: #{response.code}"
    end
  end

  # Fetches a user's orders.
  # @param user_id [Integer] the user ID
  # @return [Array<Hash>] the user's orders
  def get_user_orders(user_id)
    uri = URI("#{@base_url}/users/#{user_id}/orders")
    response = Net::HTTP.get_response(uri)

    case response.code.to_i
    when 200
      JSON.parse(response.body, symbolize_names: true)
    when 404
      raise NotFoundError, "User #{user_id} not found"
    else
      raise ApiError, "API error: #{response.code}"
    end
  end
end
