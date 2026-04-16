import { useState } from 'react'
import UploadView from './views/UploadView'
import SummaryView from './views/SummaryView'
import CategoryDetailView from './views/CategoryDetailView'
import UncategorizedView from './views/UncategorizedView'
import type { Session } from './types'

type AppState =
  | { view: 'upload' }
  | { view: 'summary'; session: Session }
  | { view: 'category'; session: Session; categoryName: string }
  | { view: 'uncategorized'; session: Session }

export default function App() {
  const [state, setState] = useState<AppState>({ view: 'upload' })

  function setSession(session: Session) {
    setState((prev) => {
      if (prev.view === 'summary') return { view: 'summary', session }
      if (prev.view === 'category') return { view: 'category', session, categoryName: prev.categoryName }
      if (prev.view === 'uncategorized') return { view: 'uncategorized', session }
      return prev
    })
  }

  switch (state.view) {
    case 'upload':
      return (
        <UploadView
          onUploaded={(session) => setState({ view: 'summary', session })}
        />
      )

    case 'summary':
      return (
        <SummaryView
          session={state.session}
          onSelectCategory={(name) => setState({ view: 'category', session: state.session, categoryName: name })}
          onViewUncategorized={() => setState({ view: 'uncategorized', session: state.session })}
          onBack={() => setState({ view: 'upload' })}
        />
      )

    case 'category':
      return (
        <CategoryDetailView
          session={state.session}
          categoryName={state.categoryName}
          onBack={() => setState({ view: 'summary', session: state.session })}
        />
      )

    case 'uncategorized':
      return (
        <UncategorizedView
          session={state.session}
          onSessionUpdated={(session) => {
            setSession(session)
            if (session.uncategorized.length === 0) {
              setState({ view: 'summary', session })
            }
          }}
          onBack={() => setState({ view: 'summary', session: state.session })}
        />
      )
  }
}
