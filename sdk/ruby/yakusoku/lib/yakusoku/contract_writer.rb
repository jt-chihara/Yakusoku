# frozen_string_literal: true

require "json"
require "fileutils"

module Yakusoku
  class ContractWriter
    def initialize(consumer:, provider:, interactions:)
      @consumer = consumer
      @provider = provider
      @interactions = interactions
    end

    def write(dir)
      FileUtils.mkdir_p(dir)

      filename = "#{normalize_name(@consumer)}-#{normalize_name(@provider)}.json"
      filepath = File.join(dir, filename)

      File.write(filepath, JSON.pretty_generate(to_contract))
      filepath
    end

    private

    def normalize_name(name)
      name.downcase.gsub(/[^a-z0-9]+/, "_").gsub(/^_|_$/, "")
    end

    def to_contract
      {
        consumer: { name: @consumer },
        provider: { name: @provider },
        interactions: @interactions.map { |i| interaction_to_h(i) },
        metadata: {
          pactSpecification: { version: "3.0.0" },
          client: {
            name: "yakusoku-ruby",
            version: VERSION
          }
        }
      }
    end

    def interaction_to_h(interaction)
      h = {
        description: interaction.description,
        request: interaction.request,
        response: interaction.response
      }
      h[:providerState] = interaction.provider_state if interaction.provider_state
      h
    end
  end
end
