import {
    createI18n
} from 'vue-i18n'
import enUS from './en-US'
import zhCN from './zh-CN'

const messages = {
    'en-US': enUS,
    'zh-CN': zhCN
}

export const LOCAL_LANGUAGE_KEY = 'cacao_local_language_key'

const getBrowserLanguage = () => {
    const locale = localStorage.getItem(LOCAL_LANGUAGE_KEY)
    if (Object.keys(messages).includes(locale)) {
        return locale
    }
    const language = navigator.language || navigator.userLanguage
    return messages[language] ? language : 'en-US'
}

const i18n = createI18n({
    locale: getBrowserLanguage(),
    fallbackLocale: 'en-US',
    messages
})

export default i18n
