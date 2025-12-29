import { useEffect, useState } from 'react'
import { Link } from 'react-router-dom'
import { fetchContracts, type ContractSummary } from '../api'

export default function ContractsList() {
  const [contracts, setContracts] = useState<ContractSummary[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    fetchContracts()
      .then(setContracts)
      .catch((e) => setError(e.message))
      .finally(() => setLoading(false))
  }, [])

  if (loading) {
    return (
      <div className="flex justify-center items-center h-64">
        <div className="text-gray-500">Loading...</div>
      </div>
    )
  }

  if (error) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4">
        <p className="text-red-700">Error: {error}</p>
      </div>
    )
  }

  if (contracts.length === 0) {
    return (
      <div className="bg-white rounded-lg shadow p-6 text-center">
        <p className="text-gray-500">No contracts found</p>
        <p className="text-sm text-gray-400 mt-2">
          Publish a contract to see it here
        </p>
      </div>
    )
  }

  // Group contracts by provider
  const byProvider = contracts.reduce(
    (acc, c) => {
      if (!acc[c.provider]) acc[c.provider] = []
      acc[c.provider].push(c)
      return acc
    },
    {} as Record<string, ContractSummary[]>
  )

  return (
    <div className="space-y-6">
      <h1 className="text-2xl font-bold text-gray-900">Contracts</h1>
      {Object.entries(byProvider).map(([provider, providerContracts]) => (
        <div key={provider} className="bg-white rounded-lg shadow">
          <div className="border-b border-gray-200 px-4 py-3">
            <h2 className="text-lg font-semibold text-gray-800">
              Provider: {provider}
            </h2>
          </div>
          <ul className="divide-y divide-gray-200">
            {providerContracts.map((c) => (
              <li key={`${c.consumer}-${c.version}`}>
                <Link
                  to={`/contracts/${c.provider}/${c.consumer}/${c.version}`}
                  className="block px-4 py-4 hover:bg-gray-50 transition-colors"
                >
                  <div className="flex items-center justify-between">
                    <div>
                      <span className="font-medium text-gray-900">
                        {c.consumer}
                      </span>
                      <span className="text-gray-500 mx-2">→</span>
                      <span className="text-gray-700">{c.provider}</span>
                    </div>
                    <span className="text-sm text-gray-500 bg-gray-100 px-2 py-1 rounded">
                      v{c.version}
                    </span>
                  </div>
                </Link>
              </li>
            ))}
          </ul>
        </div>
      ))}
    </div>
  )
}
