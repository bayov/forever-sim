// The room's performance recorder, to find what makes it stutter.
//
// Every frame we time each part of the frame's work (showing the session's updates, moving,
// animating, drawing the scene, the HUD and the combat text) and note what the renderer drew.
// Things that often cause a hitch get noted with their time: a model or shader that loads mid
// fight, a long task, Chrome's long animation frames with the scripts that ran in them, and drops
// in the JS heap (a garbage collection). When the fight ends, the report goes to the practice
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
			programs: info.programs?.length ?? 0,
			geometries: info.memory.geometries,
			textures: info.memory.textures,
			heapMb: heap === undefined ? undefined : round(heap / 1048576),
		};
		this.lastT = c.t;
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
			partsAvgMs: Object.fromEntries(parts.map((p, i) => [p, round(partSums[i])])),
			slowest: slow,
			notes: this.notes,
			// Every frame, as [time, gap, cpu, ...parts, calls, triangles, heap MB].
			frames: frames.map(f => [f.t, f.gap, f.cpu, ...f.parts, f.calls, f.tris, f.heapMb ?? 0]),
			frameColumns: ['t', 'gap', 'cpu', ...parts, 'calls', 'tris', 'heapMb'],
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
