import type { Session, Category } from '../types'

const BASE = '/api'

async function fetchJSON<T>(url: string, init?: RequestInit): Promise<T> {
  const res = await fetch(url, init)
  const data = await res.json()
  if (!res.ok) throw new Error(data.error ?? `HTTP ${res.status}`)
  return data as T
}

export async function uploadFile(provider: string, file: File): Promise<Session> {
  const form = new FormData()
  form.append('provider', provider)
  form.append('file', file)
  return fetchJSON<Session>(`${BASE}/upload`, { method: 'POST', body: form })
}

export async function getSession(id: string): Promise<Session> {
  return fetchJSON<Session>(`${BASE}/sessions/${id}`)
}

export async function getCategories(): Promise<Category[]> {
  return fetchJSON<Category[]>(`${BASE}/categories`)
}

export async function addMatcher(categoryName: string, matcher: string): Promise<Category> {
  return fetchJSON<Category>(`${BASE}/categories/${encodeURIComponent(categoryName)}/matchers`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ matcher }),
  })
}

export async function categorizeExpense(
  sessionId: string,
  expenseId: string,
  categoryName: string,
  matcher?: string,
): Promise<Session> {
  return fetchJSON<Session>(`${BASE}/expenses/${sessionId}/categorize`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify({ expenseId, categoryName, matcher: matcher ?? '' }),
  })
}
