/**
 * Bindings Go; regenere com `wails generate module` se assinaturas mudarem.
 */
export function StartDualBenchmark(arg1, arg2) {
  const fn = window.go?.main?.App?.StartDualBenchmark
  if (typeof fn !== 'function') {
    return Promise.reject(
      new Error(
        'Bridge Wails indisponível (window.go). Abra o app pela janela do DualBench, não pelo navegador na URL do Vite.',
      ),
    )
  }
  return fn(arg1, arg2)
}
