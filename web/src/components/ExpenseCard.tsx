import type { Expense } from '../types'

interface Props {
  expense: Expense
  index?: number
  total?: number
}

export default function ExpenseCard({ expense, index, total }: Props) {
  return (
    <div className="border border-violet-600/40 rounded-lg p-4 bg-gray-900/50">
      {index !== undefined && total !== undefined && (
        <p className="text-gray-500 text-xs mb-2">
          Expense {index + 1} of {total}
        </p>
      )}
      <p className="text-gray-200 font-medium truncate">{expense.description}</p>
      <div className="flex justify-between items-center mt-2">
        <span className="text-amber-400 font-bold">€{Math.abs(expense.amount).toFixed(2)}</span>
        <span className="text-gray-500 text-xs uppercase tracking-wide">{expense.provider}</span>
      </div>
    </div>
  )
}
