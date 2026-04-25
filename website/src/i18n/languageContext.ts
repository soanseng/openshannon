import { createContext } from 'react'

export type Locale = 'en' | 'zh-TW'

export interface LanguageContextType {
  locale: Locale
  setLocale: (locale: Locale) => void
  t: (key: string) => string
}

export const LanguageContext = createContext<LanguageContextType | null>(null)
