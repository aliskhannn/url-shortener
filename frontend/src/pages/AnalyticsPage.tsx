import { useParams } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { getAnalytics } from '../api/api'
import { AnalyticsTable } from '../components/AnalyticsTable'
import type { Analytics } from '../entities/analytics'

export const AnalyticsPage = () => {
  const { alias } = useParams<{ alias: string }>()
  const [data, setData] = useState<Analytics[]>([])
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!alias) return
    getAnalytics(alias)
      .then(setData)
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [alias])

  if (!alias) return <p className="p-4 text-red-600">Alias not specified</p>
  if (loading) return <p className="p-4">Loading...</p>
  if (error) return <p className="p-4 text-red-600">{error}</p>

  return (
    <div className="max-w-3xl mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Analytics for {alias}</h1>
      <AnalyticsTable data={data} />
    </div>
  )
}