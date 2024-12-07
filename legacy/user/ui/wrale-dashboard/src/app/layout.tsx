import type { Metadata } from 'next'
import './globals.css'
import { Inter } from 'next/font/google'
import { Navbar } from '@/components/navigation/Navbar'
import { ErrorBoundary } from '@/components/error/ErrorBoundary'
import { LoadingProvider } from '@/components/ui/LoadingProvider'

const inter = Inter({ subsets: ['latin'] })

export const metadata: Metadata = {
  title: 'Wrale Fleet Dashboard',
  description: 'Physical-first Raspberry Pi fleet management system',
}

export default function RootLayout({
  children,
}: {
  children: React.ReactNode
}) {
  return (
    <html lang="en">
      <body className={inter.className}>
        <ErrorBoundary>
          <LoadingProvider>
            <div className="min-h-screen bg-gray-100 flex">
              <Navbar />
              <div className="flex-1 ml-64">
                {children}
              </div>
            </div>
          </LoadingProvider>
        </ErrorBoundary>
      </body>
    </html>
  )
}