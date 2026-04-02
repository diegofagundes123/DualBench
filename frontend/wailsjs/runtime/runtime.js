/**
 * Bridge mínimo; em `wails dev` / build oficial o Wails pode substituir por versão gerada.
 */
export function EventsOn(eventName, callback) {
  if (!window.runtime || typeof window.runtime.EventsOn !== 'function') {
    console.warn('runtime.EventsOn indisponível (modo browser puro)')
    return () => {}
  }
  return window.runtime.EventsOn(eventName, callback)
}

export function EventsOff(eventName, ...additionalEventNames) {
  if (!window.runtime || typeof window.runtime.EventsOff !== 'function') {
    return
  }
  return window.runtime.EventsOff(eventName, ...additionalEventNames)
}
