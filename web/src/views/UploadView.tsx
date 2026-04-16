import { useState, useRef } from 'react'
import { uploadFile } from '../api/client'
import type { Session } from '../types'

interface Props {
  onUploaded: (session: Session) => void
}

const PROVIDERS = ['wise', 'revolut', 'seb']

export default function UploadView({ onUploaded }: Props) {
  const [provider, setProvider] = useState('wise')
  const [file, setFile] = useState<File | null>(null)
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')
  const inputRef = useRef<HTMLInputElement>(null)

  async function handleSubmit(e: React.FormEvent) {
    e.preventDefault()
    if (!file) return
    setLoading(true)
    setError('')
    try {
      const session = await uploadFile(provider, file)
      onUploaded(session)
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Upload failed')
    } finally {
      setLoading(false)
    }
  }

  return (
    <div className="min-h-screen flex items-center justify-center p-4">
      <div className="w-full max-w-md">
        <h1 className="text-3xl font-bold text-violet-400 mb-2 tracking-tight">Expense Tracker</h1>
        <p className="text-gray-500 mb-8 text-sm">Upload a bank statement to get started</p>

        <form onSubmit={handleSubmit} className="space-y-5">
          <div>
            <label className="block text-gray-400 text-xs uppercase tracking-widest mb-2">
              Provider
            </label>
            <div className="flex gap-2">
              {PROVIDERS.map((p) => (
                <button
                  key={p}
                  type="button"
                  onClick={() => setProvider(p)}
                  className={`flex-1 py-2 px-3 rounded border text-sm uppercase tracking-wide transition-colors ${
                    provider === p
                      ? 'border-violet-500 bg-violet-600/20 text-violet-300'
                      : 'border-gray-700 text-gray-500 hover:border-gray-500'
                  }`}
                >
                  {p}
                </button>
              ))}
            </div>
          </div>

          <div>
            <label className="block text-gray-400 text-xs uppercase tracking-widest mb-2">
              CSV File
            </label>
            <div
              onClick={() => inputRef.current?.click()}
              className={`border-2 border-dashed rounded-lg p-6 text-center cursor-pointer transition-colors ${
                file
                  ? 'border-emerald-500/60 bg-emerald-500/5'
                  : 'border-gray-700 hover:border-gray-500'
              }`}
            >
              <input
                ref={inputRef}
                type="file"
                accept=".csv"
                className="hidden"
                onChange={(e) => setFile(e.target.files?.[0] ?? null)}
              />
              {file ? (
                <p className="text-emerald-400 text-sm">{file.name}</p>
              ) : (
                <p className="text-gray-600 text-sm">Click to select a CSV file</p>
              )}
            </div>
          </div>

          {error && <p className="text-red-400 text-sm">{error}</p>}

          <button
            type="submit"
            disabled={!file || loading}
            className="w-full py-3 bg-violet-600 hover:bg-violet-500 disabled:bg-gray-800 disabled:text-gray-600 text-white rounded-lg font-medium transition-colors"
          >
            {loading ? 'Parsing...' : 'Parse Expenses'}
          </button>
        </form>
      </div>
    </div>
  )
}
