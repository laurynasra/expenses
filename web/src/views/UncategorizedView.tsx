import { useState } from 'react'
import ExpenseCard from '../components/ExpenseCard'
import { categorizeExpense } from '../api/client'
import type { Session } from '../types'

interface Props {
  session: Session
  onSessionUpdated: (session: Session) => void
  onBack: () => void
}

export default function UncategorizedView({ session, onSessionUpdated, onBack }: Props) {
  const [skipSet, setSkipSet] = useState<Set<string>>(new Set())
  const [selectedCategory, setSelectedCategory] = useState('')
  const [matcher, setMatcher] = useState('')
  const [loading, setLoading] = useState(false)
  const [error, setError] = useState('')

  const visible = session.uncategorized.filter((e) => !skipSet.has(e.id))
  const current = visible[0]

  async function handleCategorize() {
    if (!current || !selectedCategory) return
    setLoading(true)
    setError('')
    try {
      const updated = await categorizeExpense(session.id, current.id, selectedCategory, matcher || undefined)
      onSessionUpdated(updated)
      setSelectedCategory('')
      setMatcher('')
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to categorize')
    } finally {
      setLoading(false)
    }
  }

  function handleSkip() {
    if (!current) return
    setSkipSet((prev) => new Set(prev).add(current.id))
    setSelectedCategory('')
    setMatcher('')
  }

  if (visible.length === 0) {
    return (
      <div className="min-h-screen p-6 max-w-2xl mx-auto">
        <button onClick={onBack} className="text-gray-500 hover:text-gray-300 text-sm mb-6 transition-colors">
          ← Summary
        </button>
        <div className="text-center mt-20">
          <p className="text-emerald-400 text-2xl mb-2">All done!</p>
          <p className="text-gray-500 text-sm">All expenses have been categorized or skipped.</p>
          <button onClick={onBack} className="mt-6 px-4 py-2 bg-violet-600 hover:bg-violet-500 text-white rounded-lg text-sm transition-colors">
            Back to summary
          </button>
        </div>
      </div>
    )
  }

  return (
    <div className="min-h-screen p-6 max-w-2xl mx-auto">
      <button onClick={onBack} className="text-gray-500 hover:text-gray-300 text-sm mb-6 transition-colors">
        ← Summary
      </button>

      <h1 className="text-2xl font-bold text-violet-400 mb-1">Uncategorized</h1>
      <p className="text-gray-500 text-sm mb-6">
        {visible.length} expense{visible.length !== 1 ? 's' : ''} remaining
      </p>

      <ExpenseCard expense={current} index={0} total={visible.length} />

      <div className="mt-6">
        <p className="text-gray-400 text-xs uppercase tracking-widest mb-3">Select category</p>
        <div className="grid grid-cols-2 gap-2">
          {session.categories.map((cat) => (
            <button
              key={cat.name}
              onClick={() => setSelectedCategory(cat.name)}
              className={`py-2 px-3 rounded border text-sm text-left transition-colors ${
                selectedCategory === cat.name
                  ? 'border-emerald-500 bg-emerald-500/10 text-emerald-300'
                  : 'border-gray-800 text-gray-400 hover:border-gray-600'
              }`}
            >
              {cat.name}
            </button>
          ))}
        </div>
      </div>

      {selectedCategory && (
        <div className="mt-4">
          <label className="block text-gray-400 text-xs uppercase tracking-widest mb-2">
            Matcher keyword <span className="text-gray-600">(optional — saves for auto-categorization)</span>
          </label>
          <input
            type="text"
            value={matcher}
            onChange={(e) => setMatcher(e.target.value)}
            placeholder={`e.g. "${current.description.split(' ')[0].toLowerCase()}"`}
            className="w-full bg-gray-900 border border-gray-700 rounded-lg px-3 py-2 text-gray-200 text-sm focus:outline-none focus:border-violet-500"
          />
        </div>
      )}

      {error && <p className="text-red-400 text-sm mt-3">{error}</p>}

      <div className="flex gap-3 mt-6">
        <button
          onClick={handleCategorize}
          disabled={!selectedCategory || loading}
          className="flex-1 py-2 bg-violet-600 hover:bg-violet-500 disabled:bg-gray-800 disabled:text-gray-600 text-white rounded-lg text-sm font-medium transition-colors"
        >
          {loading ? 'Saving...' : 'Categorize'}
        </button>
        <button
          onClick={handleSkip}
          className="px-4 py-2 border border-gray-700 hover:border-gray-500 text-gray-400 hover:text-gray-300 rounded-lg text-sm transition-colors"
        >
          Skip
        </button>
      </div>
    </div>
  )
}
