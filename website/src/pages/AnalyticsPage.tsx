import { useParams } from 'react-router-dom'
import { useEffect, useState } from 'react'
import { getAnalytics } from '../api/api'

export interface AnalyticsSummary {
  alias: string
  total_clicks: number
  daily: Record<string, number>
  user_agent: Record<string, number>
}

interface TableRow {
  label: string
  value: number
}

export const AnalyticsPage = () => {
  const { alias } = useParams<{ alias: string }>()
  const [summary, setSummary] = useState<AnalyticsSummary | null>(null)
  const [loading, setLoading] = useState(true)
  const [error, setError] = useState<string | null>(null)

  useEffect(() => {
    if (!alias) return
    setLoading(true)
    getAnalytics(alias)
      .then(setSummary)
      .catch(err => setError(err.message))
      .finally(() => setLoading(false))
  }, [alias])

  if (!alias) return <p className="p-4 text-red-600">Alias not specified</p>
  if (loading) return <p className="p-4">Loading...</p>
  if (error) return <p className="p-4 text-red-600">{error}</p>
  if (!summary) return <p className="p-4">No analytics available</p>

  // Преобразуем объекты в массивы для безопасного map
  const dailyArray: TableRow[] = Object.entries(summary.daily).map(([day, clicks]) => ({
    label: day,
    value: clicks,
  }))
  const uaArray: TableRow[] = Object.entries(summary.user_agent).map(([agent, clicks]) => ({
    label: agent,
    value: clicks,
  }))

  return (
    <div className="max-w-3xl mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Analytics for {alias}</h1>
      <p className="mb-2">Total clicks: {summary.total_clicks}</p>

      <h2 className="text-xl font-semibold mt-4 mb-2">Clicks by Day</h2>
      {dailyArray.length === 0 ? (
        <p>No data available</p>
      ) : (
        <table className="table-auto border-collapse border border-gray-300 w-full mb-4">
          <thead>
            <tr>
              <th className="border border-gray-300 px-2 py-1">Day</th>
              <th className="border border-gray-300 px-2 py-1">Clicks</th>
            </tr>
          </thead>
          <tbody>
            {dailyArray.map(row => (
              <tr key={row.label}>
                <td className="border border-gray-300 px-2 py-1">{row.label}</td>
                <td className="border border-gray-300 px-2 py-1">{row.value}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}

      <h2 className="text-xl font-semibold mt-4 mb-2">Clicks by User Agent</h2>
      {uaArray.length === 0 ? (
        <p>No data available</p>
      ) : (
        <table className="table-auto border-collapse border border-gray-300 w-full">
          <thead>
            <tr>
              <th className="border border-gray-300 px-2 py-1">User Agent</th>
              <th className="border border-gray-300 px-2 py-1">Clicks</th>
            </tr>
          </thead>
          <tbody>
            {uaArray.map(row => (
              <tr key={row.label}>
                <td className="border border-gray-300 px-2 py-1">{row.label}</td>
                <td className="border border-gray-300 px-2 py-1">{row.value}</td>
              </tr>
            ))}
          </tbody>
        </table>
      )}
    </div>
  )
}
