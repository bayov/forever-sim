// The practice worker: runs practice.wasm with its own clock (docs/practice_room.md, "Running in
// real time"). It steps the sim on each input, and sets a timer to the sim's next event, so it
// never keeps a CPU core busy.

import { ClockState, FromWorker, simTimeAt, ToWorker, wallNow } from './protocol';

declare const self: DedicatedWorkerGlobalScope;

type GoGlobal = {
	practiceStart(args: string): string;
	practiceInput(input: string, pressedAtMs: number, nowMs: number): string;
	practiceAdvance(nowMs: number): string;
	practiceStartEncounter(nowMs: number): string;
	practiceStop(nowMs: number): string;
	practiceAttempt(): string;
};
const go = self as unknown as GoGlobal;

let clock: ClockState = { wallAtStart: 0, startMs: 0, shiftMs: 0, pausedMs: 0 };
let timer: ReturnType<typeof setTimeout> | undefined;
let heartbeat: ReturnType<typeof setInterval> | undefined;
let running = false;

function post(msg: FromWorker) {
	self.postMessage(msg);
}

function simNow(wall = wallNow()): number {
	return simTimeAt(clock, wall);
}

async function loadWasm() {
	const execText = await (await fetch('/wasm_exec.js')).text();
	// wasm_exec.js is a classic script that puts Go on the global object.
	new Function(execText)();
	const GoClass = (self as unknown as { Go: new () => { importObject: WebAssembly.Imports; run(i: WebAssembly.Instance): Promise<void> } }).Go;
	const goRuntime = new GoClass();
	const ready = new Promise<void>(resolve => {
		(self as unknown as { wasmready: () => void }).wasmready = resolve;
	});
	const result = await WebAssembly.instantiateStreaming(fetch('/practice.wasm'), goRuntime.importObject);
	goRuntime.run(result.instance);
	await ready;
}

// handle reads a reply from Go, sends its batch, and sets the timer to the next event.
function handle(replyText: string) {
	const reply = JSON.parse(replyText);
	if (reply.error) {
		post({ type: 'error', message: reply.error });
		return;
	}
	const batch = reply.batch;
	if (batch?.rebase) {
		clock = { ...clock, shiftMs: clock.shiftMs + batch.rebase.shiftMs };
	}
	if (batch) {
		post({ type: 'batch', batch, simNowMs: simNow(), clock });
		if (batch.ended) {
			running = false;
		}
	}
	schedule(reply.nextEventMs);
}

function schedule(nextEventMs: number | undefined) {
	if (timer !== undefined) {
		clearTimeout(timer);
		timer = undefined;
	}
	if (!running || clock.pausedAt !== undefined || nextEventMs === undefined || nextEventMs < 0) {
		return;
	}
	const delay = Math.max(0, nextEventMs - simNow());
	timer = setTimeout(tick, Math.min(delay, 1000));
}

function tick() {
	timer = undefined;
	if (!running || clock.pausedAt !== undefined) return;
	handle(go.practiceAdvance(simNow()));
}

async function onMessage(msg: ToWorker) {
	switch (msg.type) {
		case 'start': {
			const reply = JSON.parse(go.practiceStart(JSON.stringify({ request: msg.request, settings: msg.settings, seed: msg.seed })));
			if (reply.error) {
				post({ type: 'error', message: reply.error });
				return;
			}
			clock = { wallAtStart: msg.wallNow, startMs: reply.startMs, shiftMs: 0, pausedMs: 0 };
			running = true;
			post({ type: 'started', toolkit: reply.toolkit, startMs: reply.startMs, batch: reply.batch, wallAtStart: msg.wallNow });
			post({ type: 'batch', batch: reply.batch, simNowMs: simNow(), clock });
			if (heartbeat === undefined) {
				// The HUD's state between events, a few times a second.
				heartbeat = setInterval(() => running && clock.pausedAt === undefined && tick(), 250);
			}
			tick();
			break;
		}
		case 'input': {
			if (!running) return;
			const pressedAt = simNow(msg.pressedAtWall);
			handle(go.practiceInput(JSON.stringify(msg.input), pressedAt, simNow()));
			break;
		}
		case 'startEncounter':
			if (!running) return;
			handle(go.practiceStartEncounter(simNow(msg.wallNow)));
			break;
		case 'stop':
			if (!running) return;
			handle(go.practiceStop(simNow(msg.wallNow)));
			running = false;
			break;
		case 'pause':
			if (clock.pausedAt === undefined) {
				clock = { ...clock, pausedAt: msg.wallNow };
				schedule(undefined);
			}
			break;
		case 'resume':
			if (clock.pausedAt !== undefined) {
				clock = { ...clock, pausedMs: clock.pausedMs + (msg.wallNow - clock.pausedAt), pausedAt: undefined };
				tick();
			}
			break;
		case 'attempt': {
			const attempt = JSON.parse(go.practiceAttempt());
			post({ type: 'attempt', attempt });
			break;
		}
	}
}

const queue: ToWorker[] = [];
let loaded = false;
self.onmessage = (e: MessageEvent<ToWorker>) => {
	if (!loaded) {
		queue.push(e.data);
		return;
	}
	onMessage(e.data);
};

loadWasm()
	.then(() => {
		loaded = true;
		post({ type: 'ready' });
		for (const msg of queue.splice(0)) onMessage(msg);
	})
	.catch(err => post({ type: 'error', message: String(err) }));
