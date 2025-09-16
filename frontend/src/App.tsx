import { BrowserRouter as Router, Routes, Route } from 'react-router-dom'
import { Home } from './pages/Home'
import { AnalyticsPage } from './pages/AnalyticsPage'

export const App = () => (
  <Router>
    <Routes>
      <Route path="/" element={<Home />} />
      <Route path="/analytics/:alias" element={<AnalyticsPage />} />
    </Routes>
  </Router>
)