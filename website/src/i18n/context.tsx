import { useState, useEffect, type ReactNode } from 'react'
import en from './en'
import zhTW from './zh-TW'
import { LanguageContext, type Locale } from './languageContext'

const dictionaries: Record<Locale, Record<string, unknown>> = {
  en: en as unknown as Record<string, unknown>,
  'zh-TW': zhTW as unknown as Record<string, unknown>,
}

function getNestedValue(obj: Record<string, unknown>, path: string): string {
  const keys = path.split('.')
  let current: unknown = obj
  for (const key of keys) {
    if (current === null || current === undefined || typeof current !== 'object') {
      return path
    }
    current = (current as Record<string, unknown>)[key]
  }
  return typeof current === 'string' ? current : path
}

const STORAGE_KEY = 'openshannon-locale'

export function LanguageProvider({ children }: { children: ReactNode }) {
  const [locale, setLocaleState] = useState<Locale>(() => {
    const stored = localStorage.getItem(STORAGE_KEY)
    if (stored === 'en' || stored === 'zh-TW') return stored
    return 'en'
  })

  useEffect(() => {
    localStorage.setItem(STORAGE_KEY, locale)
  }, [locale])

  const setLocale = (newLocale: Locale) => {
    setLocaleState(newLocale)
  }

  const t = (key: string): string => {
    return getNestedValue(dictionaries[locale], key)
  }

  return (
    <LanguageContext.Provider value={{ locale, setLocale, t }}>
      {children}
    </LanguageContext.Provider>
  )
}
