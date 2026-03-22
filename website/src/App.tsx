import { Routes, Route } from 'react-router-dom'
import Landing from './pages/Landing'
import DocsLayout from './layouts/DocsLayout'
import GettingStarted from './pages/docs/GettingStarted'
import Commands from './pages/docs/Commands'
import Configuration from './pages/docs/Configuration'
import GoogleServices from './pages/docs/GoogleServices'
import ImageGeneration from './pages/docs/ImageGeneration'
import Security from './pages/docs/Security'

function App() {
  return (
    <Routes>
      <Route path="/" element={<Landing />} />
      <Route path="/docs" element={<DocsLayout />}>
        <Route index element={<GettingStarted />} />
        <Route path="getting-started" element={<GettingStarted />} />
        <Route path="commands" element={<Commands />} />
        <Route path="configuration" element={<Configuration />} />
        <Route path="google-services" element={<GoogleServices />} />
        <Route path="image-generation" element={<ImageGeneration />} />
        <Route path="security" element={<Security />} />
      </Route>
    </Routes>
  )
}

export default App
