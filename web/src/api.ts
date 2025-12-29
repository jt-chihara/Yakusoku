export interface ContractSummary {
  consumer: string
  provider: string
  version: string
}

export interface Interaction {
  description: string
  providerState?: string
  providerStates?: Array<{ name: string; params?: Record<string, unknown> }>
  request: {
    method: string
    path: string
    query?: Record<string, string[]>
    headers?: Record<string, string>
    body?: unknown
  }
  response: {
    status: number
    headers?: Record<string, string>
    body?: unknown
  }
}

export interface Contract {
  consumer: { name: string }
  provider: { name: string }
  interactions: Interaction[]
  metadata: {
    pactSpecification: { version: string }
  }
}

const API_BASE = ''

export async function fetchContracts(): Promise<ContractSummary[]> {
  const res = await fetch(`${API_BASE}/pacts`)
  if (!res.ok) throw new Error('Failed to fetch contracts')
  return res.json()
}

export async function fetchContract(
  provider: string,
  consumer: string,
  version: string
): Promise<Contract> {
  const res = await fetch(
    `${API_BASE}/pacts/provider/${provider}/consumer/${consumer}/version/${version}`
  )
  if (!res.ok) throw new Error('Failed to fetch contract')
  return res.json()
}
