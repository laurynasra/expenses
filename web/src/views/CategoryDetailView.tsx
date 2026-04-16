import type { Session } from '../types'

interface Props {
  session: Session
  categoryName: string
  onBack: () => void
}

export default function CategoryDetailView({ session, categoryName, onBack }: Props) {
  const cat = session.categories.find((c) => c.name === categoryName)
  if (!cat) return null

  return (
    <div className="min-h-screen p-6 max-w-2xl mx-auto">
      <button onClick={onBack} className="text-gray-500 hover:text-gray-300 text-sm mb-6 transition-colors">
        ← Summary
      </button>

      <h1 className="text-2xl font-bold text-emerald-400 mb-1">{cat.name}</h1>
      <p className="text-amber-400 text-xl font-bold mb-6">€{Math.abs(cat.amount).toFixed(2)}</p>

      <div className="space-y-1">
        {cat.expenses.map((exp) => (
          <div
            key={exp.id}
            className="flex items-center justify-between py-2 px-3 rounded border border-gray-800/60 hover:border-gray-700"
          >
            <span className="text-gray-300 text-sm truncate flex-1 mr-4">{exp.description}</span>
            <span className="text-amber-400 text-sm font-mono shrink-0">
              €{Math.abs(exp.amount).toFixed(2)}
            </span>
          </div>
        ))}
      </div>

      {cat.expenses.length === 0 && (
        <p className="text-gray-600 text-sm">No expenses in this category.</p>
      )}
    </div>
  )
}
