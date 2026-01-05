# frozen_string_literal: true

module Yakusoku
  class Interaction
    attr_accessor :description, :provider_state, :request, :response

    def initialize
      @description = nil
      @provider_state = nil
      @request = {}
      @response = {}
    end

    def to_h
      h = {
        description: @description,
        request: @request,
        response: @response
      }
      h[:providerState] = @provider_state if @provider_state
      h
    end
  end

  class Request
    attr_accessor :method, :path, :query, :headers, :body

    def initialize(method:, path:, query: nil, headers: nil, body: nil)
      @method = method
      @path = path
      @query = query
      @headers = headers
      @body = body
    end

    def to_h
      h = { method: @method, path: @path }
      h[:query] = @query if @query
      h[:headers] = @headers if @headers
      h[:body] = @body if @body
      h
    end
  end

  class Response
    attr_accessor :status, :headers, :body

    def initialize(status:, headers: nil, body: nil)
      @status = status
      @headers = headers
      @body = body
    end

    def to_h
      h = { status: @status }
      h[:headers] = @headers if @headers
      h[:body] = @body if @body
      h
    end
  end
end
