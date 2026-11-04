// The room's performance recorder, to find what makes it stutter.
//
// Every frame we time each part of the frame's work (showing the session's updates, moving,
// animating, drawing the scene, the HUD and the combat text) and note what the renderer drew.
// Things that often cause a hitch get noted with their time: a model or shader that loads mid
// fight, a long task, Chrome's long animation frames with the scripts that ran in them, and drops
// in the JS heap (a garbage collection). We also time the GPU's work on drawing the scene by pass
// (the uploads, the shadows, the solid things, and the see-through things with the effects), when the browser
// lets us (EXT_disjoint_timer_query_webgl2). When the fight ends, the report goes to the practice
// server, which keeps it in ~/forever-practice/perf/.

import type * as THREE from 'three';

// The parts of a frame we time, in the order they run.
export const parts = ['inbox', 'move', 'units', 'render', 'hud', 'sct'] as const;
export type Part = (typeof parts)[number];

// The passes of drawing the scene that we time on the GPU, in the order they run.
//
// 'upload' is what comes before the shadows: the renderer sends the frame's changed vertices and
// textures to the GPU.
//
// The see-through things come in kinds (drawOrder, wow/particles.ts): 'transparent' is the units'
// own see-through parts, then the raid test's boss, its raiders' gear and bodies, the spell
// effects, the particles and the ribbons.
export const passes = ['upload', 'shadow', 'opaque', 'transparent', 'boss', 'gear', 'effects', 'particles', 'ribbons'] as const;
export type Pass = (typeof passes)[number];

// One frame: when it started, the time since the frame before, the time of each part, and the
// renderer's numbers.
type Frame = {
	t: number;
	gap: number;
	cpu: number;
	parts: number[];
	calls: number;
	// How many particles the frame drew.
	particles: number;
	// The frame's uploads to the GPU: how many calls, and how many kilobytes.
	uploads: number;
	uploadKb: number;
	tris: number;
	// The GPU's time for drawing the scene, in all and by pass. It comes in a few frames later, so
	// it stays -1 when a result never came (or the browser can't time the GPU).
	gpu: number;
	gpuPasses: number[];
	// How many of the frame's passes are still out.
	gpuOut: number;
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
	private queries: { query: WebGLQuery; frame: Frame; pass: number; at: number }[] = [];
	private frameNo = 0;
	private freeQueries: WebGLQuery[] = [];
	private open?: { query: WebGLQuery; pass: number };
	// The passes timed in the frame so far, which go out with the frame at its end.
	private framePasses: { query: WebGLQuery; pass: number }[] = [];

	// particles counts the particles drawn in the frame (set by the room).
	particles?: () => number;
	private uploads = 0;
	private uploadBytes = 0;
	// The WebGL calls by kind since the recording started (with ?glcount=1), and over how many frames.
	private glCounters = new Map<string, { n: number }>();

	constructor(readonly renderer: THREE.WebGLRenderer) {
		// We count the uploads to the GPU (buffers and textures), each with its size.
		const gl = renderer.getContext() as WebGL2RenderingContext;
		// With ?glcount=1 we count every WebGL call by kind (it costs a little time each), since
		// Chrome's GPU process takes time for each command we send, not only each draw.
		if (new URLSearchParams(location.search).get('glcount') === '1') {
			const own = gl as unknown as Record<string, unknown>;
			for (const name of Object.getOwnPropertyNames(WebGL2RenderingContext.prototype)) {
				const d = Object.getOwnPropertyDescriptor(WebGL2RenderingContext.prototype, name);
				if (!d || typeof d.value !== 'function' || name === 'constructor') continue;
				const f = (d.value as (...a: unknown[]) => unknown).bind(gl);
				const before = own[name] !== d.value ? (own[name] as (...a: unknown[]) => unknown) : f;
				// A counter per kind, and no array of the arguments (spread, each call made garbage:
				// 2300 calls a frame).
				const count = { n: 0 };
				this.glCounters.set(name, count);
				own[name] = function (this: unknown) {
					count.n++;
					// eslint-disable-next-line prefer-rest-params
					return before.apply(gl, arguments as unknown as unknown[]);
				};
			}
		}
		for (const name of ['bufferSubData', 'bufferData', 'texSubImage2D', 'texImage2D', 'texSubImage3D', 'texImage3D'] as const) {
			const f = (gl[name] as (...a: unknown[]) => void).bind(gl);
			const perf = this;
			(gl as unknown as Record<string, unknown>)[name] = function (this: unknown) {
				perf.uploads++;
				for (let i = 0; i < arguments.length; i++) {
					// eslint-disable-next-line prefer-rest-params
					const x = arguments[i];
					if (ArrayBuffer.isView(x)) perf.uploadBytes += x.byteLength;
				}
				// eslint-disable-next-line prefer-rest-params
				return f.apply(gl, arguments as unknown as unknown[]);
			};
		}
		// When the GPU drops the page's drawing (a lost context), the queries out are gone, and the
		// timer is found again once it's back.
		const canvas = renderer.domElement;
		canvas.addEventListener('webglcontextlost', () => {
			this.note('contextlost');
			for (const q of this.queries) q.frame.gpuOut = -1;
			this.queries = [];
			this.freeQueries = [];
			this.framePasses = [];
			this.open = undefined;
			this.timer = null;
		});
		canvas.addEventListener('webglcontextrestored', () => {
			this.note('contextrestored');
			this.timer = undefined;
		});
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

	// gpuPass starts timing a pass of drawing the scene on the GPU, and ends the one before. With no
	// pass it ends the last one, once the scene is drawn.
	gpuPass(pass: Pass | undefined) {
		if (this.timer === undefined) {
			const gl = this.renderer.getContext();
			this.gl = gl instanceof WebGL2RenderingContext ? gl : undefined;
			this.timer = this.gl?.getExtension('EXT_disjoint_timer_query_webgl2') ?? null;
		}
		const gl = this.gl;
		if (!this.timer || !gl) return;
		if (this.open) {
			gl.endQuery(this.timer.TIME_ELAPSED_EXT);
			this.framePasses.push(this.open);
			this.open = undefined;
		}
		if (!pass || !this.current) return;
		const query = this.freeQueries.pop() ?? gl.createQuery();
		if (!query) return;
		gl.beginQuery(this.timer.TIME_ELAPSED_EXT, query);
		this.open = { query, pass: passes.indexOf(pass) };
	}

	// readGpu picks up the GPU times that came in. A disjoint (the GPU did something else in
	// between, like another tab) makes them all meaningless, so we drop them.
	private readGpu() {
		const gl = this.gl;
		const timer = this.timer;
		if (!gl || !timer || !this.queries.length) return;
		const disjoint = gl.getParameter(timer.GPU_DISJOINT_EXT);
		while (this.queries.length) {
			const q = this.queries[0];
			if (!disjoint && !gl.getQueryParameter(q.query, gl.QUERY_RESULT_AVAILABLE)) {
				// A result comes a few frames later. One that never comes (it happened in a headless
				// run) would hold up all the others, so after a second we give up on it.
				if (this.frameNo - q.at < 240) break;
				q.frame.gpuOut = -1;
				this.queries.shift();
				gl.deleteQuery(q.query);
				continue;
			}
			const f = q.frame;
			if (disjoint) f.gpuOut = -1;
			else if (f.gpuOut > 0) {
				f.gpuPasses[q.pass] += gl.getQueryParameter(q.query, gl.QUERY_RESULT) / 1e6;
				if (--f.gpuOut === 0) {
					f.gpuPasses = f.gpuPasses.map(round);
					f.gpu = round(f.gpuPasses.reduce((a, b) => a + b, 0));
				}
			}
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
			uploads: this.uploads,
			particles: this.particles?.() ?? 0,
			uploadKb: round(this.uploadBytes / 1024),
			tris: info.render.triangles,
			gpu: -1,
			gpuPasses: passes.map(() => 0),
			gpuOut: 0,
			programs: info.programs?.length ?? 0,
			geometries: info.memory.geometries,
			textures: info.memory.textures,
			heapMb: heap === undefined ? undefined : round(heap / 1048576),
		};
		this.lastT = c.t;
		this.uploads = 0;
		this.uploadBytes = 0;
		frame.gpuOut = this.framePasses.length;
		this.frameNo++;
		for (const p of this.framePasses) this.queries.push({ ...p, frame, at: this.frameNo });
		this.framePasses = [];
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
		const gpuPasses = timed.length ? Object.fromEntries(passes.map((p, i) => [p, round(timed.reduce((a, f) => a + f.gpuPasses[i], 0) / timed.length)])) : undefined;
		const gl = this.renderer.getContext();
		const info = gl.getExtension('WEBGL_debug_renderer_info');
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
			// The GPU's time by pass, and the uploads to it a frame, on average.
			gpuPassesAvgMs: gpuPasses,
			// The WebGL calls a frame by kind, with ?glcount=1.
			glCallsAvg: this.glCounters.size
				? Object.fromEntries(
						[...this.glCounters]
							.filter(([, c]) => c.n > 0)
							.sort((a, b) => b[1].n - a[1].n)
							.slice(0, 25)
							.map(([k, c]) => [k, round(c.n / Math.max(1, frames.length))]),
					)
				: undefined,
			uploadsAvg: { calls: round(frames.reduce((a, f) => a + f.uploads, 0) / Math.max(1, frames.length)), kb: round(frames.reduce((a, f) => a + f.uploadKb, 0) / Math.max(1, frames.length)) },
			gpuInfo: {
				renderer: info ? gl.getParameter(info.UNMASKED_RENDERER_WEBGL) : gl.getParameter(gl.RENDERER),
				vendor: info ? gl.getParameter(info.UNMASKED_VENDOR_WEBGL) : gl.getParameter(gl.VENDOR),
				canvas: [gl.drawingBufferWidth, gl.drawingBufferHeight],
				antialias: gl.getContextAttributes()?.antialias,
				samples: gl.getParameter(gl.SAMPLES),
				software: softwareRenderer(String(info ? gl.getParameter(info.UNMASKED_RENDERER_WEBGL) : gl.getParameter(gl.RENDERER))),
			},
			slowest: slow,
			notes: this.notes,
			// Every frame, as [time, gap, cpu, ...parts, calls, triangles, heap MB, GPU ms].
			frames: frames.map(f => [f.t, f.gap, f.cpu, ...f.parts, f.calls, f.tris, f.heapMb ?? 0, f.uploads, f.uploadKb, f.particles, f.gpu, ...(f.gpu >= 0 ? f.gpuPasses : passes.map(() => -1))]),
			frameColumns: ['t', 'gap', 'cpu', ...parts, 'calls', 'tris', 'heapMb', 'uploads', 'uploadKb', 'particles', 'gpu', ...passes.map(p => `gpu_${p}`)],
			...extra,
		};
	}

	// reset starts a new recording, for the next fight.
	reset() {
		this.frames = [];
		this.notes = [];
		this.lastT = 0;
		for (const c of this.glCounters.values()) c.n = 0;
		this.startedAt = performance.now();
	}
}

// softwareRenderer says a WebGL renderer's name is one that draws without the graphics card.
export function softwareRenderer(name: string): boolean {
	return /Basic Render Driver|SwiftShader|llvmpipe|softpipe|Software/i.test(name);
}

function round(x: number): number {
	return Math.round(x * 100) / 100;
}
