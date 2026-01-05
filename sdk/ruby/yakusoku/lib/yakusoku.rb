# frozen_string_literal: true

require_relative "yakusoku/version"
require_relative "yakusoku/interaction"
require_relative "yakusoku/mock_server"
require_relative "yakusoku/contract_writer"
require_relative "yakusoku/pact"

module Yakusoku
  class Error < StandardError; end
  class VerificationError < Error; end
end
