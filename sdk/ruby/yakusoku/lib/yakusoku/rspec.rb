# frozen_string_literal: true

require "yakusoku"

module Yakusoku
  module RSpec
    # Helper methods for RSpec integration
    module Helpers
      # Creates a new Pact instance with the given configuration.
      # @param consumer [String] the consumer name
      # @param provider [String] the provider name
      # @param pact_dir [String] the directory to write pact files (default: ./pacts)
      # @return [Pact] a new Pact instance
      def pact(consumer:, provider:, pact_dir: "./pacts")
        Yakusoku::Pact.new(consumer: consumer, provider: provider, pact_dir: pact_dir)
      end
    end

    # DSL for defining contract tests
    module DSL
      def self.included(base)
        base.extend(ClassMethods)
      end

      module ClassMethods
        # Defines a contract test context.
        # @param consumer [String] the consumer name
        # @param provider [String] the provider name
        # @param pact_dir [String] the directory to write pact files
        # @yield the block containing the contract tests
        def contract_test(consumer:, provider:, pact_dir: "./pacts", &block)
          describe "Contract: #{consumer} -> #{provider}" do
            let(:pact) { Yakusoku::Pact.new(consumer: consumer, provider: provider, pact_dir: pact_dir) }

            after { pact.teardown }

            instance_eval(&block)
          end
        end
      end
    end
  end
end

# Auto-configure RSpec if available
if defined?(::RSpec)
  ::RSpec.configure do |config|
    config.include Yakusoku::RSpec::Helpers
    config.extend Yakusoku::RSpec::DSL::ClassMethods
  end
end
