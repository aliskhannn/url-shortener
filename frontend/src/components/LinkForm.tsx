import { useState } from 'react'
import { createLink } from '../api/api'

export const LinkForm = () => {
  const [url, setUrl] = useState('')
  const [alias, setAlias] = useState('')
  const [result, setResult] = useState<string | null>(null)
  const [error, setError] = useState<string | null>(null)

  const handleSubmit = async (e: React.FormEvent) => {
    e.preventDefault()
    setError(null)
    setResult(null)

    try {
      const data = await createLink({ url, alias: alias || undefined })
      setResult(`${window.location.origin}/s/${data.alias}`)
    } catch (err: any) {
      setError(err.message)
    }
  }

  return (
    <div className="max-w-md mx-auto p-4">
      <h1 className="text-2xl font-bold mb-4">Shorten URL</h1>
      <form onSubmit={handleSubmit} className="flex flex-col gap-2">
        <input
          type="url"
          placeholder="Enter URL"
          value={url}
          onChange={e => setUrl(e.target.value)}
          required
          className="border p-2 rounded"
        />
        <input
          type="text"
          placeholder="Custom alias (optional)"
          value={alias}
          onChange={e => setAlias(e.target.value)}
          className="border p-2 rounded"
        />
        <button type="submit" className="bg-blue-500 text-white p-2 rounded hover:bg-blue-600">
          Shorten
        </button>
      </form>

      {result && (
        <div className="mt-4 text-green-600">
          Short URL: <a href={result} target="_blank" className="underline">{result}</a>
        </div>
      )}
      {error && <div className="mt-4 text-red-600">{error}</div>}
    </div>
  )
}
