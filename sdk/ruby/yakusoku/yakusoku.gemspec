# frozen_string_literal: true

require_relative "lib/yakusoku/version"

Gem::Specification.new do |spec|
  spec.name = "yakusoku"
  spec.version = Yakusoku::VERSION
  spec.authors = ["yakusoku contributors"]
  spec.summary = "Consumer-driven contract testing for Ruby/Rails"
  spec.description = "A lightweight contract testing SDK compatible with Pact Specification v3"
  spec.homepage = "https://github.com/jt-chihara/yakusoku"
  spec.license = "MIT"
  spec.required_ruby_version = ">= 3.0.0"

  spec.files = Dir["lib/**/*", "LICENSE", "README.md"]
  spec.require_paths = ["lib"]

  spec.add_dependency "webrick", "~> 1.8"

  spec.metadata["homepage_uri"] = spec.homepage
  spec.metadata["source_code_uri"] = spec.homepage
end
