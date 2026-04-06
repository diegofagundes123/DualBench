/** Indica se o runtime Wails injetou window.go (só existe na janela do app, não no Chrome/Firefox “puro”). */
export function isWailsGoReady() {
  return typeof window.go?.main?.App?.StartDualBenchmark === 'function'
}
