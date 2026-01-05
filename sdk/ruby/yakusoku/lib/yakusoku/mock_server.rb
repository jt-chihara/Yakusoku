# frozen_string_literal: true

require "webrick"
require "json"

module Yakusoku
  class MockServer
    attr_reader :port

    def initialize(interactions)
      @interactions = interactions
      @matched_interactions = []
      @server = nil
      @thread = nil
      @port = nil
    end

    def start
      @port = find_available_port
      @server = WEBrick::HTTPServer.new(
        Port: @port,
        Logger: WEBrick::Log.new("/dev/null"),
        AccessLog: []
      )

      @server.mount_proc "/" do |req, res|
        handle_request(req, res)
      end

      @thread = Thread.new { @server.start }

      # Wait for server to be ready
      sleep 0.1 until server_ready?
    end

    def stop
      @server&.shutdown
      @thread&.join(1)
    end

    def url
      "http://localhost:#{@port}"
    end

    def unmatched_interactions
      @interactions - @matched_interactions
    end

    private

    def find_available_port
      server = TCPServer.new("127.0.0.1", 0)
      port = server.addr[1]
      server.close
      port
    end

    def server_ready?
      TCPSocket.new("127.0.0.1", @port).close
      true
    rescue Errno::ECONNREFUSED
      false
    end

    def handle_request(req, res)
      interaction = find_matching_interaction(req)

      if interaction
        @matched_interactions << interaction
        response = interaction.response

        res.status = response[:status]
        (response[:headers] || {}).each do |key, value|
          res[key] = value
        end

        if response[:body]
          res["Content-Type"] ||= "application/json"
          res.body = response[:body].is_a?(String) ? response[:body] : JSON.generate(response[:body])
        end
      else
        res.status = 500
        res["Content-Type"] = "application/json"
        res.body = JSON.generate({
          error: "No matching interaction found",
          request: {
            method: req.request_method,
            path: req.path,
            query: req.query
          }
        })
      end
    end

    def find_matching_interaction(req)
      @interactions.find do |interaction|
        request = interaction.request
        matches_method?(request, req) &&
          matches_path?(request, req) &&
          matches_query?(request, req) &&
          matches_headers?(request, req) &&
          matches_body?(request, req)
      end
    end

    def matches_method?(expected, actual)
      expected[:method].to_s.upcase == actual.request_method.upcase
    end

    def matches_path?(expected, actual)
      expected[:path] == actual.path
    end

    def matches_query?(expected, actual)
      return true unless expected[:query]

      expected[:query].all? do |key, value|
        actual.query[key.to_s] == value
      end
    end

    def matches_headers?(expected, actual)
      return true unless expected[:headers]

      expected[:headers].all? do |key, value|
        actual[key.to_s]&.downcase == value.to_s.downcase
      end
    end

    def matches_body?(expected, actual)
      return true unless expected[:body]

      actual_body = actual.body
      return false if actual_body.nil? || actual_body.empty?

      expected_body = expected[:body]
      actual_parsed = JSON.parse(actual_body, symbolize_names: true)

      if expected_body.is_a?(Hash)
        hash_matches?(expected_body, actual_parsed)
      else
        expected_body == actual_parsed
      end
    rescue JSON::ParserError
      false
    end

    def hash_matches?(expected, actual)
      return false unless actual.is_a?(Hash)

      expected.all? do |key, value|
        actual_value = actual[key.to_sym] || actual[key.to_s]
        if value.is_a?(Hash)
          hash_matches?(value, actual_value)
        else
          value == actual_value
        end
      end
    end
  end
end
