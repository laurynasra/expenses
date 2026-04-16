interface Props {
  pct: number
}

export default function ProgressBar({ pct }: Props) {
  return (
    <div className="flex-1 bg-gray-800 rounded-full h-2 overflow-hidden">
      <div
        className="h-full bg-violet-600 rounded-full transition-all duration-300"
        style={{ width: `${Math.min(pct, 100)}%` }}
      />
    </div>
  )
}
