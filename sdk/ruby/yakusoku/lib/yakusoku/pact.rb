# frozen_string_literal: true

module Yakusoku
  class Pact
    attr_reader :consumer, :provider, :pact_dir, :interactions

    def initialize(consumer:, provider:, pact_dir: "./pacts")
      @consumer = consumer
      @provider = provider
      @pact_dir = pact_dir
      @interactions = []
      @current_interaction = nil
      @mock_server = nil
    end

    # Sets the provider state for the current interaction.
    # @param state [String] the provider state description
    # @return [Pact] self for method chaining
    def given(state)
      @current_interaction ||= Interaction.new
      @current_interaction.provider_state = state
      self
    end

    # Sets the interaction description.
    # @param description [String] a description of the interaction
    # @return [Pact] self for method chaining
    def upon_receiving(description)
      @current_interaction ||= Interaction.new
      @current_interaction.description = description
      self
    end

    # Sets the expected request.
    # @param request [Hash, Request] the expected request
    # @return [Pact] self for method chaining
    def with_request(request)
      @current_interaction ||= Interaction.new
      @current_interaction.request = normalize_request(request)
      self
    end

    # Sets the expected response and finalizes the interaction.
    # @param response [Hash, Response] the expected response
    # @return [Pact] self for method chaining
    def will_respond_with(response)
      @current_interaction ||= Interaction.new
      @current_interaction.response = normalize_response(response)

      @interactions << @current_interaction
      @current_interaction = nil
      self
    end

    # Returns the mock server URL.
    # @return [String] the base URL of the mock server
    def server_url
      @mock_server&.url
    end

    # Verifies the interactions by running the given block.
    # Starts the mock server, yields to the block, and writes the contract file.
    # @yield the block containing the consumer code to test
    # @return [void]
    def verify(&block)
      @mock_server = MockServer.new(@interactions)
      @mock_server.start

      begin
        block.call(@mock_server.url)

        unmatched = @mock_server.unmatched_interactions
        unless unmatched.empty?
          raise VerificationError, "Unmatched interactions: #{unmatched.map(&:description).join(', ')}"
        end

        write_contract
      ensure
        @mock_server.stop
      end
    end

    # Stops the mock server and cleans up resources.
    # @return [void]
    def teardown
      @mock_server&.stop
    end

    private

    def normalize_request(request)
      case request
      when Hash
        request
      when Request
        request.to_h
      else
        raise ArgumentError, "request must be a Hash or Yakusoku::Request"
      end
    end

    def normalize_response(response)
      case response
      when Hash
        response
      when Response
        response.to_h
      else
        raise ArgumentError, "response must be a Hash or Yakusoku::Response"
      end
    end

    def write_contract
      writer = ContractWriter.new(
        consumer: @consumer,
        provider: @provider,
        interactions: @interactions
      )
      writer.write(@pact_dir)
    end
  end
end
