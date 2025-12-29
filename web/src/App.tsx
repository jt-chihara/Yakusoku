import { Routes, Route } from 'react-router-dom'
import Layout from './components/Layout'
import ContractsList from './pages/ContractsList'
import ContractDetail from './pages/ContractDetail'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Layout />}>
        <Route index element={<ContractsList />} />
        <Route path="contracts/:provider/:consumer/:version" element={<ContractDetail />} />
      </Route>
    </Routes>
  )
}

export default App
