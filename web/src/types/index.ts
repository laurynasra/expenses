export interface Expense {
  id: string
  amount: number
  description: string
  date: string
  provider: string
  category: string
  matched: boolean
}

export interface Category {
  name: string
  amount: number
  matchers: string[]
  expenses: Expense[]
}

export interface Session {
  id: string
  provider: string
  categories: Category[]
  uncategorized: Expense[]
  totalAmount: number
}
