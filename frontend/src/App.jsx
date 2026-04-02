import { useCallback, useEffect, useState } from 'react'
import { StartDualBenchmark } from '../wailsjs/go/main/App.js'
import { EventsOff, EventsOn } from '../wailsjs/runtime/runtime.js'
import './App.css'

const EVT = 'benchmark:progress'

const emptySlot = () => ({
  phase: 'idle',
  pct: 0,
  speedMBps: 0,
})

export default function App() {
  const [driveA, setDriveA] = useState('E:')
  const [driveB, setDriveB] = useState('F:')
  const [loading, setLoading] = useState(false)
  const [uiError, setUiError] = useState('')
  const [progress, setProgress] = useState({
    overallPct: 0,
    drive1: emptySlot(),
    drive2: emptySlot(),
  })
  const [results, setResults] = useState(null)

  useEffect(() => {
    const off = EventsOn(EVT, (payload) => {
      if (!payload) return
      setProgress({
        overallPct: payload.overallPct ?? 0,
        drive1: {
          phase: payload.drive1?.phase ?? 'idle',
          pct: payload.drive1?.pct ?? 0,
          speedMBps: payload.drive1?.speedMBps ?? 0,
        },
        drive2: {
          phase: payload.drive2?.phase ?? 'idle',
          pct: payload.drive2?.pct ?? 0,
          speedMBps: payload.drive2?.speedMBps ?? 0,
        },
      })
    })
    return () => {
      if (typeof off === 'function') off()
      EventsOff(EVT)
    }
  }, [])

  const onStart = useCallback(async () => {
    setUiError('')
    setResults(null)
    setLoading(true)
    setProgress({
      overallPct: 0,
      drive1: emptySlot(),
      drive2: emptySlot(),
    })

    try {
      const summary = await StartDualBenchmark(driveA.trim(), driveB.trim())
      setResults(summary)
    } catch (e) {
      setUiError(e?.message || String(e))
    } finally {
      setLoading(false)
    }
  }, [driveA, driveB])

  const disabled = loading

  return (
    <div className="app">
      <h1>DualBench</h1>
      <p className="sub">
        Benchmark paralelo de leitura e escrita em dois volumes, com I/O sem cache do sistema
        (medição mais próxima da velocidade real do pendrive). ~128&nbsp;MB por drive — aguarde o término.
      </p>

      {uiError ? <div className="banner-err">{uiError}</div> : null}

      <div className="grid">
        <div className="field">
          <label htmlFor="d1">Caminho drive 1</label>
          <input
            id="d1"
            value={driveA}
            onChange={(ev) => setDriveA(ev.target.value)}
            disabled={disabled}
            placeholder="E: ou E:\"
            autoComplete="off"
          />
        </div>
        <div className="field">
          <label htmlFor="d2">Caminho drive 2</label>
          <input
            id="d2"
            value={driveB}
            onChange={(ev) => setDriveB(ev.target.value)}
            disabled={disabled}
            placeholder="F: ou F:\"
            autoComplete="off"
          />
        </div>
      </div>

      <div className="actions">
        <button type="button" className="btn" onClick={onStart} disabled={disabled}>
          {loading ? (
            <>
              <span className="spinner" aria-hidden />
              Executando benchmark…
            </>
          ) : (
            'Iniciar benchmark'
          )}
        </button>
      </div>

      {(loading || progress.drive1.phase !== 'idle' || progress.drive2.phase !== 'idle') && (
        <div className="panel">
          <h2>Progresso ao vivo</h2>
          <p className="meta">
            <span>
              Geral: <strong>{progress.overallPct.toFixed(1)}%</strong>
            </span>
          </p>
          <DriveBar label="Drive 1" slot={progress.drive1} />
          <DriveBar label="Drive 2" slot={progress.drive2} />
        </div>
      )}

      {results && (
        <div className="panel summary">
          <h2>Resultado</h2>
          <div className="summary-grid">
            <ResultCard title="Drive 1" path={results.drive1?.path} data={results.drive1} />
            <ResultCard title="Drive 2" path={results.drive2?.path} data={results.drive2} />
          </div>
        </div>
      )}
    </div>
  )
}

function DriveBar({ label, slot }) {
  const phaseLabel =
    slot.phase === 'write'
      ? 'Escrita'
      : slot.phase === 'read'
        ? 'Leitura'
        : slot.phase === 'done'
          ? 'Concluído'
          : slot.phase === 'error'
            ? 'Erro'
            : '—'

  return (
    <div className="drive-progress">
      <h3>
        {label} — {phaseLabel}
      </h3>
      <div className="bar-wrap">
        <div className="bar" style={{ width: `${Math.min(100, Math.max(0, slot.pct))}%` }} />
      </div>
      <div className="meta">
        <span>
          Progresso: <strong>{slot.pct.toFixed(1)}%</strong>
        </span>
        <span>
          Atual: <strong>{slot.speedMBps.toFixed(2)} MB/s</strong>
        </span>
      </div>
    </div>
  )
}

function ResultCard({ title, path, data }) {
  if (!data) return null
  return (
    <div className="card">
      <h4>
        {title}
        {path ? ` (${path})` : ''}
      </h4>
      {data.error ? (
        <p className="err">{data.error}</p>
      ) : (
        <>
          <p className="stat">
            <span>Escrita média:</span> {data.writeMBps?.toFixed(2) ?? '—'} MB/s
          </p>
          <p className="stat">
            <span>Leitura média:</span> {data.readMBps?.toFixed(2) ?? '—'} MB/s
          </p>
          <p className="stat">
            <span>Bytes escritos / lidos:</span> {formatBytes(data.writeBytes)} / {formatBytes(data.readBytes)}
          </p>
          <p className="stat">
            <span>Tempo total (drive):</span> {data.durationMs ?? '—'} ms
          </p>
        </>
      )}
    </div>
  )
}

function formatBytes(n) {
  if (n == null || Number.isNaN(n)) return '—'
  if (n < 1024) return `${n} B`
  if (n < 1024 * 1024) return `${(n / 1024).toFixed(1)} KB`
  return `${(n / (1024 * 1024)).toFixed(2)} MB`
}
