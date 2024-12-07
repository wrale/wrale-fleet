'use client'

import React, { createContext, useContext, useState } from 'react'

interface LoadingContextType {
  isLoading: boolean
  setIsLoading: (loading: boolean) => void
  loadingText?: string
  setLoadingText: (text?: string) => void
}

const LoadingContext = createContext<LoadingContextType | undefined>(undefined)

export function useLoading() {
  const context = useContext(LoadingContext)
  if (!context) {
    throw new Error('useLoading must be used within a LoadingProvider')
  }
  return context
}

interface LoadingProviderProps {
  children: React.ReactNode
}

export function LoadingProvider({ children }: LoadingProviderProps) {
  const [isLoading, setIsLoading] = useState(false)
  const [loadingText, setLoadingText] = useState<string>()

  return (
    <LoadingContext.Provider value={{ isLoading, setIsLoading, loadingText, setLoadingText }}>
      {children}
      {isLoading && (
        <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
          <div className="bg-white rounded-lg p-6 max-w-sm mx-4">
            <div className="flex items-center justify-center mb-4">
              <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-wrale-primary"></div>
            </div>
            {loadingText && (
              <p className="text-center text-gray-600">{loadingText}</p>
            )}
          </div>
        </div>
      )}
    </LoadingContext.Provider>
  )
}