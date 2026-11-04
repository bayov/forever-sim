// The page's side of the pose worker (poseWorker.ts), which poses the raid test's crowd off the
// page's thread.
//
// A model posed there asks for its pose each update, like it would pose itself (model.ts), and
// gets the pose a frame later: once a frame (flush) we send the worker every model's latest
// request, and when it's done each model reads the new set of its bone matrices from shared
// memory. Your own character always poses itself, with no lag.
//
// Shared memory needs a cross-origin isolated page (the practice server sends the headers).
// Without it the crowd poses itself, as before.

import type { Model } from './model';
import { PoseRequest, RECORD, ToWorker } from './pose';

type Handle = {
	id: number;
	model: Model;
	// The two sets of bone matrices, and the one the page reads.
	placeSets: Float64Array[];
	skinSets: Float32Array[];
	read: number;
	// The latest request, waiting for the next flush.
	record: Float64Array;
	waiting: boolean;
};

class PoseClient {
	readonly enabled = typeof SharedArrayBuffer !== 'undefined' && typeof crossOriginIsolated !== 'undefined' && crossOriginIsolated;
	private worker?: Worker;
	private nextId = 1;
	private handles = new Map<Model, Handle>();
	private sentFiles = new Set<number>();
	private sentAnims = new Set<string>();
	private frame = 0;
	// The frame the worker is posing, and the models in it with the set each is written to.
	private busy?: { frame: number; posed: [Handle, number][] };
	private data = new Float64Array(RECORD * 64);

	private post(msg: ToWorker, transfer: Transferable[] = []) {
		if (!this.worker) {
			this.worker = new Worker(new URL('./poseWorker.ts', import.meta.url), { type: 'module' });
			this.worker.onmessage = (e: MessageEvent<{ type: 'done'; frame: number }>) => this.done(e.data.frame);
		}
		this.worker.postMessage(msg, transfer);
	}

	// attach has the worker pose a model from now on. It starts from the pose the model has.
	attach(model: Model) {
		if (!this.enabled || this.handles.has(model)) return;
		const { fdid, file, anims } = model.files;
		if (!fdid) return;
		if (!this.sentFiles.has(fdid)) {
			this.sentFiles.add(fdid);
			this.post({ type: 'file', fdid, bytes: file.slice(0) });
		}
		for (const seq of anims.keys()) this.sendAnim(model, seq);
		const own = model.poser;
		const places = [0, 1].map(() => new SharedArrayBuffer(own.places.byteLength));
		const skinning = [0, 1].map(() => new SharedArrayBuffer(own.skinning.byteLength));
		const h: Handle = {
			id: this.nextId++,
			model,
			placeSets: places.map(b => new Float64Array(b)),
			skinSets: skinning.map(b => new Float32Array(b)),
			read: 0,
			record: new Float64Array(RECORD),
			waiting: false,
		};
		for (const p of h.placeSets) p.set(own.places);
		for (const p of h.skinSets) p.set(own.skinning);
		this.handles.set(model, h);
		this.post({ type: 'model', id: h.id, fdid, places, skinning });
		model.remote = {
			places: h.placeSets[0],
			skinning: h.skinSets[0],
			request: req => this.request(h, req),
		};
		model.poseVersion++;
	}

	// detach has the model pose itself again, from the worker's last pose.
	detach(model: Model) {
		const h = this.handles.get(model);
		if (!h) return;
		this.handles.delete(model);
		model.poser.places.set(h.placeSets[h.read]);
		model.poser.skinning.set(h.skinSets[h.read]);
		model.remote = undefined;
		model.poseVersion++;
		this.post({ type: 'drop', id: h.id });
	}

	private sendAnim(model: Model, seq: number) {
		const { fdid, anims } = model.files;
		const key = `${fdid}:${seq}`;
		const bytes = anims.get(seq);
		if (!bytes || this.sentAnims.has(key)) return;
		this.sentAnims.add(key);
		this.post({ type: 'anim', fdid, seq, bytes: bytes.slice(0) });
	}

	// request keeps a model's latest request for the next flush. A snapshot (the start of a fade)
	// stays asked for until it's sent.
	private request(h: Handle, req: PoseRequest) {
		const r = h.record;
		const m2 = h.model.poser.m2;
		const needs = (seq: number, source?: DataView) => {
			if (source && !m2.inline(seq)) this.sendAnim(h.model, seq);
			return source ? 1 : 0;
		};
		r[0] = h.id;
		r[2] = req.globalTime;
		r[3] = req.base.seq;
		r[4] = req.base.time;
		r[5] = needs(req.base.seq, req.base.source);
		r[6] = req.fade;
		if (req.snapshot) r[7] = 1;
		const o = req.overlay;
		r[8] = o ? o.seq : -1;
		r[9] = o ? o.time : 0;
		r[10] = o ? needs(o.seq, o.source) : 0;
		r[11] = req.overlayWeight;
		const hands = req.hands;
		r[12] = hands ? hands.seq : -1;
		r[13] = hands ? needs(hands.seq, hands.source) : 0;
		r[14] = hands?.right ? 1 : 0;
		r[15] = hands?.left ? 1 : 0;
		r[16] = req.twist;
		r[17] = req.camera && req.world ? 1 : 0;
		if (req.camera && req.world) {
			req.camera.toArray(r, 18);
			req.world.toArray(r, 22);
		}
		h.waiting = true;
	}

	// flush sends the worker every model's latest request, once a frame, unless it's still posing
	// the last frame's (then they wait for the next).
	flush() {
		if (this.busy) return;
		const posed: [Handle, number][] = [];
		for (const h of this.handles.values()) if (h.waiting) posed.push([h, 1 - h.read]);
		if (posed.length === 0) return;
		if (this.data.length < posed.length * RECORD) this.data = new Float64Array(posed.length * RECORD * 2);
		const d = this.data;
		posed.forEach(([h, target], i) => {
			h.record[1] = target;
			d.set(h.record, i * RECORD);
			h.record[7] = 0;
			h.waiting = false;
		});
		this.frame++;
		this.busy = { frame: this.frame, posed };
		this.post({ type: 'frame', frame: this.frame, data: d.slice(0, posed.length * RECORD) });
	}

	// done swaps in the new poses, and moves what hangs off the bones to them.
	private done(frame: number) {
		const busy = this.busy;
		if (!busy || busy.frame !== frame) return;
		this.busy = undefined;
		for (const [h, target] of busy.posed) {
			if (this.handles.get(h.model) !== h || !h.model.remote) continue;
			h.read = target;
			h.model.remote.places = h.placeSets[target];
			h.model.remote.skinning = h.skinSets[target];
			h.model.poseVersion++;
			h.model.placeSockets();
		}
	}
}

export const poseClient = new PoseClient();
