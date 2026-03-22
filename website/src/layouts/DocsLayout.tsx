import { Outlet } from 'react-router-dom'
import Navbar from '../components/Navbar'
import Footer from '../components/Footer'
import DocsSidebar from '../components/DocsSidebar'
import MobileDocNav from '../components/MobileDocNav'

export default function DocsLayout() {
  return (
    <div className="min-h-screen flex flex-col">
      <Navbar />
      <MobileDocNav />
      <div className="flex-1 max-w-6xl mx-auto w-full px-6 flex">
        <DocsSidebar />
        <main className="flex-1 py-8 lg:pl-8 min-w-0">
          <div className="docs-content max-w-3xl">
            <Outlet />
          </div>
        </main>
      </div>
      <Footer />
    </div>
  )
}
