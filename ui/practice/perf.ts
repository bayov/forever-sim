// The room's performance recorder, to find what makes it stutter.
//
// Every frame we time each part of the frame's work (showing the session's updates, moving,
// animating, drawing the scene, the HUD and the combat text) and note what the renderer drew.
// Things that often cause a hitch get noted with their time: a model or shader that loads mid
// fight, a long task, Chrome's long animation frames with the scripts that ran in them, and drops
// in the JS heap (a garbage collection). We also time the GPU's work on drawing the scene, when
// the browser lets us (EXT_disjoint_timer_query_webgl2). When the fight ends, the report goes to the practice
// server, which keeps it in ~/forever-practice/perf/.

import type * as THREE from 'three';

// The parts of a frame we time, in the order they run.
export const parts = ['inbox', 'move', 'units', 'render', 'hud', 'sct'] as const;
export type Part = (typeof parts)[number];

// One frame: when it started, the time since the frame before, the time of each part, and the
// renderer's numbers.
type Frame = {
	t: number;
	gap: number;
	cpu: number;
	parts: number[];
	calls: number;
	tris: number;
	// The GPU's time for drawing the scene. It comes in a few frames later, so it stays -1 when
	// the result never came (or the browser can't time the GPU).
	gpu: number;
	programs: number;
	geometries: number;
	textures: number;
	heapMb?: number;
};

type Note = { t: number; kind: string; ms?: number; detail?: string };

// Ten minutes at 165 FPS.
const MAX_FRAMES = 600 * 165;

type MemoryInfo = { usedJSHeapSize: number };

export class Perf {
	private frames: Frame[] = [];
	private notes: Note[] = [];
	private current?: { t: number; start: number; parts: number[]; mark: number };
	private lastT = 0;
	private lastHeap = 0;
	private lastPrograms = 0;
	private startedAt = performance.now();
	private observers: PerformanceObserver[] = [];
	private gl?: WebGL2RenderingContext;
	private timer?: { TIME_ELAPSED_EXT: number; GPU_DISJOINT_EXT: number } | null;
	private queries: { query: WebGLQuery; frame: Frame }[] = [];
	private freeQueries: WebGLQuery[] = [];
	private open?: WebGLQuery;

	constructor(readonly renderer: THREE.WebGLRenderer) {
		this.observe('longtask', e =>
			this.note(
				'longtask',
				e.startTime,
				e.duration,
				(e as PerformanceEntry & { attribution?: { containerType?: string; name?: string }[] }).attribution
					?.map(a => a.name ?? a.containerType)
					.join(','),
			),
		);
		this.observe('long-animation-frame', e => {
			const loaf = e as PerformanceEntry & {
				blockingDuration?: number;
				renderStart?: number;
				styleAndLayoutStart?: number;
				scripts?: { invoker?: string; sourceFunctionName?: string; sourceURL?: string; duration: number; forcedStyleAndLayoutDuration?: number }[];
			};
			const scripts = (loaf.scripts ?? [])
				.filter(s => s.duration >= 2)
				.map(
					s =>
						`${s.invoker ?? ''} ${s.sourceFunctionName ?? ''} ${(s.sourceURL ?? '').split('/').pop()} ${s.duration.toFixed(0)}ms${
							s.forcedStyleAndLayoutDuration ? ` layout ${s.forcedStyleAndLayoutDuration.toFixed(0)}ms` : ''
						}`,
				)
				.join(' | ');
			const style = loaf.styleAndLayoutStart && loaf.renderStart ? ` render ${(loaf.startTime + loaf.duration - loaf.renderStart).toFixed(0)}ms` : '';
			this.note('loaf', e.startTime, e.duration, `blocking ${(loaf.blockingDuration ?? 0).toFixed(0)}ms${style}${scripts ? ' · ' + scripts : ''}`);
		});
	}

	private observe(type: string, fn: (e: PerformanceEntry) => void) {
		try {
			const o = new PerformanceObserver(list => list.getEntries().forEach(fn));
			o.observe({ type, buffered: false });
			this.observers.push(o);
		} catch {
			// This browser doesn't report that kind of entry.
		}
	}

	// gpuBegin and gpuEnd go around drawing the scene, to time the GPU's work on it.
	gpuBegin() {
		if (this.timer === undefined) {
			const gl = this.renderer.getContext();
			this.gl = gl instanceof WebGL2RenderingContext ? gl : undefined;
			this.timer = this.gl?.getExtension('EXT_disjoint_timer_query_webgl2') ?? null;
		}
		if (!this.timer || !this.gl || !this.current) return;
		const query = this.freeQueries.pop() ?? this.gl.createQuery();
		if (!query) return;
		this.gl.beginQuery(this.timer.TIME_ELAPSED_EXT, query);
		this.open = query;
	}

	gpuEnd() {
		if (!this.open || !this.gl || !this.timer) return;
		this.gl.endQuery(this.timer.TIME_ELAPSED_EXT);
		this.pendingQuery = this.open;
		this.open = undefined;
	}

	private pendingQuery?: WebGLQuery;

	// readGpu picks up the GPU times that came in. A disjoint (the GPU did something else in
	// between, like another tab) makes them all meaningless, so we drop them.
	private readGpu() {
		const gl = this.gl;
		const timer = this.timer;
		if (!gl || !timer || !this.queries.length) return;
		const disjoint = gl.getParameter(timer.GPU_DISJOINT_EXT);
		while (this.queries.length) {
			const q = this.queries[0];
			if (!disjoint && !gl.getQueryParameter(q.query, gl.QUERY_RESULT_AVAILABLE)) break;
			if (!disjoint) q.frame.gpu = round(gl.getQueryParameter(q.query, gl.QUERY_RESULT) / 1e6);
			this.queries.shift();
			this.freeQueries.push(q.query);
		}
	}

	// note records something that happened, at a performance.now() time.
	note(kind: string, t = performance.now(), ms?: number, detail?: string) {
		if (this.notes.length < 20000) this.notes.push({ t: round(t), kind, ms: ms === undefined ? undefined : round(ms), detail });
	}

	// begin starts a frame's timing, with requestAnimationFrame's time.
	begin(t: number) {
		const now = performance.now();
		this.current = { t, start: now, parts: parts.map(() => 0), mark: now };
	}

	// part ends the timing of one part of the frame (the time since the part before).
	part(p: Part) {
		const c = this.current;
		if (!c) return;
		const now = performance.now();
		c.parts[parts.indexOf(p)] += now - c.mark;
		c.mark = now;
	}

	end() {
		const c = this.current;
		if (!c) return;
		this.current = undefined;
		const info = this.renderer.info;
		const heap = (performance as Performance & { memory?: MemoryInfo }).memory?.usedJSHeapSize;
		const frame: Frame = {
			t: round(c.t),
			gap: this.lastT ? round(c.t - this.lastT) : 0,
			cpu: round(performance.now() - c.start),
			parts: c.parts.map(round),
			calls: info.render.calls,
			tris: info.render.triangles,
			gpu: -1,
			programs: info.programs?.length ?? 0,
			geometries: info.memory.geometries,
			textures: info.memory.textures,
			heapMb: heap === undefined ? undefined : round(heap / 1048576),
		};
		this.lastT = c.t;
		if (this.pendingQuery) {
			this.queries.push({ query: this.pendingQuery, frame });
			this.pendingQuery = undefined;
		}
		this.readGpu();
		// A heap that shrank by more than 2MB was a garbage collection.
		if (heap !== undefined) {
			if (this.lastHeap - heap > 2 * 1048576) this.note('gc', c.t, undefined, `${round((this.lastHeap - heap) / 1048576)}MB freed`);
			this.lastHeap = heap;
		}
		if (frame.programs > this.lastPrograms && this.lastPrograms > 0) this.note('shader', c.t, undefined, `${frame.programs - this.lastPrograms} new`);
		this.lastPrograms = frame.programs;
		this.frames.push(frame);
		if (this.frames.length > MAX_FRAMES) this.frames.splice(0, this.frames.length - MAX_FRAMES);
	}

	// report sums up the frames and lists the slowest ones with what happened around them.
	report(extra: Record<string, unknown>) {
		const frames = this.frames;
		const gaps = frames.map(f => f.gap).filter(g => g > 0);
		const sorted = [...gaps].sort((a, b) => a - b);
		const pct = (q: number) => (sorted.length ? sorted[Math.min(sorted.length - 1, Math.floor(q * sorted.length))] : 0);
		const median = pct(0.5);
		const partSums = parts.map((_, i) => frames.reduce((a, f) => a + f.parts[i], 0) / Math.max(1, frames.length));
		const timed = frames.filter(f => f.gpu >= 0);
		const gpu = timed.length ? round(timed.reduce((a, f) => a + f.gpu, 0) / timed.length) : undefined;
		// A stutter is a frame that came more than twice the usual frame time after the one before.
		const slow = frames
			.filter(f => f.gap > Math.max(2 * median, median + 8))
			.sort((a, b) => b.gap - a.gap)
			.slice(0, 60)
			.map(f => ({
				...f,
				parts: Object.fromEntries(parts.map((p, i) => [p, f.parts[i]])),
				notes: this.notes.filter(n => n.t >= f.t - f.gap - 5 && n.t <= f.t + 5),
			}));
		return {
			version: 1,
			at: new Date().toISOString(),
			userAgent: navigator.userAgent,
			screen: { width: innerWidth, height: innerHeight, pixelRatio: devicePixelRatio },
			seconds: round((performance.now() - this.startedAt) / 1000),
			frameCount: frames.length,
			frameMs: { p50: median, p90: pct(0.9), p99: pct(0.99), max: sorted[sorted.length - 1] ?? 0 },
			over: {
				'20ms': gaps.filter(g => g > 20).length,
				'33ms': gaps.filter(g => g > 33).length,
				'50ms': gaps.filter(g => g > 50).length,
				'100ms': gaps.filter(g => g > 100).length,
			},
			partsAvgMs: { ...Object.fromEntries(parts.map((p, i) => [p, round(partSums[i])])), gpu },
			slowest: slow,
			notes: this.notes,
			// Every frame, as [time, gap, cpu, ...parts, calls, triangles, heap MB, GPU ms].
			frames: frames.map(f => [f.t, f.gap, f.cpu, ...f.parts, f.calls, f.tris, f.heapMb ?? 0, f.gpu]),
			frameColumns: ['t', 'gap', 'cpu', ...parts, 'calls', 'tris', 'heapMb', 'gpu'],
			...extra,
		};
	}

	// reset starts a new recording, for the next fight.
	reset() {
		this.frames = [];
		this.notes = [];
		this.lastT = 0;
		this.startedAt = performance.now();
	}
}

function round(x: number): number {
	return Math.round(x * 100) / 100;
}
