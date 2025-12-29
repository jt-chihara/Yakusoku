import { useEffect, useState } from 'react'
import { useParams, Link } from 'react-router-dom'
import { fetchContract, type Contract } from '../api'

export default function ContractDetail() {
  const { provider, consumer, version } = useParams<{
    provider: string
    consumer: string
    version: string
  }>()
  const [contract, setContract] = useState<Contract | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!provider || !consumer || !version) return
    fetchContract(provider, consumer, version)
      .then(setContract)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [provider, consumer, version])

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500">Loading...</div>
      </div>
    )
  }

  if (error || !contract) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <p className="text-red-700">Error: {error || 'Contract not found'}</p>
        <Link to="/" className="text-blue-600 hover:underline mt-2 inline-block">
          Back to contracts
        </Link>
      </div>
    )
  }

  return (
    <div className="space-y-6">
      <div className="flex items-center gap-4">
        <Link
          to="/"
          className="text-gray-500 hover:text-gray-700 transition-colors"
        >
          ← Back
        </Link>
      </div>

      <div className="bg-white rounded-lg shadow p-6">
        <h1 className="text-2xl font-bold text-gray-900 mb-4">
          {contract.consumer.name} → {contract.provider.name}
        </h1>
        <div className="text-sm text-gray-500">
          Pact Specification v{contract.metadata.pactSpecification.version}
        </div>
      </div>

      <div className="space-y-4">
        <h2 className="text-xl font-semibold text-gray-900">
          Interactions ({contract.interactions.length})
        </h2>
        {contract.interactions.map((interaction, idx) => (
          <div key={idx} className="bg-white rounded-lg shadow overflow-hidden">
            <div className="border-b border-gray-200 px-4 py-3 bg-gray-50">
              <h3 className="font-medium text-gray-900">
                {interaction.description}
              </h3>
              {interaction.providerState && (
                <p className="text-sm text-gray-500 mt-1">
                  Given: {interaction.providerState}
                </p>
              )}
              {interaction.providerStates &&
                interaction.providerStates.length > 0 && (
                  <p className="text-sm text-gray-500 mt-1">
                    Given:{' '}
                    {interaction.providerStates.map((s) => s.name).join(', ')}
                  </p>
                )}
            </div>
            <div className="p-4 grid md:grid-cols-2 gap-4">
              {/* Request */}
              <div>
                <h4 className="font-medium text-gray-700 mb-2">Request</h4>
                <div className="bg-gray-50 rounded p-3 font-mono text-sm">
                  <div className="flex items-center gap-2 mb-2">
                    <span className="px-2 py-1 bg-blue-100 text-blue-800 rounded text-xs font-semibold">
                      {interaction.request.method}
                    </span>
                    <span className="text-gray-800">
                      {interaction.request.path}
                    </span>
                  </div>
                  {interaction.request.headers &&
                    Object.keys(interaction.request.headers).length > 0 && (
                      <div className="mt-2">
                        <div className="text-gray-500 text-xs mb-1">Headers</div>
                        <pre className="text-xs overflow-auto">
                          {JSON.stringify(interaction.request.headers, null, 2)}
                        </pre>
                      </div>
                    )}
                  {interaction.request.body !== undefined && (
                    <div className="mt-2">
                      <div className="text-gray-500 text-xs mb-1">Body</div>
                      <pre className="text-xs overflow-auto whitespace-pre-wrap">
                        {JSON.stringify(interaction.request.body, null, 2) as string}
                      </pre>
                    </div>
                  )}
                </div>
              </div>

              {/* Response */}
              <div>
                <h4 className="font-medium text-gray-700 mb-2">Response</h4>
                <div className="bg-gray-50 rounded p-3 font-mono text-sm">
                  <div className="mb-2">
                    <span
                      className={`px-2 py-1 rounded text-xs font-semibold ${
                        interaction.response.status >= 200 &&
                        interaction.response.status < 300
                          ? 'bg-green-100 text-green-800'
                          : interaction.response.status >= 400
                            ? 'bg-red-100 text-red-800'
                            : 'bg-yellow-100 text-yellow-800'
                      }`}
                    >
                      {interaction.response.status}
                    </span>
                  </div>
                  {interaction.response.headers &&
                    Object.keys(interaction.response.headers).length > 0 && (
                      <div className="mt-2">
                        <div className="text-gray-500 text-xs mb-1">Headers</div>
                        <pre className="text-xs overflow-auto">
                          {JSON.stringify(interaction.response.headers, null, 2)}
                        </pre>
                      </div>
                    )}
                  {interaction.response.body !== undefined && (
                    <div className="mt-2">
                      <div className="text-gray-500 text-xs mb-1">Body</div>
                      <pre className="text-xs overflow-auto whitespace-pre-wrap">
                        {JSON.stringify(interaction.response.body, null, 2) as string}
                      </pre>
                    </div>
                  )}
                </div>
              </div>
            </div>
          </div>
        ))}
      </div>
    </div>
  )
}
