// Pacote principal do DualBench: benchmark paralelo de dois volumes com I/O sem cache
// (Windows: NO_BUFFERING + WRITE_THROUGH; Linux: O_DIRECT; macOS: F_NOCACHE).
package main

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"time"
	"unsafe"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

const (
	tempBenchFileName      = ".dualbench_uncached.dat"
	eventBenchmarkProgress = "benchmark:progress"
)

type BenchProgressPayload struct {
	Drive1     DriveSlot `json:"drive1"`
	Drive2     DriveSlot `json:"drive2"`
	OverallPct float64   `json:"overallPct"`
}

type DriveSlot struct {
	Phase   string  `json:"phase"`
	Pct     float64 `json:"pct"`
	SpeedMB float64 `json:"speedMBps"`
}

type DriveSummary struct {
	Path       string  `json:"path"`
	WriteMBps  float64 `json:"writeMBps"`
	ReadMBps   float64 `json:"readMBps"`
	WriteBytes int64   `json:"writeBytes"`
	ReadBytes  int64   `json:"readBytes"`
	DurationMs int64   `json:"durationMs"`
	Error      string  `json:"error,omitempty"`
}

type BenchmarkSummary struct {
	Drive1 DriveSummary `json:"drive1"`
	Drive2 DriveSummary `json:"drive2"`
}

// App expõe métodos ao Wails; ctx é preenchido em startup para EventsEmit.
type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) StartDualBenchmark(drivePathA, drivePathB string) (BenchmarkSummary, error) {
	var empty BenchmarkSummary
	if a.ctx == nil {
		return empty, errors.New("aplicativo ainda não inicializou o contexto")
	}
	return dualBenchmark(a.ctx, drivePathA, drivePathB, func(p BenchProgressPayload) {
		wailsruntime.EventsEmit(a.ctx, eventBenchmarkProgress, p)
	})
}

type progressAggregator struct {
	mu sync.Mutex
	d  [2]DriveSlot
}

func newProgressAggregator() *progressAggregator {
	return new(progressAggregator)
}

func (p *progressAggregator) store(slot int, phase string, pct float64, speedMB float64) {
	p.mu.Lock()
	p.d[slot] = DriveSlot{Phase: phase, Pct: pct, SpeedMB: speedMB}
	p.mu.Unlock()
}

func (p *progressAggregator) snapshot() BenchProgressPayload {
	p.mu.Lock()
	defer p.mu.Unlock()
	return BenchProgressPayload{
		Drive1:     p.d[0],
		Drive2:     p.d[1],
		OverallPct: (p.d[0].Pct + p.d[1].Pct) / 2,
	}
}

// dualBenchmark executa escrita e leitura sem cache nos dois volumes em paralelo.
// Se emit != nil, snapshots periódicos são enviados (~125 ms) até concluir.
func dualBenchmark(ctx context.Context, drivePathA, drivePathB string, emit func(BenchProgressPayload)) (BenchmarkSummary, error) {
	var empty BenchmarkSummary

	pathA := normalizeDriveRoot(drivePathA)
	pathB := normalizeDriveRoot(drivePathB)

	if err := ensureDirRoot(pathA); err != nil {
		return empty, fmt.Errorf("drive 1 (%s): %w", pathA, err)
	}
	if err := ensureDirRoot(pathB); err != nil {
		return empty, fmt.Errorf("drive 2 (%s): %w", pathB, err)
	}

	align := minDirectIOAlignment()
	chunk := defaultChunkSize(align)
	totalBytes := benchTotalBytes(align, chunk)

	agg := newProgressAggregator()
	ctxEmit, cancelEmit := context.WithCancel(ctx)
	defer cancelEmit()

	if emit != nil {
		go func() {
			t := time.NewTicker(125 * time.Millisecond)
			defer t.Stop()
			for {
				select {
				case <-ctxEmit.Done():
					return
				case <-t.C:
					emit(agg.snapshot())
				}
			}
		}()
	}

	var wg sync.WaitGroup
	var sum1, sum2 DriveSummary

	wg.Add(2)
	go func() {
		defer wg.Done()
		sum1 = benchOneDrive(ctx, pathA, 0, totalBytes, chunk, agg)
	}()
	go func() {
		defer wg.Done()
		sum2 = benchOneDrive(ctx, pathB, 1, totalBytes, chunk, agg)
	}()

	wg.Wait()
	cancelEmit()

	agg.store(0, "done", 100, 0)
	agg.store(1, "done", 100, 0)
	if emit != nil {
		emit(agg.snapshot())
	}

	return BenchmarkSummary{Drive1: sum1, Drive2: sum2}, nil
}

// benchOneDrive: por drive, sequência escrita → leitura no mesmo arquivo temporário.
func benchOneDrive(ctx context.Context, root string, slot int, total int64, chunk int, agg *progressAggregator) DriveSummary {
	sum := DriveSummary{Path: root}
	tmp := filepath.Join(root, tempBenchFileName)
	t0 := time.Now()

	defer func() {
		sum.DurationMs = time.Since(t0).Milliseconds()
		_ = os.Remove(tmp)
	}()

	agg.store(slot, "write", 0, 0)
	wf, err := openUncachedRW(tmp, false)
	if err != nil {
		sum.Error = fmt.Sprintf("abrir escrita: %v", err)
		agg.store(slot, "error", 0, 0)
		return sum
	}

	buf := alignedBuffer(chunk)
	var written int64
	wStart := time.Now()
	lastMark := wStart
	lastBytes := int64(0)

	for written < total {
		if ctx.Err() != nil {
			wf.Close()
			sum.Error = "cancelado"
			agg.store(slot, "error", float64(written)/float64(total)*50, 0)
			return sum
		}

		nr, err := wf.Write(buf)
		if err != nil {
			wf.Close()
			sum.Error = fmt.Sprintf("escrever: %v", err)
			agg.store(slot, "error", float64(written)/float64(total)*50, 0)
			return sum
		}
		written += int64(nr)

		pct := float64(written) / float64(total) * 50
		now := time.Now()
		if now.Sub(lastMark) >= 125*time.Millisecond {
			dt := now.Sub(lastMark).Seconds()
			var inst float64
			if dt > 0 {
				inst = float64(written-lastBytes) / (1024 * 1024) / dt
			}
			agg.store(slot, "write", pct, inst)
			lastMark = now
			lastBytes = written
		}
	}
	wf.Close()

	if written > 0 {
		sum.WriteBytes = written
		sum.WriteMBps = float64(written) / (1024 * 1024) / time.Since(wStart).Seconds()
	}

	agg.store(slot, "read", 50, 0)
	rf, err := openUncachedRW(tmp, true)
	if err != nil {
		sum.Error = fmt.Sprintf("abrir leitura: %v", err)
		agg.store(slot, "error", 50, 0)
		return sum
	}
	defer rf.Close()

	var read int64
	rStart := time.Now()
	lastMark = rStart
	lastBytes = 0

	for {
		if ctx.Err() != nil {
			sum.Error = "cancelado"
			agg.store(slot, "error", 50+float64(read)/float64(total)*50, 0)
			return sum
		}
		nr, err := rf.Read(buf)
		if nr > 0 {
			read += int64(nr)
			pct := 50 + float64(read)/float64(total)*50
			now := time.Now()
			if now.Sub(lastMark) >= 125*time.Millisecond {
				dt := now.Sub(lastMark).Seconds()
				var inst float64
				if dt > 0 {
					inst = float64(read-lastBytes) / (1024 * 1024) / dt
				}
				agg.store(slot, "read", pct, inst)
				lastMark = now
				lastBytes = read
			}
			continue
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			sum.Error = fmt.Sprintf("ler: %v", err)
			agg.store(slot, "error", 50+float64(read)/float64(total)*50, 0)
			return sum
		}
	}

	sum.ReadBytes = read
	if read > 0 {
		sum.ReadMBps = float64(read) / (1024 * 1024) / time.Since(rStart).Seconds()
	}
	agg.store(slot, "done", 100, 0)
	return sum
}

func defaultChunkSize(align int) int {
	const want = 4 * 1024 * 1024
	if want%align != 0 {
		return align * (want / align)
	}
	return want
}

func benchTotalBytes(align, chunk int) int64 {
	const want = int64(128) * 1024 * 1024
	n := want - (want % int64(align))
	if n <= 0 {
		n = int64(align) * 1024
	}
	n -= n % int64(chunk)
	if n <= 0 {
		n = int64(chunk)
	}
	return n
}

func normalizeDriveRoot(p string) string {
	p = strings.TrimSpace(p)
	p = filepath.Clean(p)
	if runtime.GOOS == "windows" {
		if len(p) == 2 && p[1] == ':' {
			p = p + `\`
		}
		return p
	}
	// Linux/macOS: montagens são absolutas; sem "/" o path vira relativo ao CWD e o stat falha.
	if p != "" && p != "." && !strings.HasPrefix(p, "/") && !strings.HasPrefix(p, ".") {
		p = "/" + p
	}
	return p
}

func ensureDirRoot(path string) error {
	fi, err := os.Stat(path)
	if err != nil {
		return err
	}
	if !fi.IsDir() {
		return fmt.Errorf("caminho não é diretório ou raiz de volume")
	}
	return nil
}

func alignedBuffer(size int) []byte {
	align := minDirectIOAlignment()
	if size%align != 0 {
		panic("chunk deve ser múltiplo do alinhamento de I/O direto")
	}
	raw := make([]byte, size+align)
	start := uintptr(unsafe.Pointer(&raw[0]))
	off := int((uintptr(align) - (start % uintptr(align))) % uintptr(align))
	return raw[off : off+size]
}
