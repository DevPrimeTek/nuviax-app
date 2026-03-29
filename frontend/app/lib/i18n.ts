// i18n.ts — NuviaX custom translation hook
// No external library — plain object lookup with type safety.
// Language detection order: localStorage('nv_lang') → 'ro' (default)
// Fallback: if key missing → returns the key itself (never crashes)

import { useState, useEffect } from 'react'

export type Lang = 'ro' | 'en' | 'ru'

// Import all locale files statically
import ro from '../public/locales/ro.json'
import en from '../public/locales/en.json'
import ru from '../public/locales/ru.json'

const locales: Record<Lang, Record<string, string>> = { ro, en, ru }

export type TranslationKey = keyof typeof ro

export function useTranslation() {
  const [lang, setLang] = useState<Lang>('ro')

  useEffect(() => {
    const stored = (localStorage.getItem('nv_lang') as Lang) || 'ro'
    setLang(stored)

    // Listen for language changes from settings page
    function onStorage(e: StorageEvent) {
      if (e.key === 'nv_lang' && e.newValue) {
        setLang(e.newValue as Lang)
      }
    }
    window.addEventListener('storage', onStorage)
    return () => window.removeEventListener('storage', onStorage)
  }, [])

  function t(key: TranslationKey | string): string {
    return locales[lang]?.[key] ?? locales['ro']?.[key] ?? key
  }

  return { t, lang }
}
