import ProgressBar from '../components/ProgressBar'
import type { Session } from '../types'

interface Props {
  session: Session
  onSelectCategory: (name: string) => void
  onViewUncategorized: () => void
  onBack: () => void
}

export default function SummaryView({ session, onSelectCategory, onViewUncategorized, onBack }: Props) {
  const total = session.totalAmount

  return (
    <div className="min-h-screen p-6 max-w-2xl mx-auto">
      <button onClick={onBack} className="text-gray-500 hover:text-gray-300 text-sm mb-6 transition-colors">
        ← New upload
      </button>

      <h1 className="text-2xl font-bold text-violet-400 mb-1">Expense Summary</h1>
      <p className="text-amber-400 text-xl font-bold mb-6">€{Math.abs(total).toFixed(2)} total</p>

      <div className="space-y-2">
        {session.categories
          .filter((c) => c.amount !== 0)
          .map((cat) => {
            const pct = total !== 0 ? (Math.abs(cat.amount) / Math.abs(total)) * 100 : 0
            return (
              <button
                key={cat.name}
                onClick={() => onSelectCategory(cat.name)}
                className="w-full text-left p-3 rounded-lg border border-gray-800 hover:border-violet-600/50 hover:bg-violet-600/5 transition-all group"
              >
                <div className="flex items-center gap-3 mb-2">
                  <span className="text-emerald-400 font-semibold w-36 truncate group-hover:text-emerald-300">
                    {cat.name}
                  </span>
                  <span className="text-amber-400 text-sm font-mono w-24 text-right">
                    €{Math.abs(cat.amount).toFixed(2)}
                  </span>
                  <span className="text-gray-600 text-xs w-10 text-right">{pct.toFixed(1)}%</span>
                </div>
                <ProgressBar pct={pct} />
              </button>
            )
          })}

        {session.uncategorized.length > 0 && (
          <button
            onClick={onViewUncategorized}
            className="w-full text-left p-3 rounded-lg border border-amber-500/30 hover:border-amber-500/60 hover:bg-amber-500/5 transition-all"
          >
            <div className="flex items-center gap-2">
              <span className="text-amber-400">⚠</span>
              <span className="text-amber-300 font-semibold">Uncategorized</span>
              <span className="text-gray-500 text-sm ml-auto">{session.uncategorized.length} expenses</span>
            </div>
          </button>
        )}
      </div>
    </div>
  )
}
